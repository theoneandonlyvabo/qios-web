// core/transaction/model.go
//
// Tipe-tipe untuk domain transaction (read-only log):
//   - Order, OrderItem, OrderWithItems → representasi baris tabel
//   - Status, PaymentMethod constants
//   - ListFilter, ListResult → filter dan pagination untuk List
//
// Domain ini hanya menyediakan akses baca ke history transaksi.
// Write operations (create, confirm, void) sudah dipindah ke domain /order.

package transaction

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("transaction not found")
)

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
	PaymentEwallet        PaymentMethod = "EWALLET"
	PaymentVirtualAccount PaymentMethod = "VIRTUAL_ACCOUNT"
	PaymentTransfer       PaymentMethod = "TRANSFER"
)

// Order merepresentasikan satu baris di orders.
type Order struct {
	ID            uuid.UUID      `json:"id"`
	BusinessID    uuid.UUID      `json:"business_id"`
	OperatorID    *uuid.UUID     `json:"operator_id"`
	OrderID       string         `json:"order_id"`
	TotalAmount   int64          `json:"total_amount"`
	PaymentMethod *PaymentMethod `json:"payment_method"`
	Status        Status         `json:"status"`
	Note          *string        `json:"note"`
	ConfirmedAt   *time.Time     `json:"confirmed_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// OrderItem merepresentasikan satu baris di order_items.
type OrderItem struct {
	ID          uuid.UUID  `json:"id"`
	ProductID   *uuid.UUID `json:"product_id"`
	ProductName string     `json:"product_name"`
	UnitPrice   int64      `json:"unit_price"`
	Quantity    int        `json:"quantity"`
	Subtotal    int64      `json:"subtotal"`
}

// OrderWithItems — response GET /transactions/:id.
type OrderWithItems struct {
	Order
	Items []*OrderItem `json:"items"`
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
