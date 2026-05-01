// domain/auth/routes.go
//
// Register semua route auth ke Echo instance.
// Handler belum diimplementasi — skeleton untuk memastikan server bisa compile.

package auth

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, cfg *config.Config, jwtSvc *jwt.Service) {
	g := e.Group("/auth")
	g.POST("/login", login(db, cfg, jwtSvc))
	g.POST("/google", googleLogin(db, cfg, jwtSvc))
	g.POST("/refresh", refresh(db, jwtSvc))
	g.POST("/logout", logout(db))
}
