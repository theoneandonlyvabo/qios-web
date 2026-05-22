// core/transaction/repository.go
//
// Layer akses data untuk domain transaction.
// Semua interaksi langsung ke tabel pos_orders dan pos_order_items ada di sini.

package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Repository mendefinisikan kontrak akses data transaksi.
type Repository interface {
	// FindProducts dipakai service saat membuat order — ambil snapshot produk.
	FindProducts(ctx context.Context, businessID uuid.UUID, ids []uuid.UUID) ([]productSnapshot, error)

	// Create insert order + semua items dalam satu transaksi DB.
	Create(ctx context.Context, order *Order, items []*OrderItem) error

	// FindByID mengambil order beserta item-itemnya.
	FindByID(ctx context.Context, id, businessID uuid.UUID) (*OrderWithItems, error)

	// List mengambil orders dengan filter dan pagination.
	List(ctx context.Context, businessID uuid.UUID, f ListFilter) ([]*Order, int, error)

	// UpdateStatus mengupdate status dan field terkait (paid_at, payment_method).
	UpdateStatus(ctx context.Context, id, businessID uuid.UUID, status Status, method *PaymentMethod, paidAt *sql.NullTime) error

	// GetBusinessQrisString mengambil qris_string dari tabel businesses.
	GetBusinessQrisString(ctx context.Context, businessID uuid.UUID) (*string, error)
}

// productSnapshot adalah data produk yang di-snapshot ke order item.
type productSnapshot struct {
	id    uuid.UUID
	name  string
	price int64
}

// PostgresRepository adalah implementasi Repository di atas database/sql.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ----------------------------------------------------------------
// FindProducts
// ----------------------------------------------------------------

func (r *PostgresRepository) FindProducts(ctx context.Context, businessID uuid.UUID, ids []uuid.UUID) ([]productSnapshot, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, price FROM products
		 WHERE id = ANY($1) AND business_id = $2 AND deleted_at IS NULL AND is_available = TRUE`,
		pq.Array(ids), businessID,
	)
	if err != nil {
		return nil, fmt.Errorf("transaction: find products: %w", err)
	}
	defer rows.Close()

	var out []productSnapshot
	for rows.Next() {
		var s productSnapshot
		if err := rows.Scan(&s.id, &s.name, &s.price); err != nil {
			return nil, fmt.Errorf("transaction: scan product: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ----------------------------------------------------------------
// Create
// ----------------------------------------------------------------

func (r *PostgresRepository) Create(ctx context.Context, order *Order, items []*OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transaction: begin: %w", err)
	}

	err = tx.QueryRowContext(ctx,
		`INSERT INTO pos_orders
		 (business_id, operator_id, order_id, total_amount, payment_method, status, note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		order.BusinessID, order.OperatorID, order.OrderID,
		order.TotalAmount, order.PaymentMethod, order.Status, order.Note,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("transaction: insert order: %w", err)
	}

	for _, item := range items {
		err = tx.QueryRowContext(ctx,
			`INSERT INTO pos_order_items (pos_order_id, product_id, product_name, unit_price, quantity)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id`,
			order.ID, item.ProductID, item.ProductName, item.UnitPrice, item.Quantity,
		).Scan(&item.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("transaction: insert item: %w", err)
		}
	}

	return tx.Commit()
}

// ----------------------------------------------------------------
// FindByID
// ----------------------------------------------------------------

const orderColumns = `id, business_id, operator_id, order_id, total_amount,
       payment_method, status, note, paid_at, created_at, updated_at`

func scanOrder(row interface{ Scan(...any) error }) (*Order, error) {
	var o Order
	var operatorID uuid.NullUUID
	var paymentMethod sql.NullString
	var note sql.NullString
	var paidAt sql.NullTime

	err := row.Scan(
		&o.ID, &o.BusinessID, &operatorID, &o.OrderID, &o.TotalAmount,
		&paymentMethod, &o.Status, &note, &paidAt, &o.CreatedAt, &o.UpdatedAt,
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
	if paidAt.Valid {
		o.PaidAt = &paidAt.Time
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

// ----------------------------------------------------------------
// GetBusinessQrisString
// ----------------------------------------------------------------

func (r *PostgresRepository) GetBusinessQrisString(ctx context.Context, businessID uuid.UUID) (*string, error) {
	var qs sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT qris_string FROM businesses WHERE id = $1`,
		businessID,
	).Scan(&qs)
	if err != nil {
		return nil, fmt.Errorf("transaction: get business qris_string: %w", err)
	}
	if qs.Valid {
		return &qs.String, nil
	}
	return nil, nil
}

// ----------------------------------------------------------------
// UpdateStatus
// ----------------------------------------------------------------

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id, businessID uuid.UUID, status Status, method *PaymentMethod, paidAt *sql.NullTime) error {
	var methodStr *string
	if method != nil {
		s := string(*method)
		methodStr = &s
	}

	var paidAtVal interface{}
	if paidAt != nil && paidAt.Valid {
		paidAtVal = paidAt.Time
	}

	res, err := r.db.ExecContext(ctx,
		`UPDATE pos_orders
		 SET status         = $1,
		     payment_method = COALESCE($2, payment_method),
		     paid_at        = COALESCE($3, paid_at),
		     updated_at     = NOW()
		 WHERE id = $4 AND business_id = $5 AND status = 'pending'`,
		status, methodStr, paidAtVal, id, businessID,
	)
	if err != nil {
		return fmt.Errorf("transaction: update status: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		// Either not found or not pending — distinguish via FindByID is too expensive;
		// return ErrNotPending as the common case (handler will 409).
		return ErrNotPending
	}
	return nil
}
