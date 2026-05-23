// core/user/handler.go
//
// Layer HTTP untuk domain user.
// Handler hanya parsing input, manggil service, dan terjemahkan error ke response.
//
// Owner profile:
//   GET   /users/me     → GetMe
//   PATCH /users/me     → UpdateMe
//
// Business info:
//   GET   /business     → GetBusiness
//   PATCH /business     → UpdateBusiness
//
// Operator CRUD (owner-only):
//   POST   /business/operators              → CreateOperator
//   GET    /business/operators              → GetOperators
//   GET    /business/operators/:operator_id → GetOperatorByID
//   PUT    /business/operators/:operator_id → UpdateOperator
//   DELETE /business/operators/:operator_id → DeleteOperator
//   POST   /business/operators/:operator_id/regenerate-qr → RegenerateQR

package user

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/response"
)

// Handler wraps Service untuk semua endpoint domain user.
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

func userIDFromCtx(c echo.Context) string {
	id, _ := c.Get("user_id").(string)
	return id
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
// User profile handlers
// ----------------------------------------------------------------

// GET /users/me
func (h *Handler) GetMe(c echo.Context) error {
	userID := userIDFromCtx(c)
	role, _ := c.Get("role").(string)

	profile, err := h.service.GetMe(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NotFound(c)
		}
		return response.Internal(c)
	}

	return response.OK(c, meResponse{
		ID:       profile.ID,
		Email:    profile.Email,
		FullName: profile.FullName,
		Phone:    profile.Phone,
		Role:     role,
		Business: businessInMe{
			ID:           profile.BusinessID,
			QiosID:       profile.QiosID,
			BusinessName: profile.BusinessName,
			Phone:        profile.BizPhone,
			Address:      profile.Address,
			City:         profile.City,
			Country:      profile.Country,
			XenditStatus: profile.XenditStatus,
		},
	})
}

// PATCH /users/me
func (h *Handler) UpdateMe(c echo.Context) error {
	userID := userIDFromCtx(c)

	var req struct {
		FullName string  `json:"full_name" validate:"omitempty,min=1,max=255"`
		Phone    *string `json:"phone"     validate:"omitempty,min=1,max=32"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.UpdateMe(c.Request().Context(), userID, req.FullName, req.Phone); err != nil {
		return response.Internal(c)
	}
	return response.NoContent(c)
}

// ----------------------------------------------------------------
// Business handlers
// ----------------------------------------------------------------

// GET /business
func (h *Handler) GetBusiness(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	biz, err := h.service.GetBusiness(c.Request().Context(), businessID.String())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NotFound(c)
		}
		return response.Internal(c)
	}

	return response.OK(c, businessResponse{
		ID:           biz.ID,
		QiosID:       biz.QiosID,
		BusinessName: biz.BusinessName,
		Phone:        biz.Phone,
		Address:      biz.Address,
		City:         biz.City,
		Country:      biz.Country,
		XenditStatus: biz.XenditStatus,
		QrisString:   biz.QrisString,
	})
}

// PATCH /business
func (h *Handler) UpdateBusiness(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
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

	if err := h.service.UpdateBusiness(c.Request().Context(), businessID.String(), req); err != nil {
		return response.Internal(c)
	}
	return response.NoContent(c)
}

// ----------------------------------------------------------------
// Operator CRUD handlers
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
	default:
		return response.Internal(c)
	}
}
