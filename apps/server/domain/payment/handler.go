// domain/payment/handler.go
// TODO: implement semua handler

package payment

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
)

func connectMidtrans(db *sql.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func getMidtransStatus(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func handleWebhook(db *sql.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}
