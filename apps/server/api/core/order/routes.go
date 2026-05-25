// core/order/routes.go

package order

import (
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	orders := e.Group("/orders", authMiddleware, appmiddleware.RateLimitOperator)
	orders.POST("", h.CreateOrder, appmiddleware.RequireOperator)
	orders.GET("", h.ListMyOrders, appmiddleware.RequireOperator)
	orders.PATCH("/:order_id/items", h.UpdateItems, appmiddleware.RequireOperator)
	orders.POST("/:order_id/checkout/begin", h.BeginCheckout, appmiddleware.RequireOperator)
	orders.POST("/:order_id/checkout/confirm", h.ConfirmCheckout, appmiddleware.RequireOperator)
	orders.POST("/:order_id/void", h.VoidOrder, appmiddleware.RequireOperator)

	sessions := e.Group("/orders/sessions", authMiddleware, appmiddleware.RateLimitOwner)
	sessions.GET("", h.ListActiveSessions, appmiddleware.RequireOwner)
	sessions.DELETE("/:session_id", h.ForceEndSession, appmiddleware.RequireOwner)
}
