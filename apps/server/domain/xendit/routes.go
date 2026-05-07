// domain/xendit/routes.go

package xendit

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, cfg *config.Config, authMiddleware echo.MiddlewareFunc) {
	m := e.Group("/payment/midtrans", authMiddleware)
	m.POST("/connect", connectMidtrans(db, cfg), appmiddleware.RequireOwner)
	m.GET("/status", getMidtransStatus(db))

	// Webhook tidak pakai Bearer auth — verifikasi via Midtrans signature key
	e.POST("/payment/midtrans/webhook", handleWebhook(db, cfg))
}
