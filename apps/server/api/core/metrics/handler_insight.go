// core/metrics/handler_insight.go
//
// Insight handler — rule-based MVP implementation.
//
// This handler currently queries the DB directly to generate rule-based
// insights. Post-MVP, when the AI service at apps/server/ai/ is live,
// this handler will be refactored to fetch from an internal AI service
// endpoint instead of querying the DB. The response schema
// (model_version, confidence_score) is already AI-ready.
//
// At that point, this handler may be moved out of metrics/ into its own
// domain or kept as a thin proxy — to be decided based on AI service shape.
//
// Endpoint: GET /metrics/insight

package metrics

import (
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/response"
)

// GET /metrics/insight
func (h *Handler) Insight(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	cards, err := h.q.Generate(c.Request().Context(), businessID)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, cards)
}
