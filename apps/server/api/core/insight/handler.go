// core/insight/handler.go

package insight

import (
	"errors"

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

// GET /insight
func (h *Handler) List(c echo.Context) error {
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
