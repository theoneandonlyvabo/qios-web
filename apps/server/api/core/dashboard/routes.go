// core/dashboard/routes.go

package dashboard

import (
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/dashboard", authMiddleware, appmiddleware.RequireOwner)

	g.GET("/summary", h.Summary)
	g.GET("/transactions/trend", h.Trend)
	g.GET("/products/top", h.TopProducts)
	g.GET("/transactions/peak-hours", h.PeakHours)
}
