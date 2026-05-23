// core/report/routes.go

package report

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, authMiddleware echo.MiddlewareFunc) {
	h := NewHandler(NewQueries(db))
	g := e.Group("/reports", authMiddleware, appmiddleware.RequireOwner)
	g.GET("/daily-sales", h.DailySales)
	g.GET("/monthly-sales", h.MonthlySales)
	g.GET("/consumption", h.Consumption)
	g.POST("/export", h.Export)
}
