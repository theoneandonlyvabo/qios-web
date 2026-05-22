// core/transaction/model.go
//
// Tipe-tipe untuk domain transaction:
//   - Order           → representasi baris di tabel pos_orders
//   - OrderItem       → representasi baris di tabel pos_order_items
//   - OrderWithItems  → Order + item-itemnya untuk response detail
//   - ConfirmResponse → response confirm; menyertakan qris_string jika QRIS
//
// Catatan: payment_method di-set saat confirm, bukan saat create.
// total_amount dihitung server-side dari items — client tidak kirim total.

package transaction

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound        = errors.New("transaction not found")
	ErrNotPending      = errors.New("transaction is not pending")
	ErrEmptyItems      = errors.New("transaction must have at least one item")
	ErrProductNotFound = errors.New("one or more products not found or unavailable")
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusPaid      Status = "paid"
	StatusFailed    Status = "failed"
	StatusExpired   Status = "expired"
	StatusCancelled Status = "cancelled"
)

type PaymentMethod string

const (
	PaymentCash           PaymentMethod = "CASH"
	PaymentQRIS           PaymentMethod = "QRIS"
	PaymentEwallet        PaymentMethod = "EWALLET"
	PaymentVirtualAccount PaymentMethod = "VIRTUAL_ACCOUNT"
)

// Order merepresentasikan satu baris di pos_orders.
type Order struct {
	ID            uuid.UUID      `json:"id"`
	BusinessID    uuid.UUID      `json:"business_id"`
	OperatorID    *uuid.UUID     `json:"operator_id"`
	OrderID       string         `json:"order_id"`
	TotalAmount   int64          `json:"total_amount"`
	PaymentMethod *PaymentMethod `json:"payment_method"`
	Status        Status         `json:"status"`
	Note          *string        `json:"note"`
	PaidAt        *time.Time     `json:"paid_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// OrderItem merepresentasikan satu baris di pos_order_items.
// product_name dan unit_price adalah snapshot saat order dibuat.
type OrderItem struct {
	ID          uuid.UUID  `json:"id"`
	ProductID   *uuid.UUID `json:"product_id"`
	ProductName string     `json:"product_name"`
	UnitPrice   int64      `json:"unit_price"`
	Quantity    int        `json:"quantity"`
	Subtotal    int64      `json:"subtotal"`
}

// OrderWithItems — response untuk GET /transactions/:id.
type OrderWithItems struct {
	Order
	Items []*OrderItem `json:"items"`
}

// ----------------------------------------------------------------
// Request types
// ----------------------------------------------------------------

// ItemInput — satu baris item pada POST /transactions.
type ItemInput struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity"   validate:"required,min=1"`
}

// CreateOrderRequest — body POST /transactions.
type CreateOrderRequest struct {
	Items []ItemInput `json:"items" validate:"required,min=1,dive"`
	Note  *string     `json:"note"  validate:"omitempty,max=500"`
}

// ConfirmOrderRequest — body POST /transactions/:id/confirm.
type ConfirmOrderRequest struct {
	PaymentMethod PaymentMethod `json:"payment_method" validate:"required,oneof=CASH QRIS EWALLET VIRTUAL_ACCOUNT"`
}

// ConfirmResponse — response POST /transactions/:id/confirm.
// QrisString diisi hanya ketika payment_method = QRIS; nil untuk metode lain.
type ConfirmResponse struct {
	Order
	QrisString *string `json:"qris_string,omitempty"`
}

// ----------------------------------------------------------------
// List filter + result
// ----------------------------------------------------------------

type ListFilter struct {
	StartDate     *time.Time
	EndDate       *time.Time
	Status        string
	PaymentMethod string
	OperatorID    *uuid.UUID
	Page          int
	Limit         int
}

type ListResult struct {
	Transactions []*Order `json:"transactions"`
	Total        int      `json:"total"`
	Page         int      `json:"page"`
	Limit        int      `json:"limit"`
}
