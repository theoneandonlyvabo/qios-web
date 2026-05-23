// core/product/routes.go

package product

import (
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/products", authMiddleware)

	// Owner + operator dapat list dan baca detail
	g.GET("", h.List, appmiddleware.RequireOperator)
	g.GET("/:product_id", h.GetByID, appmiddleware.RequireOperator)

	// Hanya owner yang bisa mutasi katalog
	g.POST("", h.Create, appmiddleware.RequireOwner)
	g.PATCH("/:product_id", h.Update, appmiddleware.RequireOwner)
	g.DELETE("/:product_id", h.Delete, appmiddleware.RequireOwner)
}
