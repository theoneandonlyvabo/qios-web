// core/insight/routes.go

package insight

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, authMiddleware echo.MiddlewareFunc) {
	h := NewHandler(NewQueries(db))
	g := e.Group("/insight", authMiddleware, appmiddleware.RequireOwner)
	g.GET("", h.List)
}
