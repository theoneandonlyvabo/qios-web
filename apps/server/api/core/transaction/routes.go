// core/transaction/routes.go

package transaction

import (
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/transactions", authMiddleware)

	// Owner-only: list semua transaksi dengan filter + pagination
	g.GET("", h.List, appmiddleware.RequireOwner)

	// Owner + operator: create, read, confirm, void
	g.POST("", h.Create, appmiddleware.RequireOperator)
	g.GET("/:transaction_id", h.GetByID, appmiddleware.RequireOperator)
	g.POST("/:transaction_id/confirm", h.Confirm, appmiddleware.RequireOperator)
	g.POST("/:transaction_id/void", h.Void, appmiddleware.RequireOperator)
}
