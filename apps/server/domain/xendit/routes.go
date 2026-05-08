// domain/xendit/routes.go
//
// Routing untuk endpoint Xendit (xenPlatform).
// Webhook tidak pakai Bearer auth — diverifikasi via Xendit signature di handler.

package xendit

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, cfg *config.Config, authMiddleware echo.MiddlewareFunc) {
	x := e.Group("/payment/xendit", authMiddleware)
	x.POST("/connect", connectXendit(db, cfg), appmiddleware.RequireOwner)
	x.GET("/status", getXenditStatus(db))

	// Webhook tidak pakai Bearer auth — verifikasi via Xendit signature key di handler.
	e.POST("/payment/xendit/webhook", handleWebhook(db, cfg))
}
