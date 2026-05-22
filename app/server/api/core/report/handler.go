// core/report/handler.go

package report

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/response"
)

type Handler struct {
	q *Queries
}

func NewHandler(q *Queries) *Handler {
	return &Handler{q: q}
}

func businessIDFromCtx(c echo.Context) (uuid.UUID, error) {
	raw, _ := c.Get("business_id").(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("business_id tidak valid di token")
	}
	return id, nil
}

// GET /reports/daily-sales?date=YYYY-MM-DD
func (h *Handler) DailySales(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	dateStr := c.QueryParam("date")
	if dateStr == "" {
		return response.BadRequest(c, "date wajib diisi (format YYYY-MM-DD)")
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return response.BadRequest(c, "date harus format YYYY-MM-DD")
	}

	rep, err := h.q.DailySales(c.Request().Context(), businessID, date)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, rep)
}

// GET /reports/monthly-sales?month=YYYY-MM
func (h *Handler) MonthlySales(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return response.BadRequest(c, "month wajib diisi (format YYYY-MM)")
	}
	t, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return response.BadRequest(c, "month harus format YYYY-MM")
	}

	rep, err := h.q.MonthlySales(c.Request().Context(), businessID, t.Year(), int(t.Month()))
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, rep)
}

// GET /reports/consumption?start_date=&end_date=&ingredient=
// MVP: returns empty entries — consumption_log belum diimplementasi.
func (h *Handler) Consumption(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	startStr := c.QueryParam("start_date")
	endStr := c.QueryParam("end_date")
	if startStr == "" || endStr == "" {
		return response.BadRequest(c, "start_date dan end_date wajib diisi")
	}
	if _, err := time.Parse("2006-01-02", startStr); err != nil {
		return response.BadRequest(c, "start_date harus format YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", endStr); err != nil {
		return response.BadRequest(c, "end_date harus format YYYY-MM-DD")
	}

	rep := &ConsumptionReport{
		StartDate:  startStr,
		EndDate:    endStr,
		BusinessID: businessID.String(),
		Entries:    []*ConsumptionEntry{},
	}
	return response.OK(c, rep)
}

// POST /reports/export
// Streams CSV untuk daily_sales atau monthly_sales. PDF belum didukung.
func (h *Handler) Export(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req struct {
		Type   string         `json:"type"`
		Format string         `json:"format"`
		Period map[string]any `json:"period"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if req.Format == "pdf" {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{
			"success": false,
			"error":   "format pdf belum didukung, gunakan csv",
		})
	}
	if req.Type == "consumption" {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{
			"success": false,
			"error":   "export consumption belum didukung",
		})
	}

	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")

	switch req.Type {
	case "daily_sales":
		dateStr, _ := req.Period["date"].(string)
		if dateStr == "" {
			return response.BadRequest(c, "period.date wajib untuk daily_sales")
		}
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return response.BadRequest(c, "period.date harus format YYYY-MM-DD")
		}
		rep, err := h.q.DailySales(c.Request().Context(), businessID, date)
		if err != nil {
			return response.Internal(c)
		}
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="daily-sales-%s.csv"`, dateStr))
		w := csv.NewWriter(c.Response())
		_ = w.Write([]string{"product_name", "quantity_sold", "revenue"})
		for _, p := range rep.BreakdownByProduct {
			_ = w.Write([]string{p.ProductName, strconv.Itoa(p.QuantitySold), strconv.FormatInt(p.Revenue, 10)})
		}
		w.Flush()
		return nil

	case "monthly_sales":
		monthStr, _ := req.Period["month"].(string)
		if monthStr == "" {
			return response.BadRequest(c, "period.month wajib untuk monthly_sales")
		}
		t, err := time.Parse("2006-01", monthStr)
		if err != nil {
			return response.BadRequest(c, "period.month harus format YYYY-MM")
		}
		rep, err := h.q.MonthlySales(c.Request().Context(), businessID, t.Year(), int(t.Month()))
		if err != nil {
			return response.Internal(c)
		}
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="monthly-sales-%s.csv"`, monthStr))
		w := csv.NewWriter(c.Response())
		_ = w.Write([]string{"date", "transaction_count", "revenue"})
		for _, d := range rep.DailyBreakdown {
			_ = w.Write([]string{d.Date, strconv.Itoa(d.TransactionCount), strconv.FormatInt(d.Revenue, 10)})
		}
		w.Flush()
		return nil

	default:
		return response.BadRequest(c, "type tidak valid: gunakan daily_sales atau monthly_sales")
	}
}
