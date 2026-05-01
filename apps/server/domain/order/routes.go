// domain/order/routes.go

package order

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, authMiddleware echo.MiddlewareFunc) {
	// Products — owner manage, operator read
	p := e.Group("/products", authMiddleware)
	p.GET("", listProducts(db))
	p.POST("", createProduct(db), appmiddleware.RequireOwner)
	p.PATCH("/:product_id", updateProduct(db), appmiddleware.RequireOwner)
	p.DELETE("/:product_id", deleteProduct(db), appmiddleware.RequireOwner)

	// Transactions
	t := e.Group("/transactions", authMiddleware)
	t.GET("", listTransactions(db), appmiddleware.RequireOwner)
	t.POST("", createOrder(db), appmiddleware.RequireOperator)
	t.GET("/:transaction_id", getTransaction(db), appmiddleware.RequireOwner)

	// Dashboard
	d := e.Group("/dashboard", authMiddleware, appmiddleware.RequireOwner)
	d.GET("/summary", getSummary(db))
	d.GET("/transactions/trend", getTransactionTrend(db))
	d.GET("/products/top", getTopProducts(db))
	d.GET("/transactions/peak-hours", getPeakHours(db))
}
