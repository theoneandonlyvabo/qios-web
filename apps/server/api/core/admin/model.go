// core/admin/model.go
//
// Tipe-tipe untuk domain admin:
//   - AdminUser                    → row di tabel admin_users
//   - Business                     → row businesses untuk admin view
//   - AdminProduct                 → row products untuk admin view
//   - AdminTransaction             → row pos_orders untuk admin view
//   - Request/Response types untuk semua endpoint admin

package admin

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAdminNotFound         = errors.New("admin not found")
	ErrBusinessNotFound      = errors.New("business not found")
	ErrProductNotFound       = errors.New("product not found")
	ErrOperatorNotFound      = errors.New("operator not found")
	ErrEmailTaken            = errors.New("email already registered")
	ErrRefreshNotFound       = errors.New("refresh token not found")
	ErrSessionExpired        = errors.New("session expired")
	ErrTransactionNotFound   = errors.New("transaction not found")
	ErrTransactionNotPending = errors.New("transaction is not pending")
)

// AdminUser — row di tabel admin_users.
type AdminUser struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AdminResponse — shape yang dikembalikan ke client untuk /admin/me.
type AdminResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginRequest — body POST /admin/auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResult struct {
	AccessToken   string
	RefreshToken  string
	RefreshExpiry time.Duration
}

type RefreshResult struct {
	AccessToken   string
	RefreshToken  string
	RefreshExpiry time.Duration
}

// Business — representasi businesses row untuk admin view.
type Business struct {
	ID           uuid.UUID `json:"id"`
	QiosID       string    `json:"qios_id"`
	UserID       uuid.UUID `json:"user_id"`
	BusinessName string    `json:"business_name"`
	Phone        *string   `json:"phone"`
	Address      *string   `json:"address"`
	City         *string   `json:"city"`
	Country      *string   `json:"country"`
	XenditStatus string    `json:"xendit_status"`
	QrisString   *string   `json:"qris_string"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateBusinessRequest — body POST /admin/businesses.
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

// UpdateBusinessRequest — body PATCH /admin/businesses/:business_id.
type UpdateBusinessRequest struct {
	BusinessName *string `json:"business_name" validate:"omitempty,min=1,max=255"`
	Phone        *string `json:"phone"         validate:"omitempty,max=32"`
	Address      *string `json:"address"`
	City         *string `json:"city"          validate:"omitempty,max=100"`
	Country      *string `json:"country"       validate:"omitempty,max=100"`
	XenditStatus *string `json:"xendit_status" validate:"omitempty,oneof=PENDING REGISTERED ACTIVE SUSPENDED"`
}

// AdminProduct — product row untuk admin (tanpa business scoping di lookup).
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

// AdminCreateProductRequest — body POST /admin/businesses/:business_id/products.
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
