// core/admin/handler.go
//
// Layer HTTP untuk domain admin (qios-admin panel, service-to-service).
// Autentikasi via X-Admin-Key header — lihat pkg/middleware.RequireAdminKey.
//
// Routes:
//   GET    /admin/owners                                    → ListOwners
//   POST   /admin/owners                                    → CreateOwner
//   GET    /admin/owners/:owner_id                          → GetOwner
//   PATCH  /admin/owners/:owner_id                          → UpdateOwner
//   PATCH  /admin/owners/:owner_id/status                   → SetOwnerStatus
//   POST   /admin/owners/:owner_id/credential               → SetOwnerCredential
//   GET    /admin/owners/:owner_id/products                 → ListOwnerProducts
//   POST   /admin/owners/:owner_id/products                 → CreateProduct
//   GET    /admin/products/:product_id                      → GetProduct
//   PATCH  /admin/products/:product_id                      → UpdateProduct
//   DELETE /admin/products/:product_id                      → DeleteProduct
//   PUT    /admin/products/:product_id/recipe               → UpdateProductRecipe
//   DELETE /admin/owners/:owner_id/operators/:operator_id   → DeleteOperator
//   GET    /admin/transactions                              → ListTransactions
//   POST   /admin/transactions/:transaction_id/void         → VoidTransaction

package admin

import (
	"errors"
	"net/http"
	"strconv"

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
// Param helpers
// ----------------------------------------------------------------

func ownerIDParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("owner_id"))
	if err != nil {
		return uuid.Nil, errors.New("owner_id tidak valid")
	}
	return id, nil
}

func productIDParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		return uuid.Nil, errors.New("product_id tidak valid")
	}
	return id, nil
}

func operatorIDParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("operator_id"))
	if err != nil {
		return uuid.Nil, errors.New("operator_id tidak valid")
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
	switch {
	case errors.Is(err, ErrOwnerNotFound),
		errors.Is(err, ErrBusinessNotFound),
		errors.Is(err, ErrProductNotFound),
		errors.Is(err, ErrOperatorNotFound),
		errors.Is(err, ErrTransactionNotFound):
		return response.NotFound(c)
	case errors.Is(err, ErrEmailTaken):
		return response.BadRequest(c, "email already registered")
	case errors.Is(err, ErrTransactionNotPending):
		return c.JSON(http.StatusConflict, map[string]any{
			"success": false,
			"error":   "transaction is not in pending status",
		})
	default:
		return response.Internal(c)
	}
}

// ----------------------------------------------------------------
// Owner handlers
// ----------------------------------------------------------------

// GET /admin/owners?page=1&limit=20
func (h *Handler) ListOwners(c echo.Context) error {
	page, limit := 1, 20
	if s := c.QueryParam("page"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 1 {
			page = n
		}
	}
	if s := c.QueryParam("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 100 {
			limit = n
		}
	}
	result, err := h.service.ListOwners(c.Request().Context(), page, limit)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, result)
}

// POST /admin/owners
func (h *Handler) CreateOwner(c echo.Context) error {
	var req CreateBusinessRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	b, err := h.service.CreateOwner(c.Request().Context(), req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, b)
}

// GET /admin/owners/:owner_id
func (h *Handler) GetOwner(c echo.Context) error {
	id, err := ownerIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	o, err := h.service.GetOwner(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, o)
}

// PATCH /admin/owners/:owner_id
func (h *Handler) UpdateOwner(c echo.Context) error {
	id, err := ownerIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req UpdateBusinessRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	b, err := h.service.UpdateOwner(c.Request().Context(), id, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, b)
}

// PATCH /admin/owners/:owner_id/status
func (h *Handler) SetOwnerStatus(c echo.Context) error {
	id, err := ownerIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req SetOwnerStatusRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if err := h.service.SetOwnerStatus(c.Request().Context(), id, req.Enabled); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

// POST /admin/owners/:owner_id/credential
func (h *Handler) SetOwnerCredential(c echo.Context) error {
	id, err := ownerIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req SetOwnerCredentialRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.SetOwnerCredential(c.Request().Context(), id, req); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

// ----------------------------------------------------------------
// Product handlers
// ----------------------------------------------------------------

// GET /admin/owners/:owner_id/products
func (h *Handler) ListOwnerProducts(c echo.Context) error {
	businessID, err := ownerIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	products, err := h.service.ListProducts(c.Request().Context(), businessID)
	if err != nil {
		return response.Internal(c)
	}
	if products == nil {
		products = []*AdminProduct{}
	}
	return response.OK(c, map[string]any{"products": products})
}

// POST /admin/owners/:owner_id/products
func (h *Handler) CreateProduct(c echo.Context) error {
	businessID, err := ownerIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req AdminCreateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	p, err := h.service.CreateProduct(c.Request().Context(), adminID, businessID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, p)
}

// GET /admin/products/:product_id
func (h *Handler) GetProduct(c echo.Context) error {
	productID, err := productIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	p, err := h.service.GetProduct(c.Request().Context(), productID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, p)
}

// PATCH /admin/products/:product_id
func (h *Handler) UpdateProduct(c echo.Context) error {
	adminID, err := adminIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c)
	}
	productID, err := productIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req AdminUpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	p, err := h.service.UpdateProduct(c.Request().Context(), adminID, productID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, p)
}

// DELETE /admin/products/:product_id
func (h *Handler) DeleteProduct(c echo.Context) error {
	adminID, err := adminIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c)
	}
	productID, err := productIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.DeleteProduct(c.Request().Context(), adminID, productID); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

// PUT /admin/products/:product_id/recipe
func (h *Handler) UpdateProductRecipe(c echo.Context) error {
	productID, err := productIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req UpdateRecipeRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.UpdateProductRecipe(c.Request().Context(), productID, req); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

// ----------------------------------------------------------------
// Operator handlers
// ----------------------------------------------------------------

// DELETE /admin/owners/:owner_id/operators/:operator_id
func (h *Handler) DeleteOperator(c echo.Context) error {
	businessID, err := ownerIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	operatorID, err := operatorIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.DeleteOperator(c.Request().Context(), adminID, businessID, operatorID); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

// ----------------------------------------------------------------
// Transaction handlers
// ----------------------------------------------------------------

// GET /admin/transactions
func (h *Handler) ListTransactions(c echo.Context) error {
	f := AdminListTransactionsFilter{
		Status: c.QueryParam("status"),
		Page:   1,
		Limit:  20,
	}

	if s := c.QueryParam("business_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return response.BadRequest(c, "business_id tidak valid")
		}
		f.BusinessID = &id
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

	result, err := h.service.ListTransactions(c.Request().Context(), f)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, result)
}

// POST /admin/transactions/:transaction_id/void
func (h *Handler) VoidTransaction(c echo.Context) error {
	adminID, err := adminIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c)
	}
	txID, err := transactionIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.VoidTransaction(c.Request().Context(), adminID, txID); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}
