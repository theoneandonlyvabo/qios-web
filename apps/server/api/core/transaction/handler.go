// core/transaction/handler.go
//
// Layer HTTP untuk domain transaction.
//
// Owner routes:
//   GET  /transactions              → List (dengan filter & pagination)
//
// Owner + operator routes:
//   POST /transactions              → Create
//   GET  /transactions/:id         → GetByID
//   POST /transactions/:id/confirm → Confirm (tandai paid + set payment_method)
//   POST /transactions/:id/void    → Void (batalkan pending order)

package transaction

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/response"
)

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

func transactionIDParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("transaction_id"))
	if err != nil {
		return uuid.Nil, errors.New("transaction_id tidak valid")
	}
	return id, nil
}

func mapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return response.NotFound(c)
	case errors.Is(err, ErrNotPending):
		return c.JSON(http.StatusConflict, map[string]any{
			"success": false,
			"error":   "transaction is not in pending status",
		})
	case errors.Is(err, ErrProductNotFound):
		return response.BadRequest(c, err.Error())
	default:
		return response.Internal(c)
	}
}

// ----------------------------------------------------------------
// Handlers
// ----------------------------------------------------------------

// GET /transactions
// Owner-only. Filter: start_date, end_date, status, payment_method, operator_id, page, limit.
func (h *Handler) List(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	f := ListFilter{
		Status:        c.QueryParam("status"),
		PaymentMethod: c.QueryParam("payment_method"),
	}

	if s := c.QueryParam("start_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.BadRequest(c, "start_date harus format YYYY-MM-DD")
		}
		f.StartDate = &t
	}
	if s := c.QueryParam("end_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.BadRequest(c, "end_date harus format YYYY-MM-DD")
		}
		f.EndDate = &t
	}
	if s := c.QueryParam("operator_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return response.BadRequest(c, "operator_id tidak valid")
		}
		f.OperatorID = &id
	}
	if s := c.QueryParam("page"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			return response.BadRequest(c, "page harus angka positif")
		}
		f.Page = n
	}
	if s := c.QueryParam("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 100 {
			return response.BadRequest(c, "limit harus antara 1–100")
		}
		f.Limit = n
	}

	result, err := h.service.List(c.Request().Context(), businessID, f)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, result)
}

// POST /transactions
func (h *Handler) Create(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	operatorID := operatorIDFromCtx(c)

	out, err := h.service.Create(c.Request().Context(), businessID, operatorID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, out)
}

// GET /transactions/:transaction_id
func (h *Handler) GetByID(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	txID, err := transactionIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	out, err := h.service.GetByID(c.Request().Context(), businessID, txID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, out)
}

// POST /transactions/:transaction_id/confirm
func (h *Handler) Confirm(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	txID, err := transactionIDParam(c)
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

	out, err := h.service.Confirm(c.Request().Context(), businessID, txID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, out)
}

// POST /transactions/:transaction_id/void
func (h *Handler) Void(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	txID, err := transactionIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.Void(c.Request().Context(), businessID, txID); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}
