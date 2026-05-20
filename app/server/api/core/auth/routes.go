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
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout)

	// Google OAuth — post-MVP
	auth.POST("/google/login", h.GoogleLogin)

	// Operator (kasir) login dipindah ke domain/operator → /kasir/auth/login
}
