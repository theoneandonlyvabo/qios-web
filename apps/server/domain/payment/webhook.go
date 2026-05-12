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
	"log"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

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
// Payload shapes (Xendit QR Code webhook — legacy event "qr.payment")
// ----------------------------------------------------------------

// qrWebhookPayload adalah subset dari Xendit QR Code webhook body.
// Field lain disimpan di raw_payload JSONB — tidak perlu di-parse semua.
type qrWebhookPayload struct {
	Event      string `json:"event"`
	BusinessID string `json:"business_id"` // Xendit sub-account id
	Created    string `json:"created"`
	Data       struct {
		ID          string `json:"id"`           // Xendit QR/payment id
		ReferenceID string `json:"reference_id"` // pos_orders.order_id
		ExternalID  string `json:"external_id"`  // fallback ref jika reference_id kosong
		Amount      int64  `json:"amount"`
		Status      string `json:"status"`
		QRString    string `json:"qr_string"`
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

	var payload qrWebhookPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		// Xendit expects 200 untuk semua webhook yang sukses divalidasi.
		// Payload tidak dikenali kita log tapi tidak fail.
		log.Printf("xendit webhook: cannot parse body: %v", err)
		return response.OK(c, map[string]string{"status": "ignored"})
	}

	switch payload.Event {
	case "qr.payment":
		if err := h.handleQRISPaid(c.Request().Context(), payload, raw); err != nil {
			log.Printf("xendit webhook: qr.payment failed: %v", err)
			if errors.Is(err, ErrOrderNotFound) {
				// Tidak ada order — possibly test event. Return 200 supaya Xendit tidak retry.
				return response.OK(c, map[string]string{"status": "no_match"})
			}
			return response.Internal(c)
		}
		return response.OK(c, map[string]string{"status": "processed"})

	default:
		// Event lain (account.activated, dll) belum di-handle. Tetap 200.
		log.Printf("xendit webhook: ignored event=%q", payload.Event)
		return response.OK(c, map[string]string{"status": "ignored"})
	}
}

// verifyCallbackToken memverifikasi x-callback-token header dari Xendit.
// Xendit mengirim token ini di setiap webhook — harus cocok dengan XENDIT_WEBHOOK_TOKEN.
func (h *WebhookHandler) verifyCallbackToken(c echo.Context) bool {
	if h.webhookToken == "" {
		return false
	}
	token := c.Request().Header.Get("x-callback-token")
	return token != "" && token == h.webhookToken
}

// handleQRISPaid menandai order sebagai PAID atomik dengan update xendit_payments.
func (h *WebhookHandler) handleQRISPaid(ctx context.Context, payload qrWebhookPayload, raw []byte) error {
	orderRef := payload.Data.ReferenceID
	if orderRef == "" {
		orderRef = payload.Data.ExternalID
	}
	if orderRef == "" {
		return fmt.Errorf("webhook: missing reference_id")
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
	statusOK := paymentSuccess(payload.Data.Status)

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

// paymentSuccess mengklasifikasi status string dari Xendit ke success/non-success.
// Legacy QR API menggunakan COMPLETED/SUCCEEDED untuk sukses.
func paymentSuccess(status string) bool {
	switch strings.ToUpper(status) {
	case "COMPLETED", "SUCCEEDED", "PAID", "SUCCESS":
		return true
	default:
		return false
	}
}

// RegisterWebhookRoute mendaftarkan webhook endpoint ke Echo.
// Tidak pakai authMiddleware — Xendit tidak bisa kirim Bearer token.
func RegisterWebhookRoute(e *echo.Echo, h *WebhookHandler) {
	e.POST("/webhooks/xendit", h.HandleWebhook)
}
