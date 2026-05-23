// core/product/handler.go
//
// Layer HTTP untuk domain product.
// Handler hanya parsing input, manggil service, dan terjemahkan error ke response.
// Tidak ada query DB — semua di service / repository.
//
// Owner routes (butuh JWT owner):
//   POST   /products              → Create
//   PATCH  /products/:product_id  → Update
//   DELETE /products/:product_id  → Delete
//
// Owner + operator routes:
//   GET    /products              → List
//   GET    /products/:product_id  → GetByID

package product

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ----------------------------------------------------------------
// Helper context readers
// ----------------------------------------------------------------

func businessIDFromCtx(c echo.Context) (uuid.UUID, error) {
	raw, _ := c.Get("business_id").(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid business_id in token")
	}
	return id, nil
}

func productIDParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		return uuid.Nil, errors.New("product_id tidak valid")
	}
	return id, nil
}

func mapServiceError(c echo.Context, err error) error {
	if errors.Is(err, ErrNotFound) {
		return response.NotFound(c)
	}
	return response.Internal(c)
}

// ----------------------------------------------------------------
// Handlers
// ----------------------------------------------------------------

// GET /products
func (h *Handler) List(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	products, err := h.service.List(c.Request().Context(), businessID)
	if err != nil {
		return response.Internal(c)
	}

	if products == nil {
		products = []*Product{}
	}
	return response.OK(c, map[string]any{"products": products})
}

// POST /products
func (h *Handler) Create(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	p, err := h.service.Create(c.Request().Context(), businessID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, p)
}

// GET /products/:product_id
func (h *Handler) GetByID(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	productID, err := productIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	p, err := h.service.GetByID(c.Request().Context(), businessID, productID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, p)
}

// PATCH /products/:product_id
func (h *Handler) Update(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	productID, err := productIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var req UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	p, err := h.service.Update(c.Request().Context(), businessID, productID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, p)
}

// DELETE /products/:product_id
func (h *Handler) Delete(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	productID, err := productIDParam(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.service.Delete(c.Request().Context(), businessID, productID); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}
