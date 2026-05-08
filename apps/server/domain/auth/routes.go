// domain/auth/routes.go
//
// Routing untuk domain auth.
// Semua endpoint auth tidak memerlukan middleware JWT (public).

package auth

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/payment"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
)

// RegisterRoutes mendaftarkan semua endpoint domain auth ke Echo.
// xenditSvc dipanggil dari handler register saat onboarding owner baru.
func RegisterRoutes(
	e *echo.Echo,
	db *sql.DB,
	cfg *config.Config,
	jwtSvc *jwt.Service,
	xenditSvc *payment.XenditService,
) {
	auth := e.Group("/auth")

	// Owner
	auth.POST("/register", register(db, cfg, jwtSvc, xenditSvc))
	auth.POST("/login", login(db, cfg, jwtSvc))
	auth.POST("/refresh", refresh(db, jwtSvc))
	auth.POST("/logout", logout(db))

	// Google OAuth — post-MVP
	auth.POST("/google/login", googleLogin(db, cfg, jwtSvc))

	// Operator (kasir) login dipindah ke domain/operator → /kasir/auth/login
}
