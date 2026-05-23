// core/transaction/repository.go
//
// Layer akses data untuk domain transaction — read-only.
// Domain ini hanya menyediakan FindByID dan List untuk history transaksi.
// Write operations sudah dipindah ke domain /pos.

package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Repository mendefinisikan kontrak akses data transaksi (read-only).
type Repository interface {
	// FindByID mengambil order beserta item-itemnya.
	FindByID(ctx context.Context, id, businessID uuid.UUID) (*OrderWithItems, error)

	// List mengambil orders dengan filter dan pagination.
	List(ctx context.Context, businessID uuid.UUID, f ListFilter) ([]*Order, int, error)
}

// PostgresRepository adalah implementasi Repository di atas database/sql.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ----------------------------------------------------------------
// FindByID
// ----------------------------------------------------------------

const orderColumns = `id, business_id, operator_id, order_id, total_amount,
       payment_method, status, note, confirmed_at, created_at, updated_at`

func scanOrder(row interface{ Scan(...any) error }) (*Order, error) {
	var o Order
	var operatorID uuid.NullUUID
	var paymentMethod sql.NullString
	var note sql.NullString
	var confirmedAt sql.NullTime

	err := row.Scan(
		&o.ID, &o.BusinessID, &operatorID, &o.OrderID, &o.TotalAmount,
		&paymentMethod, &o.Status, &note, &confirmedAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if operatorID.Valid {
		o.OperatorID = &operatorID.UUID
	}
	if paymentMethod.Valid {
		m := PaymentMethod(paymentMethod.String)
		o.PaymentMethod = &m
	}
	if note.Valid {
		o.Note = &note.String
	}
	if confirmedAt.Valid {
		o.ConfirmedAt = &confirmedAt.Time
	}
	return &o, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id, businessID uuid.UUID) (*OrderWithItems, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+orderColumns+` FROM pos_orders WHERE id = $1 AND business_id = $2`,
		id, businessID,
	)
	order, err := scanOrder(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("transaction: find by id: %w", err)
	}

	items, err := r.findItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &OrderWithItems{Order: *order, Items: items}, nil
}

func (r *PostgresRepository) findItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*OrderItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, product_id, product_name, unit_price, quantity, subtotal
		 FROM pos_order_items WHERE pos_order_id = $1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("transaction: find items: %w", err)
	}
	defer rows.Close()

	var out []*OrderItem
	for rows.Next() {
		var item OrderItem
		var productID uuid.NullUUID
		if err := rows.Scan(
			&item.ID, &productID, &item.ProductName,
			&item.UnitPrice, &item.Quantity, &item.Subtotal,
		); err != nil {
			return nil, fmt.Errorf("transaction: scan item: %w", err)
		}
		if productID.Valid {
			item.ProductID = &productID.UUID
		}
		out = append(out, &item)
	}
	return out, rows.Err()
}

// ----------------------------------------------------------------
// List
// ----------------------------------------------------------------

func (r *PostgresRepository) List(ctx context.Context, businessID uuid.UUID, f ListFilter) ([]*Order, int, error) {
	where, args := buildListWhere(businessID, f)

	var total int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM pos_orders WHERE "+where,
		args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("transaction: count: %w", err)
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	limit := f.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx,
		"SELECT "+orderColumns+" FROM pos_orders WHERE "+where+
			fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("transaction: list: %w", err)
	}
	defer rows.Close()

	var out []*Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("transaction: scan: %w", err)
		}
		out = append(out, o)
	}
	return out, total, rows.Err()
}

func buildListWhere(businessID uuid.UUID, f ListFilter) (string, []any) {
	conds := []string{"business_id = $1"}
	args := []any{businessID}

	add := func(cond string, val any) {
		args = append(args, val)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}

	if f.StartDate != nil {
		add("DATE(created_at) >= DATE($%d)", f.StartDate)
	}
	if f.EndDate != nil {
		add("DATE(created_at) <= DATE($%d)", f.EndDate)
	}
	if f.Status != "" {
		add("status = $%d", f.Status)
	}
	if f.PaymentMethod != "" {
		add("payment_method = $%d", f.PaymentMethod)
	}
	if f.OperatorID != nil {
		add("operator_id = $%d", f.OperatorID)
	}

	return strings.Join(conds, " AND "), args
}
