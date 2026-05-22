// core/admin/handler.go
//
// Layer HTTP untuk domain admin.
// Handler hanya parsing input, manggil service, dan terjemahkan error ke response.
//
// Public routes:
//   POST /admin/auth/login    → Login
//   POST /admin/auth/refresh  → Refresh
//   POST /admin/auth/logout   → Logout
//
// Protected routes (RequireAuth + RequireAdmin):
//   GET  /admin/me                                         → Me
//   GET  /admin/businesses                                 → ListBusinesses
//   POST /admin/businesses                                 → CreateBusiness
//   GET  /admin/businesses/:business_id                    → GetBusiness
//   PATCH /admin/businesses/:business_id                   → UpdateBusiness
//   GET  /admin/businesses/:business_id/products           → ListProducts
//   POST /admin/businesses/:business_id/products           → CreateProduct
//   PATCH /admin/products/:product_id                      → UpdateProduct
//   DELETE /admin/products/:product_id                     → DeleteProduct
//   DELETE /admin/businesses/:business_id/operators/:operator_id → DeleteOperator
//   GET  /admin/transactions                               → ListTransactions
//   POST /admin/transactions/:transaction_id/void          → VoidTransaction

package admin

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/response"
)

const refreshTokenCookieName = "admin_refresh_token"

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ----------------------------------------------------------------
// Cookie helpers
// ----------------------------------------------------------------

func setRefreshCookie(c echo.Context, plain string, expiry time.Duration) {
	c.SetCookie(&http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    plain,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(expiry),
		Path:     "/",
	})
}

func clearRefreshCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		Path:     "/",
	})
}

// ----------------------------------------------------------------
// Context helpers
// ----------------------------------------------------------------

func adminIDFromCtx(c echo.Context) (uuid.UUID, error) {
	raw, _ := c.Get("user_id").(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid admin_id in token")
	}
	return id, nil
}

func businessIDParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("business_id"))
	if err != nil {
		return uuid.Nil, errors.New("business_id tidak valid")
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
	case errors.Is(err, ErrAdminNotFound),
		errors.Is(err, ErrBusinessNotFound),
		errors.Is(err, ErrProductNotFound),
		errors.Is(err, ErrOperatorNotFound),
		errors.Is(err, ErrTransactionNotFound):
		return response.NotFound(c)
	case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrRefreshNotFound):
		return response.Unauthorized(c)
	case errors.Is(err, ErrSessionExpired):
		return response.Unauthorized(c)
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
// Auth handlers
// ----------------------------------------------------------------

// POST /admin/auth/login
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	out, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return mapServiceError(c, err)
	}

	setRefreshCookie(c, out.RefreshToken, out.RefreshExpiry)
	return response.OK(c, map[string]string{"access_token": out.AccessToken})
}

// POST /admin/auth/refresh
func (h *Handler) Refresh(c echo.Context) error {
	cookie, err := c.Cookie(refreshTokenCookieName)
	if err != nil || cookie.Value == "" {
		return response.Unauthorized(c)
	}

	out, err := h.service.Refresh(c.Request().Context(), cookie.Value)
	if err != nil {
		clearRefreshCookie(c)
		return mapServiceError(c, err)
	}

	setRefreshCookie(c, out.RefreshToken, out.RefreshExpiry)
	return response.OK(c, map[string]string{"access_token": out.AccessToken})
}

// POST /admin/auth/logout
func (h *Handler) Logout(c echo.Context) error {
	cookie, _ := c.Cookie(refreshTokenCookieName)
	plain := ""
	if cookie != nil {
		plain = cookie.Value
	}
	_ = h.service.Logout(c.Request().Context(), plain)
	clearRefreshCookie(c)
	return response.NoContent(c)
}

// GET /admin/me
func (h *Handler) Me(c echo.Context) error {
	adminID, err := adminIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c)
	}

	admin, err := h.service.Me(c.Request().Context(), adminID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, admin)
}

// ----------------------------------------------------------------
// Business handlers
// ----------------------------------------------------------------

// GET /admin/businesses
func (h *Handler) ListBusinesses(c echo.Context) error {
	businesses, err := h.service.ListBusinesses(c.Request().Context())
	if err != nil {
		return response.Internal(c)
	}
	if businesses == nil {
		businesses = []*Business{}
	}
	return response.OK(c, map[string]any{"businesses": businesses})
}

// POST /admin/businesses
func (h *Handler) CreateBusiness(c echo.Context) error {
	var req CreateBusinessRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	b, err := h.service.CreateBusiness(c.Request().Context(), req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, b)
}

// GET /admin/businesses/:business_id
func (h *Handler) GetBusiness(c echo.Context) error {
	id, err := businessIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	b, err := h.service.GetBusiness(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, b)
}

// PATCH /admin/businesses/:business_id
func (h *Handler) UpdateBusiness(c echo.Context) error {
	id, err := businessIDParam(c)
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

	b, err := h.service.UpdateBusiness(c.Request().Context(), id, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, b)
}

// ----------------------------------------------------------------
// Product handlers
// ----------------------------------------------------------------

// GET /admin/businesses/:business_id/products
func (h *Handler) ListProducts(c echo.Context) error {
	businessID, err := businessIDParam(c)
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

// POST /admin/businesses/:business_id/products
func (h *Handler) CreateProduct(c echo.Context) error {
	businessID, err := businessIDParam(c)
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

	p, err := h.service.CreateProduct(c.Request().Context(), businessID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, p)
}

// PATCH /admin/products/:product_id
func (h *Handler) UpdateProduct(c echo.Context) error {
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

	p, err := h.service.UpdateProduct(c.Request().Context(), productID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, p)
}

// DELETE /admin/products/:product_id
func (h *Handler) DeleteProduct(c echo.Context) error {
	productID, err := productIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.DeleteProduct(c.Request().Context(), productID); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

// ----------------------------------------------------------------
// Operator handlers
// ----------------------------------------------------------------

// DELETE /admin/businesses/:business_id/operators/:operator_id
func (h *Handler) DeleteOperator(c echo.Context) error {
	businessID, err := businessIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	operatorID, err := operatorIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.DeleteOperator(c.Request().Context(), businessID, operatorID); err != nil {
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
	txID, err := transactionIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.VoidTransaction(c.Request().Context(), txID); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}
