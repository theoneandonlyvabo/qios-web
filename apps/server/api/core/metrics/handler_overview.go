// core/metrics/handler_overview.go
//
// Handler untuk Overview (dari analytics).
// Endpoint: GET /metrics/overview

package metrics

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/response"
)

// GET /metrics/overview?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD&compare_with=previous_period|previous_year|none
func (h *Handler) Overview(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	startStr := c.QueryParam("start_date")
	endStr := c.QueryParam("end_date")
	if startStr == "" || endStr == "" {
		return response.BadRequest(c, "start_date dan end_date wajib diisi")
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return response.BadRequest(c, "start_date harus format YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return response.BadRequest(c, "end_date harus format YYYY-MM-DD")
	}
	if end.Before(start) {
		return response.BadRequest(c, "end_date tidak boleh sebelum start_date")
	}

	compareWith := c.QueryParam("compare_with")
	if compareWith == "" {
		compareWith = "previous_period"
	}

	ov, err := h.q.Overview(c.Request().Context(), businessID, start, end, compareWith)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, ov)
}
