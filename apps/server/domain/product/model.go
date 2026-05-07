package product

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound dipakai service dan handler untuk distinguish 404 dari 500.
var ErrNotFound = errors.New("product not found")

// Product adalah representasi baris di tabel products.
type Product struct {
	ID          uuid.UUID `json:"id"`
	BusinessID  uuid.UUID `json:"business_id"`
	Name        string    `json:"name"`
	Price       int64     `json:"price"`
	Category    string    `json:"category,omitempty"`
	Description string    `json:"description,omitempty"`
	IsAvailable bool      `json:"is_available"`
	TotalSold   int       `json:"total_sold"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FilterParams untuk query GET /products.
type FilterParams struct {
	Category string
	Query    string
	Sort     string // "name" | "popular" | "created_at"
}

// CreateInput untuk POST /products.
type CreateInput struct {
	Name        string `json:"name"`
	Price       int64  `json:"price"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// UpdateInput untuk PATCH /products/:id.
// Pointer dipakai supaya bisa distinguish antara field yang tidak dikirim vs dikirim kosong.
type UpdateInput struct {
	Name        *string `json:"name"`
	Price       *int64  `json:"price"`
	Category    *string `json:"category"`
	Description *string `json:"description"`
	IsAvailable *bool   `json:"is_available"`
}
