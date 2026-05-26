// core/metrics/handler_summary.go
//
// Handler untuk Summary, Trend, TopProducts, PeakHours (dari dashboard).
// Endpoint: GET /metrics/summary, /metrics/trend, /metrics/top-products, /metrics/peak-hours

package metrics

import (
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/response"
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

// GET /metrics/summary?period=today|this_week|this_month|last_month
func (h *Handler) Summary(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	period := c.QueryParam("period")
	if period == "" {
		period = "this_month"
	}

	start, end := periodRange(period)
	s, err := h.q.Summary(c.Request().Context(), businessID, start, end)
	if err != nil {
		return response.Internal(c)
	}
	s.Period = period
	return response.OK(c, s)
}

// GET /metrics/trend?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
// Default: last 7 days.
func (h *Handler) Trend(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	end := time.Now()
	start := end.AddDate(0, 0, -6)

	if s := c.QueryParam("start_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.BadRequest(c, "start_date harus format YYYY-MM-DD")
		}
		start = t
	}
	if s := c.QueryParam("end_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.BadRequest(c, "end_date harus format YYYY-MM-DD")
		}
		end = t
	}

	trend, err := h.q.Trend(c.Request().Context(), businessID, start, end)
	if err != nil {
		return response.Internal(c)
	}
	if trend == nil {
		trend = []*DailyTrend{}
	}
	return response.OK(c, trend)
}

// GET /metrics/top-products?period=&limit=10
func (h *Handler) TopProducts(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	period := c.QueryParam("period")
	if period == "" {
		period = "this_month"
	}

	limit := 10
	if s := c.QueryParam("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 50 {
			return response.BadRequest(c, "limit harus antara 1–50")
		}
		limit = n
	}

	start, end := periodRange(period)
	products, err := h.q.TopProducts(c.Request().Context(), businessID, start, end, limit)
	if err != nil {
		return response.Internal(c)
	}
	if products == nil {
		products = []*TopProduct{}
	}
	return response.OK(c, products)
}

// GET /metrics/peak-hours?period=this_week|this_month|last_month
func (h *Handler) PeakHours(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	period := c.QueryParam("period")
	if period == "" {
		period = "this_month"
	}

	start, end := periodRange(period)
	hours, err := h.q.PeakHours(c.Request().Context(), businessID, start, end)
	if err != nil {
		return response.Internal(c)
	}
	if hours == nil {
		hours = []*HourlyDistribution{}
	}
	return response.OK(c, hours)
}
