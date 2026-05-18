// domain/payment/handler.go
//
// HTTP handlers untuk payment domain.
// Handler hanya parse input, call service, map error ke response.
// Tidak ada SQL, tidak ada business logic.
//
// Endpoints:
//   POST /transactions               → CreateOrder (operator)
//   GET  /transactions               → ListOrders (owner)
//   GET  /transactions/:id           → GetOrder (owner)
//   POST /transactions/:id/complete  → CompleteCashOrder (operator)
//   POST /payment/xendit/connect     → ConnectXendit (owner)
//   GET  /payment/xendit/status      → GetXenditStatus (owner)

package payment

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
)

// Handler adalah HTTP handler untuk payment domain.
type Handler struct {
	service Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// ----------------------------------------------------------------
// Context helpers
// ----------------------------------------------------------------

func paymentBusinessIDFromCtx(c echo.Context) (uuid.UUID, error) {
	raw, _ := c.Get("business_id").(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid business_id in token")
	}
	return id, nil
}

func paymentOperatorIDFromCtx(c echo.Context) (uuid.UUID, error) {
	raw, _ := c.Get("operator_id").(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid operator_id in token")
	}
	return id, nil
}

// ----------------------------------------------------------------
// Validators
// ----------------------------------------------------------------

func validateCreateOrderRequest(req *CreateOrderRequest) string {
	if req.PaymentMethod == "" {
		return "payment_method wajib diisi"
	}
	switch req.PaymentMethod {
	case PaymentMethodCash, PaymentMethodQRIS, PaymentMethodEWallet, PaymentMethodVirtualAccount:
		// valid
	default:
		return "payment_method tidak valid"
	}
	if len(req.Items) == 0 {
		return "items wajib diisi minimal satu"
	}
	for i, item := range req.Items {
		if item.ProductID == uuid.Nil {
			return "product_id wajib diisi di setiap item"
		}
		if item.Quantity <= 0 {
			return "quantity harus lebih dari 0 di item ke-" + string(rune('1'+i))
		}
	}
	return ""
}

func validateConnectXenditRequest(req *connectXenditRequest) string {
	if req.Email == "" {
		return "email wajib diisi"
	}
	if req.BusinessName == "" {
		return "business_name wajib diisi"
	}
	return ""
}

// ----------------------------------------------------------------
// Transaction handlers
// ----------------------------------------------------------------

// POST /transactions
func (h *Handler) CreateOrder(c echo.Context) error {
	businessID, err := paymentBusinessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	operatorID, err := paymentOperatorIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if msg := validateCreateOrderRequest(&req); msg != "" {
		return response.BadRequest(c, msg)
	}

	order, err := h.service.CreateOrder(c.Request().Context(), businessID, operatorID, req)
	if err != nil {
		log.Printf("CreateOrder error: %+v", err)
		return mapPaymentError(c, err)
	}
	return response.Created(c, order)
}

// GET /transactions
func (h *Handler) ListOrders(c echo.Context) error {
	businessID, err := paymentBusinessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	filter := ListOrdersFilter{
		Status:    c.QueryParam("status"),
		StartDate: c.QueryParam("start_date"),
		EndDate:   c.QueryParam("end_date"),
		Page:      1,
		Limit:     20,
	}

	orders, total, err := h.service.ListOrders(c.Request().Context(), businessID, filter)
	if err != nil {
		return response.Internal(c)
	}
	return response.OKWithMeta(c, orders, map[string]any{"total": total})
}

// GET /transactions/:id
func (h *Handler) GetOrder(c echo.Context) error {
	businessID, err := paymentBusinessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "transaction id tidak valid")
	}

	order, err := h.service.GetOrder(c.Request().Context(), businessID, orderID)
	if err != nil {
		return mapPaymentError(c, err)
	}
	return response.OK(c, order)
}

// POST /transactions/:id/complete
func (h *Handler) CompleteCashOrder(c echo.Context) error {
	businessID, err := paymentBusinessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "transaction id tidak valid")
	}

	order, err := h.service.CompleteCashOrder(c.Request().Context(), businessID, orderID)
	if err != nil {
		return mapPaymentError(c, err)
	}
	return response.OK(c, order)
}

// ----------------------------------------------------------------
// Xendit connect/status
// ----------------------------------------------------------------

// connectXenditRequest adalah body POST /payment/xendit/connect.
// Email dan business_name dibutuhkan untuk membuat sub-account Xendit.
// Biasanya diisi dari data bisnis yang sudah ada — owner tinggal konfirmasi.
type connectXenditRequest struct {
	Email        string `json:"email"`
	BusinessName string `json:"business_name"`
}

// POST /payment/xendit/connect
func (h *Handler) ConnectXendit(c echo.Context) error {
	businessID, err := paymentBusinessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req connectXenditRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if msg := validateConnectXenditRequest(&req); msg != "" {
		return response.BadRequest(c, msg)
	}

	result, err := h.service.ConnectXendit(c.Request().Context(), businessID, req.Email, req.BusinessName)
	if err != nil {
		return mapPaymentError(c, err)
	}
	return response.OK(c, result)
}

// GET /payment/xendit/status
func (h *Handler) GetXenditStatus(c echo.Context) error {
	businessID, err := paymentBusinessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	result, err := h.service.GetXenditStatus(c.Request().Context(), businessID)
	if err != nil {
		return mapPaymentError(c, err)
	}
	return response.OK(c, result)
}

// ----------------------------------------------------------------
// Error mapper
// ----------------------------------------------------------------

func mapPaymentError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrOrderNotFound):
		return response.NotFoundMsg(c, "Order tidak ditemukan")
	case errors.Is(err, ErrOrderAlreadyPaid):
		return response.Conflict(c, "Order sudah dibayar")
	case errors.Is(err, ErrInvalidStatus):
		return response.BadRequest(c, err.Error())
	case errors.Is(err, ErrProductNotFound):
		return response.BadRequest(c, "Salah satu produk tidak ditemukan")
	case errors.Is(err, ErrInvalidTotal):
		return response.BadRequest(c, err.Error())
	case errors.Is(err, ErrXenditNotActive):
		return response.UnprocessableEntity(c, "Akun Xendit bisnis belum aktif untuk pembayaran QRIS")
	case errors.Is(err, ErrXenditAlreadyActive):
		return response.Conflict(c, "Akun Xendit sudah terdaftar atau aktif")
	case errors.Is(err, ErrBusinessNotFound):
		return response.NotFoundMsg(c, "Bisnis tidak ditemukan")
	default:
		return response.Internal(c)
	}
}

// ----------------------------------------------------------------
// Routes
// ----------------------------------------------------------------

func RegisterRoutes(e *echo.Echo, h *Handler, authMw echo.MiddlewareFunc) {
	// Transaction endpoints
	t := e.Group("/transactions", authMw)
	t.POST("", h.CreateOrder, appmiddleware.RequireOperator)
	t.GET("", h.ListOrders, appmiddleware.RequireOwner)
	t.GET("/:id", h.GetOrder, appmiddleware.RequireOwner)
	t.POST("/:id/complete", h.CompleteCashOrder, appmiddleware.RequireOperator)

	// Xendit platform endpoints (owner only)
	x := e.Group("/payment/xendit", authMw, appmiddleware.RequireOwner)
	x.POST("/connect", h.ConnectXendit)
	x.GET("/status", h.GetXenditStatus)
}
