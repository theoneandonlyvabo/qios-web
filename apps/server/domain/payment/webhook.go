// domain/payment/webhook.go
//
// Webhook handler untuk notifikasi dari Xendit.
// Endpoint ini TIDAK pakai Bearer auth — diverifikasi via Xendit callback token.
//
// Flow verifikasi (requirement PG-05):
//   1. Baca header x-callback-token
//   2. Bandingkan dengan XENDIT_WEBHOOK_TOKEN dari config
//   3. Tolak kalau tidak cocok (401)
//   4. Parse payload, update xendit_payments + pos_orders dalam satu tx
//
// Endpoint: POST /webhooks/xendit

package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	applogger "github.com/theoneandonlyvabo/qios-web/apps/server/platform/logger"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
)

// WebhookHandler menangani webhook dari Xendit.
// Dipisah dari Handler karena tidak butuh Service — langsung ke Repository
// dan harus bisa dipanggil tanpa auth middleware.
type WebhookHandler struct {
	db           *sql.DB
	repo         Repository
	webhookToken string // XENDIT_WEBHOOK_TOKEN dari config
}

func NewWebhookHandler(db *sql.DB, repo Repository, webhookToken string) *WebhookHandler {
	return &WebhookHandler{db: db, repo: repo, webhookToken: webhookToken}
}

// ----------------------------------------------------------------
// Payload shapes
// ----------------------------------------------------------------

// qrWebhookPayload adalah subset dari Xendit QR Code webhook body.
// Field lain disimpan di raw_payload JSONB — tidak perlu di-parse semua.
type qrWebhookPayload struct {
	Event   string `json:"event"`
	ID      string `json:"id"` // qrpy_xxx (payment id)
	Amount  int64  `json:"amount"`
	Status  string `json:"status"` // "COMPLETED" untuk sukses
	Created string `json:"created"`
	QRCode  struct {
		ID         string `json:"id"`          // qr_xxx
		ExternalID string `json:"external_id"` // pos_orders.order_id
		QRString   string `json:"qr_string"`
		Type       string `json:"type"`
	} `json:"qr_code"`
}

// accountActivatedPayload adalah subset webhook account.activated dari Xendit.
// Dikirim ketika merchant menyelesaikan KYC dan sub-account resmi aktif.
type accountActivatedPayload struct {
	Event string `json:"event"`
	Data  struct {
		ID     string `json:"id"`     // Xendit sub-account id (xendit_account_id di businesses)
		Status string `json:"status"` // "ACTIVE"
	} `json:"data"`
}

// ----------------------------------------------------------------
// Handler
// ----------------------------------------------------------------

// HandleWebhook menangani semua webhook Xendit di satu endpoint.
// POST /webhooks/xendit
func (h *WebhookHandler) HandleWebhook(c echo.Context) error {
	if !h.verifyCallbackToken(c) {
		return response.Unauthorized(c)
	}

	raw, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return response.BadRequest(c, "cannot read body")
	}

	// Parse event field saja dulu untuk routing.
	var envelope struct {
		Event string `json:"event"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		applogger.Warn("webhook: cannot parse body: %v", err)
		return response.OK(c, map[string]string{"status": "ignored"})
	}

	switch envelope.Event {
	case "qr.payment":
		var payload qrWebhookPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			applogger.Warn("webhook: cannot parse qr.payment: %v", err)
			return response.OK(c, map[string]string{"status": "ignored"})
		}
		if err := h.handleQRISPaid(c.Request().Context(), payload, raw); err != nil {
			applogger.Error("webhook: qr.payment failed: %v", err)
			if errors.Is(err, ErrOrderNotFound) {
				return response.OK(c, map[string]string{"status": "no_match"})
			}
			return response.Internal(c)
		}
		applogger.Webhook("qr.payment", payload.QRCode.ExternalID, payload.Status, payload.Amount)
		return response.OK(c, map[string]string{"status": "processed"})

	case "account.activated":
		var payload accountActivatedPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			applogger.Warn("webhook: cannot parse account.activated: %v", err)
			return response.OK(c, map[string]string{"status": "ignored"})
		}
		if err := h.handleAccountActivated(c.Request().Context(), payload); err != nil {
			applogger.Error("webhook: account.activated failed: %v", err)
			return response.Internal(c)
		}
		applogger.Webhook("account.activated", payload.Data.ID, payload.Data.Status, 0)
		return response.OK(c, map[string]string{"status": "processed"})

	default:
		applogger.Info("webhook: ignored event=%q", envelope.Event)
		return response.OK(c, map[string]string{"status": "ignored"})
	}
}

// verifyCallbackToken memverifikasi x-callback-token header dari Xendit.
func (h *WebhookHandler) verifyCallbackToken(c echo.Context) bool {
	if h.webhookToken == "" {
		return false
	}
	token := c.Request().Header.Get("x-callback-token")
	return token != "" && token == h.webhookToken
}

// handleQRISPaid menandai order sebagai PAID atomik dengan update xendit_payments.
func (h *WebhookHandler) handleQRISPaid(ctx context.Context, payload qrWebhookPayload, raw []byte) error {
	orderRef := payload.QRCode.ExternalID
	if orderRef == "" {
		return fmt.Errorf("webhook: missing external_id")
	}

	order, err := h.repo.FindByOrderID(ctx, orderRef)
	if err != nil {
		return err
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("webhook: begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	statusOK := paymentSuccess(payload.Status)

	if statusOK {
		if err := h.repo.MarkXenditPaymentPaid(ctx, tx, order.ID, "PAID", now, raw); err != nil && !errors.Is(err, ErrXenditPaymentNotFound) {
			return err
		}
		if order.Status != OrderStatusPaid {
			if err := h.repo.UpdateStatus(ctx, tx, order.ID, OrderStatusPaid, &now); err != nil {
				return err
			}
		}
	} else {
		if err := h.repo.MarkXenditPaymentPaid(ctx, tx, order.ID, "FAILED", now, raw); err != nil && !errors.Is(err, ErrXenditPaymentNotFound) {
			return err
		}
		if err := h.repo.UpdateStatus(ctx, tx, order.ID, OrderStatusFailed, nil); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("webhook: commit: %w", err)
	}
	committed = true
	return nil
}

// handleAccountActivated mengupdate xendit_status bisnis ke ACTIVE.
func (h *WebhookHandler) handleAccountActivated(ctx context.Context, payload accountActivatedPayload) error {
	accountID := payload.Data.ID
	if accountID == "" {
		return fmt.Errorf("webhook: account.activated missing data.id")
	}
	if err := h.repo.UpdateBusinessXenditStatus(ctx, accountID, StatusActive); err != nil {
		return fmt.Errorf("webhook: update xendit status: %w", err)
	}
	return nil
}

// paymentSuccess mengklasifikasi status string dari Xendit ke success/non-success.
func paymentSuccess(status string) bool {
	switch strings.ToUpper(status) {
	case "COMPLETED", "SUCCEEDED", "PAID", "SUCCESS":
		return true
	default:
		return false
	}
}

// RegisterWebhookRoute mendaftarkan webhook endpoint ke Echo.
func RegisterWebhookRoute(e *echo.Echo, h *WebhookHandler) {
	e.POST("/webhooks/xendit", h.HandleWebhook)
}
