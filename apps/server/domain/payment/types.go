// domain/payment/types.go
//
// Domain types, sentinel errors, dan DTOs untuk payment domain.
// Pisah dari model.go yang berisi tipe Xendit platform (ManagedAccountInput, dll).

package payment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ----------------------------------------------------------------
// Sentinel errors
// ----------------------------------------------------------------

var ErrOrderNotFound          = errors.New("order not found")
var ErrOrderAlreadyPaid       = errors.New("order already paid")
var ErrInvalidStatus          = errors.New("invalid order status transition")
var ErrProductNotFound        = errors.New("product not found")
var ErrInvalidTotal           = errors.New("order total must be greater than zero")
var ErrBusinessNotFound       = errors.New("business not found")
var ErrXenditNotActive        = errors.New("business xendit account is not active")
var ErrXenditPaymentNotFound  = errors.New("xendit payment not found")

// ----------------------------------------------------------------
// Enums
// ----------------------------------------------------------------

// PaymentMethod adalah metode pembayaran yang didukung.
// Ditambah ke pos_orders.payment_method — lihat migration 014.
type PaymentMethod string

const (
	PaymentMethodCash           PaymentMethod = "CASH"
	PaymentMethodQRIS           PaymentMethod = "QRIS"
	PaymentMethodEWallet        PaymentMethod = "EWALLET"
	PaymentMethodVirtualAccount PaymentMethod = "VIRTUAL_ACCOUNT"
)

// OrderStatus mengikuti status di pos_orders.
// Catatan: migration 008 pakai lowercase — perlu konsistensi di migration 014.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusFailed    OrderStatus = "FAILED"
	OrderStatusExpired   OrderStatus = "EXPIRED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// ----------------------------------------------------------------
// Domain structs
// ----------------------------------------------------------------

// PosOrder merepresentasikan baris di tabel pos_orders.
type PosOrder struct {
	ID            uuid.UUID     `json:"id"`
	BusinessID    uuid.UUID     `json:"business_id"`
	OperatorID    *uuid.UUID    `json:"operator_id,omitempty"`
	OrderID       string        `json:"order_id"` // format: QIOS-YYYYMMDD-xxxx
	TotalAmount   int64         `json:"total_amount"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Status        OrderStatus   `json:"status"`
	Note          string        `json:"note,omitempty"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// XenditPayment merepresentasikan baris di tabel xendit_payments.
// Disimpan saat QR di-generate dan diupdate saat webhook payment masuk.
type XenditPayment struct {
	ID              uuid.UUID
	PosOrderID      uuid.UUID
	XenditAccountID string
	XenditInvoiceID string
	XenditChargeID  string
	PaymentMethod   PaymentMethod
	Amount          int64
	Status          string // "PENDING" | "PAID" | "FAILED" | "EXPIRED" | ... (uppercase di table)
	QRString        string
	RawPayload      []byte
	PaidAt          *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// OrderItem merepresentasikan baris di tabel pos_order_items.
type OrderItem struct {
	ID          uuid.UUID  `json:"id"`
	PosOrderID  uuid.UUID  `json:"pos_order_id"`
	ProductID   *uuid.UUID `json:"product_id,omitempty"`
	ProductName string     `json:"product_name"` // snapshot saat transaksi
	UnitPrice   int64      `json:"unit_price"`   // snapshot saat transaksi
	Quantity    int        `json:"quantity"`
	Subtotal    int64      `json:"subtotal"` // generated: quantity * unit_price
}

// ----------------------------------------------------------------
// Request DTOs
// ----------------------------------------------------------------

// CreateOrderItemInput adalah satu item dalam request create order.
type CreateOrderItemInput struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// CreateOrderRequest adalah body POST /transactions dari kasir.
type CreateOrderRequest struct {
	PaymentMethod PaymentMethod          `json:"payment_method"`
	Items         []CreateOrderItemInput `json:"items"`
	Note          string                 `json:"note"`
}

// ListOrdersFilter adalah query params untuk GET /transactions.
type ListOrdersFilter struct {
	Status    string // "pending" | "paid" | "failed" | "expired" | ""
	StartDate string // YYYY-MM-DD
	EndDate   string // YYYY-MM-DD
	Page      int    // 1-based, default 1
	Limit     int    // default 20
}

// ----------------------------------------------------------------
// Response DTOs
// ----------------------------------------------------------------

// OrderItemResponse adalah representasi item untuk response.
type OrderItemResponse struct {
	ProductID   *uuid.UUID `json:"product_id,omitempty"`
	ProductName string     `json:"product_name"`
	UnitPrice   int64      `json:"unit_price"`
	Quantity    int        `json:"quantity"`
	Subtotal    int64      `json:"subtotal"`
}

// OrderResponse adalah response untuk create dan get order.
// QRString hanya diisi pada CreateOrder QRIS — kasir render langsung dari field ini.
type OrderResponse struct {
	ID            uuid.UUID           `json:"id"`
	OrderID       string              `json:"order_id"`
	TotalAmount   int64               `json:"total_amount"`
	PaymentMethod PaymentMethod       `json:"payment_method"`
	Status        OrderStatus         `json:"status"`
	Note          string              `json:"note,omitempty"`
	PaidAt        *time.Time          `json:"paid_at,omitempty"`
	Items         []OrderItemResponse `json:"items"`
	CreatedAt     time.Time           `json:"created_at"`
	QRString      string              `json:"qr_string,omitempty"`
}
