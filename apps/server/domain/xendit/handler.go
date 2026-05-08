// domain/xendit/handler.go
//
// Handler endpoint Xendit (xenPlatform). Sebagian besar webhook + connect/status.
// Di sprint ini hanya stub 501 — implementasi penuh menyusul setelah register flow stabil.
// Lihat docs/qios-api.yaml untuk kontrak penuh:
//   POST /payment/xendit/connect
//   GET  /payment/xendit/status
//   POST /payment/xendit/webhook

package xendit

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
)

func connectXendit(_ *sql.DB, _ *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		return response.NotImplemented(c, "connect xendit not yet implemented")
	}
}

func getXenditStatus(_ *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return response.NotImplemented(c, "xendit status not yet implemented")
	}
}

func handleWebhook(_ *sql.DB, _ *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		return response.NotImplemented(c, "xendit webhook not yet implemented")
	}
}
