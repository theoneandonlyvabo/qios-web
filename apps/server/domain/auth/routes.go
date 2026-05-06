// domain/auth/routes.go
//
// Routing untuk domain auth.
// Semua endpoint auth tidak memerlukan middleware JWT (public).

package auth

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, cfg *config.Config, jwtSvc *jwt.Service) {
	auth := e.Group("/auth")

	// Owner
	auth.POST("/login", login(db, cfg, jwtSvc))
	auth.POST("/refresh", refresh(db, jwtSvc))
	auth.POST("/logout", logout(db))

	// Google OAuth — post-MVP
	auth.POST("/google/login", googleLogin(db, cfg, jwtSvc))

	// Operator (kasir)
	auth.POST("/operator/login", operatorLogin(db, cfg, jwtSvc))
}
