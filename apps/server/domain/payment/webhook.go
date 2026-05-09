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
// Endpoint: POST /payment/xendit/webhook
// (pindah dari domain/xendit/handler.go saat merge)

package payment

import (
	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
)

// WebhookHandler menangani webhook dari Xendit.
// Dipisah dari Handler karena tidak butuh Service — langsung ke Repository
// dan harus bisa dipanggil tanpa auth middleware.
type WebhookHandler struct {
	repo         Repository
	webhookToken string // XENDIT_WEBHOOK_TOKEN dari config
}

func NewWebhookHandler(repo Repository, webhookToken string) *WebhookHandler {
	return &WebhookHandler{repo: repo, webhookToken: webhookToken}
}

// ----------------------------------------------------------------
// Payload shapes (Xendit QR Code webhook)
// ----------------------------------------------------------------

// qrWebhookPayload adalah subset dari Xendit QR Code webhook body.
// Field lain disimpan di raw_payload JSONB — tidak perlu di-parse semua.
type qrWebhookPayload struct {
	Event    string `json:"event"`
	BusinessID string `json:"business_id"` // Xendit's business_id (sub-account)
	Data     struct {
		ReferenceID string `json:"reference_id"` // ini adalah order_id kita (QIOS-YYYYMMDD-xxxx)
		Amount      int64  `json:"amount"`
		Status      string `json:"status"`
		QRString    string `json:"qr_string"`
	} `json:"data"`
}

// ----------------------------------------------------------------
// Handler
// ----------------------------------------------------------------

// HandleWebhook menangani semua webhook Xendit di satu endpoint.
// Routing ke handler spesifik berdasarkan field "event" di body.
// POST /payment/xendit/webhook
func (h *WebhookHandler) HandleWebhook(c echo.Context) error {
	// Step 1: Verifikasi callback token
	if !h.verifyCallbackToken(c) {
		return response.Unauthorized(c)
	}

	// TODO: implement
	// Step 2: Parse event type dari body
	// Step 3: Route ke handler berdasarkan event:
	//   - "qr.payment" → handleQRISPaid
	//   - "account.activated" → handleAccountActivated (update businesses.xendit_status)
	//   - lain-lain → log dan return 200 (Xendit expects 200 untuk semua webhook)
	return response.NotImplemented(c, "webhook handler not yet implemented")
}

// verifyCallbackToken memverifikasi x-callback-token header dari Xendit.
// Xendit mengirim token ini di setiap webhook — harus cocok dengan XENDIT_WEBHOOK_TOKEN.
func (h *WebhookHandler) verifyCallbackToken(c echo.Context) bool {
	token := c.Request().Header.Get("x-callback-token")
	return token != "" && token == h.webhookToken
}

// handleQRISPaid menangani event pembayaran QRIS sukses.
// Flow:
//   1. Lookup pos_order by reference_id (= order_id)
//   2. Validate amount cocok
//   3. Begin tx
//   4. INSERT xendit_payments
//   5. UPDATE pos_orders.status = PAID
//   6. Commit
func (h *WebhookHandler) handleQRISPaid(c echo.Context, payload qrWebhookPayload) error {
	// TODO: implement
	panic("not implemented")
}

// RegisterWebhookRoute mendaftarkan webhook endpoint ke Echo.
// Tidak pakai authMiddleware — Xendit tidak bisa kirim Bearer token.
func RegisterWebhookRoute(e *echo.Echo, h *WebhookHandler) {
	e.POST("/payment/xendit/webhook", h.HandleWebhook)
}
