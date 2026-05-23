// core/pos/model.go
//
// Tipe-tipe untuk domain pos (Point of Sale):
//   - Order, OrderItem, OrderWithItems
//   - Session, SessionWithOperator
//   - Request types: CreateOrderRequest, UpdateItemsRequest, ConfirmOrderRequest
//   - Internal: productSnapshot, RecipeItem, ConsumptionEntry
//   - Status, PaymentMethod constants
//   - Error sentinels

package pos

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ----------------------------------------------------------------
// Errors
// ----------------------------------------------------------------

var (
	ErrNotFound           = errors.New("order not found")
	ErrNotDraft           = errors.New("order is not in draft status")
	ErrNotPending         = errors.New("order is not in pending status")
	ErrCheckoutNotStarted = errors.New("checkout not started — call begin-checkout first")
	ErrGestureTooFast     = errors.New("confirm gesture too fast — wait at least 800ms after begin-checkout")
	ErrEmptyItems         = errors.New("order must have at least one item")
	ErrProductNotFound    = errors.New("one or more products not found or unavailable")
	ErrSessionNotFound    = errors.New("session not found")
)

// ----------------------------------------------------------------
// Status + PaymentMethod
// ----------------------------------------------------------------

type Status string

const (
	StatusDraft     Status = "DRAFT"
	StatusPending   Status = "PENDING"
	StatusConfirmed Status = "CONFIRMED"
	StatusVoided    Status = "VOIDED"
)

type PaymentMethod string

const (
	PaymentCash           PaymentMethod = "CASH"
	PaymentQRIS           PaymentMethod = "QRIS"
	PaymentTransfer       PaymentMethod = "TRANSFER"
)

// ----------------------------------------------------------------
// Order types
// ----------------------------------------------------------------

// Order merepresentasikan satu baris di pos_orders.
type Order struct {
	ID                 uuid.UUID      `json:"id"`
	BusinessID         uuid.UUID      `json:"business_id"`
	OperatorID         *uuid.UUID     `json:"operator_id"`
	OrderID            string         `json:"order_id"`
	TotalAmount        int64          `json:"total_amount"`
	PaymentMethod      *PaymentMethod `json:"payment_method"`
	Status             Status         `json:"status"`
	Note               *string        `json:"note"`
	CheckoutStartedAt  *time.Time     `json:"checkout_started_at"`
	ConfirmedAt        *time.Time     `json:"confirmed_at"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
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

// OrderWithItems — order + item-itemnya untuk response detail.
type OrderWithItems struct {
	Order
	Items []*OrderItem `json:"items"`
}

// ConfirmResponse — response POST /pos/orders/:id/checkout/confirm.
// QrisString diisi hanya ketika payment_method = QRIS; nil untuk metode lain.
type ConfirmResponse struct {
	Order
	QrisString *string `json:"qris_string,omitempty"`
}

// ----------------------------------------------------------------
// Session types
// ----------------------------------------------------------------

// Session merepresentasikan sesi login operator aktif.
type Session struct {
	ID         uuid.UUID  `json:"id"`
	OperatorID uuid.UUID  `json:"operator_id"`
	BusinessID uuid.UUID  `json:"business_id"`
	StartedAt  time.Time  `json:"started_at"`
	EndedAt    *time.Time `json:"ended_at"`
	LastSeenAt time.Time  `json:"last_seen_at"`
}

// SessionWithOperator — session + info operator untuk list response.
type SessionWithOperator struct {
	Session
	OperatorName string `json:"operator_name"`
	OperatorCode string `json:"operator_code"`
}

// ----------------------------------------------------------------
// Request types
// ----------------------------------------------------------------

// ItemInput — satu baris item dalam order.
type ItemInput struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity"   validate:"required,min=1"`
}

// CreateOrderRequest — body POST /pos/orders.
type CreateOrderRequest struct {
	Items []ItemInput `json:"items" validate:"required,min=1,dive"`
	Note  *string     `json:"note"  validate:"omitempty,max=500"`
}

// UpdateItemsRequest — body PATCH /pos/orders/:id/items.
type UpdateItemsRequest struct {
	Items []ItemInput `json:"items" validate:"required,min=1,dive"`
	Note  *string     `json:"note"  validate:"omitempty,max=500"`
}

// ConfirmOrderRequest — body POST /pos/orders/:id/checkout/confirm.
type ConfirmOrderRequest struct {
	PaymentMethod PaymentMethod `json:"payment_method" validate:"required,oneof=CASH QRIS TRANSFER"`
}

// ----------------------------------------------------------------
// Internal types
// ----------------------------------------------------------------

// productSnapshot adalah data produk yang di-snapshot ke order item.
type productSnapshot struct {
	id    uuid.UUID
	name  string
	price int64
}

// RecipeItem — satu baris dalam JSONB recipe produk.
type RecipeItem struct {
	Ingredient string  `json:"ingredient"`
	Quantity   float64 `json:"quantity"`
	Unit       string  `json:"unit"`
}

// ConsumptionEntry — satu baris untuk tabel consumption_log.
type ConsumptionEntry struct {
	TransactionID uuid.UUID
	BusinessID    uuid.UUID
	ProductID     *uuid.UUID
	ProductName   string
	Ingredient    string
	QuantityUsed  float64
	Unit          string
	ConfirmedAt   time.Time
}
