// domain/transaction/handler.go
// TODO: implement semua handler

package transaction

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

func listProducts(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func createProduct(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func updateProduct(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func deleteProduct(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func listTransactions(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func createOrder(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func getTransaction(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func getSummary(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func getTransactionTrend(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func getTopProducts(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}

func getPeakHours(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error { return c.JSON(http.StatusNotImplemented, nil) }
}
