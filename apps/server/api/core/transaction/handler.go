// core/transaction/handler.go
//
// Layer HTTP untuk domain transaction — read-only.
//
// Owner routes:
//   GET /transactions              → List (dengan filter & pagination)
//
// Owner + operator routes:
//   GET /transactions/:transaction_id → GetByID

package transaction

import (
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/response"
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

func transactionIDParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("transaction_id"))
	if err != nil {
		return uuid.Nil, errors.New("transaction_id tidak valid")
	}
	return id, nil
}

func mapServiceError(c echo.Context, err error) error {
	if errors.Is(err, ErrNotFound) {
		return response.NotFound(c)
	}
	return response.Internal(c)
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

	validStatuses := map[string]bool{
		"DRAFT": true, "CONFIRMED": true, "VOIDED": true,
	}
	validPaymentMethods := map[string]bool{
		"CASH": true, "QRIS": true, "EWALLET": true,
		"VIRTUAL_ACCOUNT": true, "TRANSFER": true,
	}

	f := ListFilter{}
	if s := c.QueryParam("status"); s != "" {
		if !validStatuses[s] {
			return response.BadRequest(c, "status tidak valid, pilih: DRAFT, CONFIRMED, VOIDED")
		}
		f.Status = s
	}
	if pm := c.QueryParam("payment_method"); pm != "" {
		if !validPaymentMethods[pm] {
			return response.BadRequest(c, "payment_method tidak valid, pilih: CASH, QRIS, EWALLET, VIRTUAL_ACCOUNT, TRANSFER")
		}
		f.PaymentMethod = pm
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
			return response.BadRequest(c, "limit harus antara 1-100")
		}
		f.Limit = n
	}

	result, err := h.service.List(c.Request().Context(), businessID, f)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, result)
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
