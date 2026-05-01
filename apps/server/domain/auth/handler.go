// domain/auth/handler.go
//
// Handler auth — belum diimplementasi.
// Setiap fungsi mengembalikan handler Echo yang valid supaya server bisa compile.
// Implementasi dilakukan per handler sesuai urutan prioritas.

package auth

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
)

func login(db *sql.DB, cfg *config.Config, jwtSvc *jwt.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: implement
		return c.JSON(http.StatusNotImplemented, nil)
	}
}

func googleLogin(db *sql.DB, cfg *config.Config, jwtSvc *jwt.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: implement
		return c.JSON(http.StatusNotImplemented, nil)
	}
}

func refresh(db *sql.DB, jwtSvc *jwt.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: implement
		return c.JSON(http.StatusNotImplemented, nil)
	}
}

func logout(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: implement
		return c.JSON(http.StatusNotImplemented, nil)
	}
}
