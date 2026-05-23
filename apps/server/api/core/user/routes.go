// core/user/routes.go

package user

import (
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	// Owner profile
	g := e.Group("/users", authMiddleware)
	g.GET("/me", h.GetMe)
	g.PATCH("/me", h.UpdateMe, appmiddleware.RequireOwner)

	// Business info
	b := e.Group("/business", authMiddleware)
	b.GET("", h.GetBusiness)
	b.PATCH("", h.UpdateBusiness, appmiddleware.RequireOwner)

	// Operator CRUD — owner only
	ops := e.Group("/business/operators", authMiddleware, appmiddleware.RequireOwner)
	ops.POST("", h.CreateOperator)
	ops.GET("", h.GetOperators)
	ops.GET("/:operator_id", h.GetOperatorByID)
	ops.PUT("/:operator_id", h.UpdateOperator)
	ops.DELETE("/:operator_id", h.DeleteOperator)
	ops.POST("/:operator_id/regenerate-qr", h.RegenerateQR)
}
