// domain/user/handler.go
// TODO: implement semua handler

package user

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

func getMe(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func updateMe(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func getBusiness(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func updateBusiness(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func listOperators(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func createOperator(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func deleteOperator(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}
