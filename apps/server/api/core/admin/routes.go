// core/admin/routes.go

package admin

import (
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/middleware"
)

var adminLoginRateLimiter = echomiddleware.RateLimiter(
	echomiddleware.NewRateLimiterMemoryStore(10),
)

func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	// Public auth routes — tidak butuh JWT
	e.POST("/admin/auth/login", h.Login, adminLoginRateLimiter)
	e.POST("/admin/auth/refresh", h.Refresh)
	e.POST("/admin/auth/logout", h.Logout)

	// Protected routes — butuh JWT role=admin
	g := e.Group("/admin", authMiddleware, appmiddleware.RequireAdmin, appmiddleware.RateLimitAdmin)

	g.GET("/me", h.Me)

	// Business management
	g.GET("/businesses", h.ListBusinesses)
	g.POST("/businesses", h.CreateBusiness)
	g.GET("/businesses/:business_id", h.GetBusiness)
	g.PATCH("/businesses/:business_id", h.UpdateBusiness)

	// Product management
	g.GET("/businesses/:business_id/products", h.ListProducts)
	g.POST("/businesses/:business_id/products", h.CreateProduct)
	g.PATCH("/products/:product_id", h.UpdateProduct)
	g.DELETE("/products/:product_id", h.DeleteProduct)

	// Operator management
	g.DELETE("/businesses/:business_id/operators/:operator_id", h.DeleteOperator)

	// Transaction management
	g.GET("/transactions", h.ListTransactions)
	g.POST("/transactions/:transaction_id/void", h.VoidTransaction)
}
