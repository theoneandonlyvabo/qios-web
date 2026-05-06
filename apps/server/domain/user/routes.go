// domain/user/routes.go

package user

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, authMiddleware echo.MiddlewareFunc) {
	// Owner profile
	g := e.Group("/users", authMiddleware)
	g.GET("/me", getMe(db))
	g.PATCH("/me", updateMe(db), appmiddleware.RequireOwner)

	// Business info
	b := e.Group("/business", authMiddleware)
	b.GET("", getBusiness(db))
	b.PATCH("", updateBusiness(db), appmiddleware.RequireOwner)

	// Operator management — owner only
	o := b.Group("/operators", appmiddleware.RequireOwner)
	o.GET("", listOperators(db))
	o.POST("", createOperator(db))
	o.PATCH("/:operator_id", updateOperator(db))
	o.DELETE("/:operator_id", deleteOperator(db))
	o.GET("/:operator_id/qr", getOperatorQR(db))
}
