// core/admin/model.go
//
// Tipe-tipe untuk domain admin:
//   - OwnerSummary / OwnerDetail  → owner (user+business) untuk admin view
//   - AdminProduct / AdminProductDetail → product untuk admin view
//   - AdminTransaction             → order untuk admin view
//   - Request/Response types untuk semua endpoint admin

package admin

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOwnerNotFound         = errors.New("owner not found")
	ErrBusinessNotFound      = errors.New("business not found")
	ErrProductNotFound       = errors.New("product not found")
	ErrOperatorNotFound      = errors.New("operator not found")
	ErrEmailTaken            = errors.New("email already registered")
	ErrTransactionNotFound   = errors.New("transaction not found")
	ErrTransactionNotPending = errors.New("transaction is not pending")
)

// ----------------------------------------------------------------
// Owner
// ----------------------------------------------------------------

// OwnerSummary — owner row untuk list (JOIN users+businesses).
type OwnerSummary struct {
	BusinessID     uuid.UUID `json:"business_id"`
	UserID         uuid.UUID `json:"user_id"`
	QiosID         string    `json:"qios_id"`
	Email          string    `json:"email"`
	FullName       string    `json:"full_name"`
	BusinessName   string    `json:"business_name"`
	MerchantStatus string    `json:"merchant_status"`
	IsActive       bool      `json:"is_active"`
	IsSuspended    bool      `json:"is_suspended"`
	CreatedAt      time.Time `json:"created_at"`
}

// OwnerDetail — detail owner termasuk field bisnis dan user.
type OwnerDetail struct {
	BusinessID     uuid.UUID `json:"business_id"`
	UserID         uuid.UUID `json:"user_id"`
	QiosID         string    `json:"qios_id"`
	Email          string    `json:"email"`
	FullName       string    `json:"full_name"`
	BusinessName   string    `json:"business_name"`
	Phone          *string   `json:"phone"`
	Address        *string   `json:"address"`
	City           *string   `json:"city"`
	Country        *string   `json:"country"`
	QrisString     *string   `json:"qris_string"`
	MerchantStatus string    `json:"merchant_status"`
	IsActive       bool      `json:"is_active"`
	IsSuspended    bool      `json:"is_suspended"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// OwnerListResult — response paginated GET /admin/owners.
type OwnerListResult struct {
	Owners []*OwnerSummary `json:"owners"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// SetOwnerStatusRequest — body PATCH /admin/owners/:owner_id/status.
type SetOwnerStatusRequest struct {
	Enabled bool `json:"enabled"` // true=aktif, false=suspend
}

// SetOwnerCredentialRequest — body POST /admin/owners/:owner_id/credential.
type SetOwnerCredentialRequest struct {
	Email    *string `json:"email"    validate:"omitempty,email"`
	Password string  `json:"password" validate:"required,min=6"`
}

// Business — representasi businesses row untuk admin view (create/update).
type Business struct {
	ID             uuid.UUID `json:"id"`
	QiosID         string    `json:"qios_id"`
	UserID         uuid.UUID `json:"user_id"`
	BusinessName   string    `json:"business_name"`
	Phone          *string   `json:"phone"`
	Address        *string   `json:"address"`
	City           *string   `json:"city"`
	Country        *string   `json:"country"`
	MerchantStatus string    `json:"merchant_status"`
	QrisString     *string   `json:"qris_string"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateBusinessRequest — body POST /admin/owners.
// Membuat user (owner) + business secara atomik.
type CreateBusinessRequest struct {
	Email        string  `json:"email"         validate:"required,email"`
	Password     string  `json:"password"      validate:"required,min=6"`
	FullName     string  `json:"full_name"     validate:"required,min=1,max=255"`
	BusinessName string  `json:"business_name" validate:"required,min=1,max=255"`
	Phone        *string `json:"phone"         validate:"omitempty,max=32"`
	Address      *string `json:"address"`
	City         *string `json:"city"          validate:"omitempty,max=100"`
	Country      *string `json:"country"       validate:"omitempty,max=100"`
}

// UpdateBusinessRequest — body PATCH /admin/owners/:owner_id.
type UpdateBusinessRequest struct {
	BusinessName   *string `json:"business_name"   validate:"omitempty,min=1,max=255"`
	Phone          *string `json:"phone"           validate:"omitempty,max=32"`
	Address        *string `json:"address"`
	City           *string `json:"city"            validate:"omitempty,max=100"`
	Country        *string `json:"country"         validate:"omitempty,max=100"`
	MerchantStatus *string `json:"merchant_status" validate:"omitempty,oneof=PENDING REGISTERED ACTIVE SUSPENDED"`
}

// ----------------------------------------------------------------
// Product
// ----------------------------------------------------------------

// AdminProduct — product row untuk admin (tanpa recipe).
type AdminProduct struct {
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

// Ingredient — satu bahan baku dalam recipe produk (ADR-005).
type Ingredient struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"qty"`
	Unit     string  `json:"unit"`
}

// AdminProductDetail — product row termasuk recipe JSONB.
type AdminProductDetail struct {
	ID          uuid.UUID    `json:"id"`
	BusinessID  uuid.UUID    `json:"business_id"`
	Name        string       `json:"name"`
	Price       int64        `json:"price"`
	Category    *string      `json:"category"`
	Description *string      `json:"description"`
	IsAvailable bool         `json:"is_available"`
	Recipe      []Ingredient `json:"recipe"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// AdminCreateProductRequest — body POST /admin/owners/:owner_id/products.
type AdminCreateProductRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=255"`
	Price       int64   `json:"price"       validate:"min=0"`
	Category    *string `json:"category"    validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	IsAvailable *bool   `json:"is_available"`
}

// AdminUpdateProductRequest — body PATCH /admin/products/:product_id.
type AdminUpdateProductRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
	Price       *int64  `json:"price"       validate:"omitempty,min=0"`
	Category    *string `json:"category"    validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	IsAvailable *bool   `json:"is_available"`
}

// UpdateRecipeRequest — body PUT /admin/products/:product_id/recipe.
type UpdateRecipeRequest struct {
	Recipe []Ingredient `json:"recipe" validate:"required"`
}

// ----------------------------------------------------------------
// Transaction
// ----------------------------------------------------------------

// AdminTransaction — transaction row untuk admin view.
type AdminTransaction struct {
	ID            uuid.UUID  `json:"id"`
	BusinessID    uuid.UUID  `json:"business_id"`
	OperatorID    *uuid.UUID `json:"operator_id"`
	OrderID       string     `json:"order_id"`
	TotalAmount   int64      `json:"total_amount"`
	PaymentMethod *string    `json:"payment_method"`
	Status        string     `json:"status"`
	Note          *string    `json:"note"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// AdminListTransactionsFilter — filter untuk GET /admin/transactions.
type AdminListTransactionsFilter struct {
	BusinessID *uuid.UUID
	Status     string
	Page       int
	Limit      int
}

// AdminListTransactionsResult — response GET /admin/transactions.
type AdminListTransactionsResult struct {
	Transactions []*AdminTransaction `json:"transactions"`
	Total        int                 `json:"total"`
	Page         int                 `json:"page"`
	Limit        int                 `json:"limit"`
}
