// domain/operator/handler.go
//
// Layer HTTP untuk domain operator.
// Handler hanya parsing input, manggil service, dan terjemahkan error ke response.
// Tidak ada query DB, tidak ada hashing — semua di service / repository.
//
// Owner routes  (group /business/operators, butuh JWT owner):
//   POST   /                    → CreateOperator
//   GET    /                    → GetOperators
//   GET    /:operator_id        → GetOperatorByID
//   PUT    /:operator_id        → UpdateOperator
//   DELETE /:operator_id        → DeleteOperator
//   POST   /:operator_id/regenerate-qr → RegenerateQR
//
// Operator auth routes (group /kasir/auth, public):
//   POST /login                 → LoginWithCredentials
//   POST /login/qr              → LoginWithQR
//
// CLAUDE.md API table memasang owner CRUD di /business/operators.
// Path dipertahankan supaya konsisten dengan kontrak yang sudah ada.

package operator

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/middleware"
	"github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ----------------------------------------------------------------
// Helper context readers
// ----------------------------------------------------------------

func businessIDFromCtx(c echo.Context) (uuid.UUID, error) {
	raw, _ := c.Get("business_id").(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid business_id in token")
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

// ----------------------------------------------------------------
// Validators
// ----------------------------------------------------------------

// operatorCodePattern membatasi operator_code ke karakter alfanumerik,
// dash, dan underscore — supaya aman di URL dan QR.
var operatorCodePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)

func validateCreateRequest(r *CreateOperatorRequest) string {
	if r.Name == "" {
		return "Nama operator wajib diisi"
	}
	if len(r.Name) > 255 {
		return "Nama operator terlalu panjang"
	}
	if !operatorCodePattern.MatchString(r.OperatorCode) {
		return "operator_code harus 3-64 karakter alfanumerik, dash, atau underscore"
	}
	if len(r.Password) < 6 {
		return "Password minimal 6 karakter"
	}
	if len(r.Password) > 128 {
		return "Password maksimal 128 karakter"
	}
	return ""
}

func validateUpdateRequest(r *UpdateOperatorRequest) string {
	if r.Name != nil {
		if *r.Name == "" {
			return "Nama operator tidak boleh kosong"
		}
		if len(*r.Name) > 255 {
			return "Nama operator terlalu panjang"
		}
	}
	return ""
}

// ----------------------------------------------------------------
// Owner-facing handlers
// ----------------------------------------------------------------

// POST /business/operators
func (h *Handler) CreateOperator(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req CreateOperatorRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if msg := validateCreateRequest(&req); msg != "" {
		return response.BadRequest(c, msg)
	}

	out, err := h.service.CreateOperator(c.Request().Context(), businessID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, out)
}

// GET /business/operators
type listResponse struct {
	Operators []OperatorResponse `json:"operators"`
	SlotUsed  int                `json:"slot_used"`
	SlotMax   int                `json:"slot_max"`
}

func (h *Handler) GetOperators(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	ops, err := h.service.GetOperators(c.Request().Context(), businessID)
	if err != nil {
		return response.Internal(c)
	}
	used, max, err := h.service.GetSlotInfo(c.Request().Context(), businessID)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, listResponse{
		Operators: ops,
		SlotUsed:  used,
		SlotMax:   max,
	})
}

// GET /business/operators/:operator_id
func (h *Handler) GetOperatorByID(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	operatorID, err := operatorIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	op, err := h.service.GetOperatorByID(c.Request().Context(), businessID, operatorID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, op)
}

// PUT /business/operators/:operator_id
func (h *Handler) UpdateOperator(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	operatorID, err := operatorIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req UpdateOperatorRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if msg := validateUpdateRequest(&req); msg != "" {
		return response.BadRequest(c, msg)
	}

	op, err := h.service.UpdateOperator(c.Request().Context(), businessID, operatorID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, op)
}

// DELETE /business/operators/:operator_id
func (h *Handler) DeleteOperator(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
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

// POST /business/operators/:operator_id/regenerate-qr
func (h *Handler) RegenerateQR(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	operatorID, err := operatorIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	out, err := h.service.RegenerateQR(c.Request().Context(), businessID, operatorID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, out)
}

// ----------------------------------------------------------------
// Operator-facing handlers (kasir auth — public)
// ----------------------------------------------------------------

// loginRequest membungkus credentials login dengan business identifier.
// Kasir tidak punya konteks bisnis sebelum login — owner harus
// memasukkan slug bisnis di device atau dilewatkan via header.
type loginCredentialsRequest struct {
	BusinessID   string `json:"business_id"`
	OperatorCode string `json:"operator_code"`
	Password     string `json:"password"`
}

// POST /kasir/auth/login
func (h *Handler) LoginWithCredentials(c echo.Context) error {
	var req loginCredentialsRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.BusinessID == "" || req.OperatorCode == "" || req.Password == "" {
		return response.BadRequest(c, "business_id, operator_code, dan password wajib diisi")
	}
	businessID, err := uuid.Parse(req.BusinessID)
	if err != nil {
		return response.BadRequest(c, "business_id tidak valid")
	}

	out, err := h.service.LoginWithCredentials(c.Request().Context(), businessID, OperatorLoginRequest{
		OperatorCode: req.OperatorCode,
		Password:     req.Password,
	})
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, out)
}

// POST /kasir/auth/logout
// Stateless JWT logout — client discards token. Middleware validates JWT before reaching here.
func (h *Handler) Logout(c echo.Context) error {
	return response.NoContent(c)
}

// POST /kasir/auth/login/qr
func (h *Handler) LoginWithQR(c echo.Context) error {
	var req QRLoginRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.QRToken == "" {
		return response.BadRequest(c, "qr_token wajib diisi")
	}

	out, err := h.service.LoginWithQR(c.Request().Context(), req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, out)
}

// ----------------------------------------------------------------
// Error mapper
// ----------------------------------------------------------------

func mapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return response.NotFoundMsg(c, "Operator tidak ditemukan")
	case errors.Is(err, ErrCodeTaken):
		return response.Conflict(c, "operator_code sudah digunakan di bisnis ini")
	case errors.Is(err, ErrLimitReached):
		return c.JSON(http.StatusConflict, map[string]any{
			"success": false,
			"error":   "Kuota operator pada plan saat ini sudah penuh",
		})
	case errors.Is(err, ErrInvalidCredentials):
		return response.UnauthorizedMsg(c, "operator_code atau password salah")
	case errors.Is(err, ErrInactive):
		return response.ForbiddenMsg(c, "Akun operator dinonaktifkan")
	default:
		return response.InternalError(c, err.Error())
	}
}

// ----------------------------------------------------------------
// Routes
// ----------------------------------------------------------------

// RegisterRoutes mendaftarkan semua endpoint operator ke Echo.
// authMiddleware dipakai untuk owner routes (butuh JWT valid).
// Operator auth routes (login & login/qr) tidak butuh JWT — public.
func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	// Owner-facing — butuh JWT owner.
	owner := e.Group("/business/operators", authMiddleware, appmiddleware.RequireOwner)
	owner.POST("", h.CreateOperator)
	owner.GET("", h.GetOperators)
	owner.GET("/:operator_id", h.GetOperatorByID)
	owner.PUT("/:operator_id", h.UpdateOperator)
	owner.DELETE("/:operator_id", h.DeleteOperator)
	owner.POST("/:operator_id/regenerate-qr", h.RegenerateQR)

	// Operator auth — public (no JWT required).
	kasir := e.Group("/kasir/auth")
	kasir.POST("/login", h.LoginWithCredentials)
	kasir.POST("/login/qr", h.LoginWithQR)

	// Operator auth — protected (valid operator JWT required).
	kasirProtected := e.Group("/kasir/auth", authMiddleware, appmiddleware.RequireOperatorOnly)
	kasirProtected.POST("/logout", h.Logout)
}
