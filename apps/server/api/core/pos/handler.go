// core/pos/handler.go
//
// Layer HTTP untuk domain pos.
// Handler hanya parsing input, manggil service, dan terjemahkan error ke response.
//
// Operator routes (RequireOperator):
//   POST /pos/orders                            → CreateOrder
//   GET  /pos/orders                            → ListMyOrders
//   PATCH /pos/orders/:order_id/items           → UpdateItems
//   POST /pos/orders/:order_id/checkout/begin   → BeginCheckout
//   POST /pos/orders/:order_id/checkout/confirm → ConfirmCheckout
//   POST /pos/orders/:order_id/void             → VoidOrder
//
// Owner routes (RequireOwner):
//   GET    /pos/sessions              → ListActiveSessions
//   DELETE /pos/sessions/:session_id  → ForceEndSession

package pos

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/response"
)

// Handler wraps Service untuk semua endpoint domain pos.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ----------------------------------------------------------------
// Context helpers
// ----------------------------------------------------------------

func businessIDFromCtx(c echo.Context) (uuid.UUID, error) {
	raw, _ := c.Get("business_id").(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid business_id in token")
	}
	return id, nil
}

func operatorIDFromCtx(c echo.Context) *uuid.UUID {
	raw, _ := c.Get("operator_id").(string)
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

func orderIDParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("order_id"))
	if err != nil {
		return uuid.Nil, errors.New("order_id tidak valid")
	}
	return id, nil
}

func sessionIDParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		return uuid.Nil, errors.New("session_id tidak valid")
	}
	return id, nil
}

// ----------------------------------------------------------------
// Order handlers
// ----------------------------------------------------------------

// POST /pos/orders
func (h *Handler) CreateOrder(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	operatorID := operatorIDFromCtx(c)

	var req CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	out, err := h.service.CreateOrder(c.Request().Context(), businessID, operatorID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, out)
}

// GET /pos/orders
func (h *Handler) ListMyOrders(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	operatorID := operatorIDFromCtx(c)
	if operatorID == nil {
		// Owner calling — return empty (orders are operator-scoped)
		return response.OK(c, []*Order{})
	}

	orders, err := h.service.ListMyOrders(c.Request().Context(), *operatorID, businessID)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, orders)
}

// PATCH /pos/orders/:order_id/items
func (h *Handler) UpdateItems(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	orderID, err := orderIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req UpdateItemsRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	out, err := h.service.UpdateItems(c.Request().Context(), businessID, orderID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, out)
}

// POST /pos/orders/:order_id/checkout/begin
func (h *Handler) BeginCheckout(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	orderID, err := orderIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	out, err := h.service.BeginCheckout(c.Request().Context(), businessID, orderID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, out)
}

// POST /pos/orders/:order_id/checkout/confirm
func (h *Handler) ConfirmCheckout(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	orderID, err := orderIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req ConfirmOrderRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	out, err := h.service.ConfirmCheckout(c.Request().Context(), businessID, orderID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, out)
}

// POST /pos/orders/:order_id/void
func (h *Handler) VoidOrder(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	orderID, err := orderIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.VoidOrder(c.Request().Context(), businessID, orderID); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

// ----------------------------------------------------------------
// Session handlers
// ----------------------------------------------------------------

// GET /pos/sessions
func (h *Handler) ListActiveSessions(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	sessions, err := h.service.ListActiveSessions(c.Request().Context(), businessID)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, sessions)
}

// DELETE /pos/sessions/:session_id
func (h *Handler) ForceEndSession(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	sessionID, err := sessionIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.ForceEndSession(c.Request().Context(), sessionID, businessID); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

// ----------------------------------------------------------------
// Error mapper
// ----------------------------------------------------------------

func mapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return response.NotFound(c)
	case errors.Is(err, ErrSessionNotFound):
		return response.NotFoundMsg(c, "Session tidak ditemukan")
	case errors.Is(err, ErrNotDraft):
		return c.JSON(http.StatusConflict, map[string]any{
			"success": false,
			"error":   "Order tidak dalam status DRAFT",
		})
	case errors.Is(err, ErrNotPending):
		return c.JSON(http.StatusConflict, map[string]any{
			"success": false,
			"error":   "Order tidak dalam status PENDING",
		})
	case errors.Is(err, ErrCheckoutNotStarted):
		return response.BadRequest(c, "Checkout belum dimulai — panggil begin-checkout terlebih dahulu")
	case errors.Is(err, ErrGestureTooFast):
		return c.JSON(http.StatusConflict, map[string]any{
			"success": false,
			"error":   "Konfirmasi terlalu cepat — tahan slide minimal 800ms",
		})
	case errors.Is(err, ErrEmptyItems):
		return response.BadRequest(c, err.Error())
	case errors.Is(err, ErrProductNotFound):
		return response.BadRequest(c, err.Error())
	default:
		return response.Internal(c)
	}
}
