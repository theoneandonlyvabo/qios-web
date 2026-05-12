// domain/dashboard/handler.go
//
// Placeholder handlers untuk dashboard domain.
// Akan diimplementasikan setelah payment domain stable — semua endpoint disini
// agregasi dari pos_orders, jadi harus tunggu transaction flow live.
//
// Endpoints:
//   GET /dashboard/summary                    → ringkasan performa bisnis
//   GET /dashboard/transactions/trend         → tren transaksi harian
//   GET /dashboard/transactions/peak-hours    → distribusi transaksi per jam
//   GET /dashboard/products/top               → produk terlaris

package dashboard

import (
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
)

// Handler adalah HTTP handler untuk dashboard domain.
// Belum punya Service/Repository karena semua endpoint masih stub.
// Saat implementasi nanti, ikuti pola operator/product: tambah service.go + repository.go.
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// GET /dashboard/summary
func (h *Handler) GetSummary(c echo.Context) error {
	return response.NotImplemented(c, "dashboard summary not yet implemented")
}

// GET /dashboard/transactions/trend
func (h *Handler) GetTransactionTrend(c echo.Context) error {
	return response.NotImplemented(c, "transaction trend not yet implemented")
}

// GET /dashboard/transactions/peak-hours
func (h *Handler) GetPeakHours(c echo.Context) error {
	return response.NotImplemented(c, "peak hours not yet implemented")
}

// GET /dashboard/products/top
func (h *Handler) GetTopProducts(c echo.Context) error {
	return response.NotImplemented(c, "top products not yet implemented")
}

func RegisterRoutes(e *echo.Echo, h *Handler, authMw echo.MiddlewareFunc) {
	d := e.Group("/dashboard", authMw, appmiddleware.RequireOwner)
	d.GET("/summary", h.GetSummary)
	d.GET("/transactions/trend", h.GetTransactionTrend)
	d.GET("/transactions/peak-hours", h.GetPeakHours)
	d.GET("/products/top", h.GetTopProducts)
}
