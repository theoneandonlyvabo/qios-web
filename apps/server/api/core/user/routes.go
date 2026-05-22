// domain/user/routes.go

package user

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/middleware"
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

	// Operator management dipindah ke domain/operator (path tetap /business/operators).
}
