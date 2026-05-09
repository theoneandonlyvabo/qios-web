// domain/payment/repository.go
//
// Interface dan skeleton implementasi untuk akses data payment domain.
// Semua SQL ada di sini — service dan handler tidak boleh menyentuh database langsung.

package payment

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// Repository mendefinisikan kontrak akses data untuk pos_orders dan pos_order_items.
type Repository interface {
	// CreateWithItems membuat pos_order dan pos_order_items dalam satu transaksi.
	// tx dioper dari service karena create order bisa perlu external call (Xendit) di tx yang sama.
	CreateWithItems(ctx context.Context, tx *sql.Tx, order *PosOrder, items []*OrderItem) error

	// FindByID mengambil order by UUID primary key.
	FindByID(ctx context.Context, id uuid.UUID) (*PosOrder, []*OrderItem, error)

	// FindByOrderID mengambil order by order_id string (QIOS-YYYYMMDD-xxxx).
	// Dipakai webhook handler untuk matching notifikasi Xendit ke order.
	FindByOrderID(ctx context.Context, orderID string) (*PosOrder, error)

	// FindByBusinessID mengambil list order dengan filter dan pagination.
	FindByBusinessID(ctx context.Context, businessID uuid.UUID, filter ListOrdersFilter) ([]*PosOrder, int, error)

	// UpdateStatus mengupdate status dan paid_at order.
	// Dipakai webhook handler dan cash complete handler.
	UpdateStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status OrderStatus, paidAt *string) error
}

// PostgresRepository adalah implementasi Repository di atas database/sql.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateWithItems(ctx context.Context, tx *sql.Tx, order *PosOrder, items []*OrderItem) error {
	// TODO: implement
	// 1. INSERT pos_orders RETURNING id
	// 2. INSERT pos_order_items (batch)
	// 3. Isi order.ID dan item IDs dari RETURNING
	panic("not implemented")
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*PosOrder, []*OrderItem, error) {
	// TODO: implement
	// SELECT pos_orders + JOIN pos_order_items WHERE pos_orders.id = $1
	panic("not implemented")
}

func (r *PostgresRepository) FindByOrderID(ctx context.Context, orderID string) (*PosOrder, error) {
	// TODO: implement
	// SELECT * FROM pos_orders WHERE order_id = $1
	panic("not implemented")
}

func (r *PostgresRepository) FindByBusinessID(ctx context.Context, businessID uuid.UUID, filter ListOrdersFilter) ([]*PosOrder, int, error) {
	// TODO: implement
	// SELECT + COUNT(*) OVER() untuk pagination
	// Filter: status, start_date, end_date
	// Pagination: LIMIT + OFFSET
	panic("not implemented")
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status OrderStatus, paidAt *string) error {
	// TODO: implement
	// UPDATE pos_orders SET status = $1, paid_at = $2, updated_at = NOW() WHERE id = $3
	panic("not implemented")
}
