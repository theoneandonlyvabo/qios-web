// domain/auth/register.go
//
// Endpoint POST /auth/register.
// Handler hanya parse + validate + call service.Register + map error.
// Seluruh business logic (transaksi, Xendit, password hashing) ada di service.

package auth

import (
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
)

// registerRequest adalah body POST /auth/register.
type registerRequest struct {
	Email        string `json:"email"         validate:"required,email,max=255"`
	Password     string `json:"password"      validate:"required,min=8,max=72"`
	FullName     string `json:"full_name"     validate:"required,min=1,max=255"`
	Phone        string `json:"phone"         validate:"required,min=4,max=32"`
	BusinessName string `json:"business_name" validate:"required,min=1,max=255"`
	Address      string `json:"address"       validate:"required,min=1,max=1024"`
	City         string `json:"city"          validate:"required,min=1,max=100"`
	Country      string `json:"country"       validate:"required,min=2,max=100"`
}

// registerResponse adalah body 201 Created.
// Field qm_id dipertahankan untuk backward compat di JSON.
type registerResponse struct {
	AccessToken  string `json:"access_token"`
	UserID       string `json:"user_id"`
	BusinessID   string `json:"business_id"`
	QMID         string `json:"qm_id"`
	XenditStatus string `json:"xendit_status"`
}

// Register — owner onboarding (user + business + Xendit sub-account).
// POST /auth/register
func (h *Handler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	out, err := h.service.Register(c.Request().Context(), RegisterInput{
		Email:        req.Email,
		Password:     req.Password,
		FullName:     req.FullName,
		Phone:        req.Phone,
		BusinessName: req.BusinessName,
		Address:      req.Address,
		City:         req.City,
		Country:      req.Country,
	})
	if err != nil {
		return mapServiceError(c, err)
	}

	setRefreshCookie(c, out.RefreshToken, out.RefreshExpiry)
	return response.Created(c, registerResponse{
		AccessToken:  out.AccessToken,
		UserID:       out.UserID,
		BusinessID:   out.BusinessID,
		QMID:         out.QiosID,
		XenditStatus: out.XenditStatus,
	})
}
