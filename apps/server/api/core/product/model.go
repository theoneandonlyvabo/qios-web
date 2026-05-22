// core/product/model.go
//
// Tipe-tipe untuk domain product:
//   - Product              → representasi baris di tabel products
//   - CreateProductRequest → input owner saat tambah produk baru
//   - UpdateProductRequest → input owner saat update (partial, pointer fields)

package product

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("product not found")

// Product merepresentasikan baris di tabel products.
type Product struct {
	ID          uuid.UUID `json:"id"`
	BusinessID  uuid.UUID `json:"business_id"`
	Name        string    `json:"name"`
	Price       int64     `json:"price"`
	Category    *string   `json:"category"`
	Description *string   `json:"description"`
	IsAvailable bool      `json:"is_available"`
	TotalSold   int       `json:"total_sold"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest — body POST /products.
type CreateProductRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=255"`
	Price       int64   `json:"price"       validate:"min=0"`
	Category    *string `json:"category"    validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	IsAvailable *bool   `json:"is_available"`
}

// UpdateProductRequest — body PATCH /products/:product_id.
// Pointer fields supaya partial update — field yang tidak dikirim tidak berubah.
type UpdateProductRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
	Price       *int64  `json:"price"       validate:"omitempty,min=0"`
	Category    *string `json:"category"    validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	IsAvailable *bool   `json:"is_available"`
}
