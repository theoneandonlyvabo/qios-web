// core/metrics/routes.go

package metrics

import (
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/metrics", authMiddleware, appmiddleware.RequireOwner)

	g.GET("/summary", h.Summary)
	g.GET("/trend", h.Trend)
	g.GET("/top-products", h.TopProducts)
	g.GET("/peak-hours", h.PeakHours)
	g.GET("/overview", h.Overview)
	g.GET("/insight", h.Insight)

	reports := g.Group("/reports")
	reports.GET("/daily-sales", h.DailySales)
	reports.GET("/monthly-sales", h.MonthlySales)
	reports.GET("/consumption", h.Consumption)
	reports.POST("/export", h.Export)
}
