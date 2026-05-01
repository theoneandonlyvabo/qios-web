// domain/user/routes.go

package user

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/users", authMiddleware)
	g.GET("/me", getMe(db))
	g.PATCH("/me", updateMe(db), appmiddleware.RequireOwner)

	b := e.Group("/business", authMiddleware)
	b.GET("", getBusiness(db))
	b.PATCH("", updateBusiness(db), appmiddleware.RequireOwner)

	o := b.Group("/operators", appmiddleware.RequireOwner)
	o.GET("", listOperators(db))
	o.POST("", createOperator(db))
	o.DELETE("/:operator_id", deleteOperator(db))
}
