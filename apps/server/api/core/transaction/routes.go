// core/transaction/routes.go

package transaction

import (
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/transactions", authMiddleware)

	// Owner-only: list semua transaksi dengan filter + pagination
	g.GET("", h.List, appmiddleware.RequireOwner)

	// Owner + operator: baca detail transaksi
	g.GET("/:transaction_id", h.GetByID, appmiddleware.RequireOperator)
}
