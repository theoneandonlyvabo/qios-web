// core/admin/routes.go

package admin

import "github.com/labstack/echo/v4"

func RegisterRoutes(e *echo.Echo, h *Handler, adminKeyMiddleware echo.MiddlewareFunc) {
	g := e.Group("/admin", adminKeyMiddleware)

	// Owner management
	g.GET("/owners", h.ListOwners)
	g.POST("/owners", h.CreateOwner)
	g.GET("/owners/:owner_id", h.GetOwner)
	g.PATCH("/owners/:owner_id", h.UpdateOwner)
	g.PATCH("/owners/:owner_id/status", h.SetOwnerStatus)
	g.POST("/owners/:owner_id/credential", h.SetOwnerCredential)

	// Product management
	g.GET("/owners/:owner_id/products", h.ListOwnerProducts)
	g.POST("/owners/:owner_id/products", h.CreateProduct)
	g.GET("/products/:product_id", h.GetProduct)
	g.PATCH("/products/:product_id", h.UpdateProduct)
	g.DELETE("/products/:product_id", h.DeleteProduct)
	g.PUT("/products/:product_id/recipe", h.UpdateProductRecipe)

	// Operator management
	g.DELETE("/owners/:owner_id/operators/:operator_id", h.DeleteOperator)

	// Transaction management
	g.GET("/transactions", h.ListTransactions)
	g.POST("/transactions/:transaction_id/void", h.VoidTransaction)
}
