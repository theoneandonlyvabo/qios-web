// core/user/routes.go

package user

import (
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	// Owner profile — hanya owner yang punya profil di tabel users
	g := e.Group("/users", authMiddleware, appmiddleware.RateLimitOwner)
	g.GET("/me", h.GetMe, appmiddleware.RequireOwner)
	g.PATCH("/me", h.UpdateMe, appmiddleware.RequireOwner)

	// Business info — owner dan operator bisa lihat (operator butuh ini untuk POS)
	b := e.Group("/business", authMiddleware, appmiddleware.RateLimitOwner)
	b.GET("", h.GetBusiness, appmiddleware.RequireOperator)
	b.PATCH("", h.UpdateBusiness, appmiddleware.RequireOwner)

	// Operator CRUD — owner only
	ops := e.Group("/business/operators", authMiddleware, appmiddleware.RequireOwner, appmiddleware.RateLimitOwner)
	ops.POST("", h.CreateOperator)
	ops.GET("", h.GetOperators)
	ops.GET("/:operator_id", h.GetOperatorByID)
	ops.PUT("/:operator_id", h.UpdateOperator)
	ops.DELETE("/:operator_id", h.DeleteOperator)
	ops.POST("/:operator_id/regenerate-qr", h.RegenerateQR)
}
