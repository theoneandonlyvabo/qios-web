// domain/auth/routes.go
//
// Routing untuk domain auth.
// Semua endpoint auth tidak memerlukan middleware JWT (public).

package auth

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes mendaftarkan semua endpoint domain auth ke Echo.
func RegisterRoutes(e *echo.Echo, h *Handler) {
	auth := e.Group("/auth")

	// Owner
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout)

	// Google OAuth — post-MVP
	auth.POST("/google/login", h.GoogleLogin)

	// Operator (kasir) login — public, no JWT required
	kasir := e.Group("/kasir/auth")
	kasir.POST("/login", h.OperatorLoginWithCredentials)
	kasir.POST("/login/qr", h.OperatorLoginWithQR)
}
