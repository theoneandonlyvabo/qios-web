// domain/auth/handler.go
//
// Layer HTTP untuk domain auth. Handler hanya parsing input,
// manggil service, dan menerjemahkan error ke response.
// Tidak ada SQL, tidak ada hashing — semua di service / repository.
//
// Refresh token disimpan di httpOnly cookie, bukan response body.
// Access token dikembalikan di response body, disimpan di memory client.
//
// Endpoint:
//   POST /auth/login            → Login (owner)
//   POST /auth/refresh          → Refresh
//   POST /auth/logout           → Logout
//   POST /auth/google/login     → GoogleLogin (post-MVP)
//   POST /kasir/auth/login      → OperatorLoginWithCredentials (public)
//   POST /kasir/auth/login/qr   → OperatorLoginWithQR (public)

package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/response"
)

const refreshTokenCookieName = "refresh_token"

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ----------------------------------------------------------------
// Request DTOs
// ----------------------------------------------------------------

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ----------------------------------------------------------------
// Cookie helpers
// ----------------------------------------------------------------

// setRefreshCookie menempel refresh token ke browser via httpOnly cookie.
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

// clearRefreshCookie menghapus cookie refresh token dari browser.
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
// Handlers
// ----------------------------------------------------------------

// Login — owner via email+password.
// POST /auth/login
func (h *Handler) Login(c echo.Context) error {
	var req loginRequest
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

// GoogleLogin — owner via Google OAuth.
// POST /auth/google/login
// TODO: implement post-MVP.
func (h *Handler) GoogleLogin(c echo.Context) error {
	return response.NotImplemented(c, "Google login not yet implemented")
}

// Refresh — rotate refresh token, kembalikan access token baru.
// POST /auth/refresh
// Baca refresh token dari cookie, bukan body.
func (h *Handler) Refresh(c echo.Context) error {
	cookie, err := c.Cookie(refreshTokenCookieName)
	if err != nil || cookie.Value == "" {
		return response.Unauthorized(c)
	}

	out, err := h.service.Refresh(c.Request().Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, ErrSessionExpired) {
			clearRefreshCookie(c)
		}
		return mapServiceError(c, err)
	}

	setRefreshCookie(c, out.RefreshToken, out.RefreshExpiry)
	return response.OK(c, map[string]string{"access_token": out.AccessToken})
}

// Logout — hapus refresh token dari DB dan clear cookie.
// POST /auth/logout
func (h *Handler) Logout(c echo.Context) error {
	cookie, err := c.Cookie(refreshTokenCookieName)
	if err != nil || cookie.Value == "" {
		// Sudah tidak ada cookie, anggap sudah logout.
		return response.NoContent(c)
	}

	// Service idempotent — error di DB tidak boleh menghalangi clear cookie.
	_ = h.service.Logout(c.Request().Context(), cookie.Value)
	clearRefreshCookie(c)
	return response.NoContent(c)
}

// ----------------------------------------------------------------
// Operator auth handlers (public — no JWT required)
// ----------------------------------------------------------------

type operatorCredentialsRequest struct {
	BusinessID   string `json:"business_id"   validate:"required,uuid4"`
	OperatorCode string `json:"operator_code" validate:"required"`
	Password     string `json:"password"      validate:"required"`
}

type operatorQRRequest struct {
	QRToken string `json:"qr_token" validate:"required"`
}

// OperatorLoginWithCredentials — kasir login via operator_code + password.
// POST /kasir/auth/login
func (h *Handler) OperatorLoginWithCredentials(c echo.Context) error {
	var req operatorCredentialsRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "business_id, operator_code, dan password wajib diisi")
	}
	businessID, err := uuid.Parse(req.BusinessID)
	if err != nil {
		return response.BadRequest(c, "business_id tidak valid")
	}

	out, err := h.service.OperatorLoginWithCredentials(c.Request().Context(), businessID, req.OperatorCode, req.Password)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, out)
}

// OperatorLoginWithQR — kasir login via QR token scan.
// POST /kasir/auth/login/qr
func (h *Handler) OperatorLoginWithQR(c echo.Context) error {
	var req operatorQRRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.QRToken == "" {
		return response.BadRequest(c, "qr_token wajib diisi")
	}

	out, err := h.service.OperatorLoginWithQR(c.Request().Context(), req.QRToken)
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
	case errors.Is(err, ErrInvalidCredentials),
		errors.Is(err, ErrRefreshNotFound):
		return response.Unauthorized(c)
	case errors.Is(err, ErrSessionExpired):
		return response.UnauthorizedMsg(c, "sesi telah berakhir, silakan login kembali")
	case errors.Is(err, ErrAccountInactive):
		return response.ForbiddenMsg(c, "account is inactive or suspended")
	case errors.Is(err, ErrOperatorInactive):
		return response.ForbiddenMsg(c, "Akun operator dinonaktifkan")
	case errors.Is(err, ErrGoogleOnlyAccount):
		return response.BadRequest(c, "this account uses Google login")
	case errors.Is(err, ErrEmailTaken):
		return response.Conflict(c, "email sudah terdaftar")
	case errors.Is(err, ErrQiosIDCollision):
		return response.Conflict(c, "qios_id collision, silakan coba lagi")
	default:
		return response.Internal(c)
	}
}

