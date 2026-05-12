package product

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func businessIDFromCtx(c echo.Context) (uuid.UUID, error) {
	raw, _ := c.Get("business_id").(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid business_id in token")
	}
	return id, nil
}

// GET /products
func (h *Handler) ListProducts(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	filter := FilterParams{
		Category: c.QueryParam("category"),
		Query:    c.QueryParam("q"),
		Sort:     c.QueryParam("sort"),
	}
	if filter.Sort == "" {
		filter.Sort = "popular"
	}

	products, err := h.service.ListProducts(businessID, filter)
	if err != nil {
		return response.Internal(c)
	}
	return response.OK(c, products)
}

// POST /products
func (h *Handler) CreateProduct(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	var input CreateInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if input.Name == "" {
		return response.BadRequest(c, "Nama produk wajib diisi")
	}
	if input.Price < 0 {
		return response.BadRequest(c, "Harga tidak boleh negatif")
	}

	product, err := h.service.CreateProduct(businessID, input)
	if err != nil {
		return response.Internal(c)
	}
	return response.Created(c, product)
}

// PATCH /products/:product_id
func (h *Handler) UpdateProduct(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		return response.BadRequest(c, "Product ID tidak valid")
	}

	var input UpdateInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	product, err := h.service.UpdateProduct(productID, businessID, input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NotFoundMsg(c, "Produk tidak ditemukan")
		}
		return response.Internal(c)
	}
	return response.OK(c, product)
}

// DELETE /products/:product_id
func (h *Handler) DeleteProduct(c echo.Context) error {
	businessID, err := businessIDFromCtx(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		return response.BadRequest(c, "Product ID tidak valid")
	}

	if err := h.service.DeleteProduct(productID, businessID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NotFoundMsg(c, "Produk tidak ditemukan")
		}
		return response.Internal(c)
	}
	return response.NoContent(c)
}

func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware, ownerGuard echo.MiddlewareFunc) {
	products := e.Group("/products", authMiddleware)
	products.GET("", h.ListProducts)
	products.POST("", h.CreateProduct, ownerGuard)
	products.PATCH("/:product_id", h.UpdateProduct, ownerGuard)
	products.DELETE("/:product_id", h.DeleteProduct, ownerGuard)
}
