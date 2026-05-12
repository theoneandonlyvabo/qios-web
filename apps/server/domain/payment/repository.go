// domain/payment/repository.go
//
// Interface dan implementasi PostgreSQL untuk akses data payment domain.
// Semua SQL ada di sini — service dan handler tidak boleh menyentuh database langsung.

package payment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Repository mendefinisikan kontrak akses data untuk pos_orders dan pos_order_items.
type Repository interface {
	// CreateWithItems membuat pos_order dan pos_order_items dalam satu transaksi.
	// tx dioper dari service karena create order bisa perlu external call (Xendit) di tx yang sama.
	CreateWithItems(ctx context.Context, tx *sql.Tx, order *PosOrder, items []*OrderItem) error

	// FindByID mengambil order by UUID primary key beserta item-itemnya.
	FindByID(ctx context.Context, id uuid.UUID) (*PosOrder, []*OrderItem, error)

	// FindByOrderID mengambil order by order_id string.
	// Dipakai webhook handler untuk matching notifikasi Xendit ke order.
	FindByOrderID(ctx context.Context, orderID string) (*PosOrder, error)

	// FindByBusinessID mengambil list order dengan filter dan pagination.
	FindByBusinessID(ctx context.Context, businessID uuid.UUID, filter ListOrdersFilter) ([]*PosOrder, int, error)

	// UpdateStatus mengupdate status dan paid_at order.
	// Dipakai webhook handler dan cash complete handler.
	UpdateStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status OrderStatus, paidAt *time.Time) error

	// GetBusinessXenditAccount mengembalikan xendit_account_id + xendit_status
	// untuk satu bisnis. Dipakai saat membuat QR dinamis atas nama merchant.
	GetBusinessXenditAccount(ctx context.Context, businessID uuid.UUID) (accountID string, status XenditStatus, err error)

	// InsertXenditPayment mencatat record pembayaran Xendit yang dibuat untuk order.
	// Dipanggil dari service saat QR berhasil di-generate; juga dari webhook saat
	// notifikasi datang untuk order yang belum punya row xendit_payments.
	InsertXenditPayment(ctx context.Context, tx *sql.Tx, row *XenditPayment) error

	// MarkXenditPaymentPaid update row xendit_payments + raw_payload via order_id.
	MarkXenditPaymentPaid(ctx context.Context, tx *sql.Tx, posOrderID uuid.UUID, status string, paidAt time.Time, raw []byte) error
}

// PostgresRepository adalah implementasi Repository di atas database/sql.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// dbStatus / domainStatus mapping case di boundary repo.
// pos_orders.status disimpan lowercase (migration 008); domain enum uppercase.
// Lihat AGENTS.md section 7 untuk konteks.
func dbStatus(s OrderStatus) string       { return strings.ToLower(string(s)) }
func domainStatus(s string) OrderStatus   { return OrderStatus(strings.ToUpper(s)) }

func (r *PostgresRepository) CreateWithItems(ctx context.Context, tx *sql.Tx, order *PosOrder, items []*OrderItem) error {
	var operator uuid.NullUUID
	if order.OperatorID != nil {
		operator = uuid.NullUUID{UUID: *order.OperatorID, Valid: true}
	}
	err := tx.QueryRowContext(ctx, `
		INSERT INTO pos_orders (business_id, operator_id, order_id, total_amount, payment_method, status, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		order.BusinessID,
		operator,
		order.OrderID,
		order.TotalAmount,
		string(order.PaymentMethod),
		dbStatus(order.Status),
		sql.NullString{String: order.Note, Valid: order.Note != ""},
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("payment: insert pos_orders: %w", err)
	}

	for _, item := range items {
		item.PosOrderID = order.ID
		var product uuid.NullUUID
		if item.ProductID != nil {
			product = uuid.NullUUID{UUID: *item.ProductID, Valid: true}
		}
		err := tx.QueryRowContext(ctx, `
			INSERT INTO pos_order_items (pos_order_id, product_id, product_name, unit_price, quantity)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, subtotal`,
			item.PosOrderID, product, item.ProductName, item.UnitPrice, item.Quantity,
		).Scan(&item.ID, &item.Subtotal)
		if err != nil {
			return fmt.Errorf("payment: insert pos_order_items: %w", err)
		}
	}
	return nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*PosOrder, []*OrderItem, error) {
	order, err := r.scanOrderRow(ctx, r.db, `
		SELECT id, business_id, operator_id, order_id, total_amount, payment_method, status, note, paid_at, created_at, updated_at
		FROM pos_orders WHERE id = $1`, id)
	if err != nil {
		return nil, nil, err
	}

	items, err := r.findItems(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return order, items, nil
}

func (r *PostgresRepository) FindByOrderID(ctx context.Context, orderID string) (*PosOrder, error) {
	return r.scanOrderRow(ctx, r.db, `
		SELECT id, business_id, operator_id, order_id, total_amount, payment_method, status, note, paid_at, created_at, updated_at
		FROM pos_orders WHERE order_id = $1`, orderID)
}

func (r *PostgresRepository) FindByBusinessID(ctx context.Context, businessID uuid.UUID, filter ListOrdersFilter) ([]*PosOrder, int, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}

	query := `
		SELECT id, business_id, operator_id, order_id, total_amount, payment_method, status, note, paid_at, created_at, updated_at,
		       COUNT(*) OVER() AS total_count
		FROM pos_orders
		WHERE business_id = $1`
	args := []any{businessID}
	idx := 2

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, strings.ToLower(filter.Status))
		idx++
	}
	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND created_at >= $%d", idx)
		args = append(args, filter.StartDate)
		idx++
	}
	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND created_at < ($%d::date + INTERVAL '1 day')", idx)
		args = append(args, filter.EndDate)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("payment: list orders query: %w", err)
	}
	defer rows.Close()

	var (
		orders []*PosOrder
		total  int
	)
	for rows.Next() {
		o, count, err := scanOrderListRow(rows)
		if err != nil {
			return nil, 0, err
		}
		total = count
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("payment: list orders iter: %w", err)
	}
	return orders, total, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status OrderStatus, paidAt *time.Time) error {
	var paid sql.NullTime
	if paidAt != nil {
		paid = sql.NullTime{Time: *paidAt, Valid: true}
	}
	var (
		result sql.Result
		err    error
	)
	const stmt = `UPDATE pos_orders SET status = $1, paid_at = $2, updated_at = NOW() WHERE id = $3`
	if tx != nil {
		result, err = tx.ExecContext(ctx, stmt, dbStatus(status), paid, id)
	} else {
		result, err = r.db.ExecContext(ctx, stmt, dbStatus(status), paid, id)
	}
	if err != nil {
		return fmt.Errorf("payment: update status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrOrderNotFound
	}
	return nil
}

// ----------------------------------------------------------------
// Scan helpers
// ----------------------------------------------------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *PostgresRepository) scanOrderRow(ctx context.Context, db *sql.DB, query string, arg any) (*PosOrder, error) {
	row := db.QueryRowContext(ctx, query, arg)
	o, _, err := scanOrder(row, false)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("payment: scan order: %w", err)
	}
	return o, nil
}

func scanOrderListRow(r rowScanner) (*PosOrder, int, error) {
	o, count, err := scanOrder(r, true)
	if err != nil {
		return nil, 0, fmt.Errorf("payment: scan order list row: %w", err)
	}
	return o, count, nil
}

// scanOrder reads kolom standar pos_orders. Kalau withCount, tambahkan total_count di akhir.
func scanOrder(r rowScanner, withCount bool) (*PosOrder, int, error) {
	var (
		o          PosOrder
		operator   uuid.NullUUID
		note       sql.NullString
		paidAt     sql.NullTime
		statusStr  string
		methodStr  string
		totalCount int
	)
	dests := []any{
		&o.ID, &o.BusinessID, &operator, &o.OrderID, &o.TotalAmount,
		&methodStr, &statusStr, &note, &paidAt, &o.CreatedAt, &o.UpdatedAt,
	}
	if withCount {
		dests = append(dests, &totalCount)
	}
	if err := r.Scan(dests...); err != nil {
		return nil, 0, err
	}
	if operator.Valid {
		opID := operator.UUID
		o.OperatorID = &opID
	}
	o.Note = note.String
	o.PaymentMethod = PaymentMethod(methodStr)
	o.Status = domainStatus(statusStr)
	if paidAt.Valid {
		t := paidAt.Time
		o.PaidAt = &t
	}
	return &o, totalCount, nil
}

// ----------------------------------------------------------------
// Business + xendit_payments queries
// ----------------------------------------------------------------

func (r *PostgresRepository) GetBusinessXenditAccount(ctx context.Context, businessID uuid.UUID) (string, XenditStatus, error) {
	var (
		accountID sql.NullString
		statusStr string
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT xendit_account_id, xendit_status
		FROM businesses
		WHERE id = $1`, businessID,
	).Scan(&accountID, &statusStr)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrBusinessNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("payment: get business xendit: %w", err)
	}
	return accountID.String, XenditStatus(statusStr), nil
}

func (r *PostgresRepository) InsertXenditPayment(ctx context.Context, tx *sql.Tx, row *XenditPayment) error {
	const stmt = `
		INSERT INTO xendit_payments
		    (pos_order_id, xendit_account_id, xendit_invoice_id, xendit_charge_id,
		     payment_method, amount, status, qr_string, raw_payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`
	args := []any{
		row.PosOrderID,
		row.XenditAccountID,
		nullString(row.XenditInvoiceID),
		nullString(row.XenditChargeID),
		nullString(string(row.PaymentMethod)),
		row.Amount,
		row.Status,
		nullString(row.QRString),
		row.RawPayload,
	}
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, stmt, args...).Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt)
	} else {
		err = r.db.QueryRowContext(ctx, stmt, args...).Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt)
	}
	if err != nil {
		return fmt.Errorf("payment: insert xendit_payment: %w", err)
	}
	return nil
}

func (r *PostgresRepository) MarkXenditPaymentPaid(ctx context.Context, tx *sql.Tx, posOrderID uuid.UUID, status string, paidAt time.Time, raw []byte) error {
	const stmt = `
		UPDATE xendit_payments
		   SET status      = $1,
		       paid_at     = $2,
		       raw_payload = COALESCE($3, raw_payload),
		       updated_at  = NOW()
		 WHERE pos_order_id = $4`
	var (
		result sql.Result
		err    error
	)
	if tx != nil {
		result, err = tx.ExecContext(ctx, stmt, status, paidAt, raw, posOrderID)
	} else {
		result, err = r.db.ExecContext(ctx, stmt, status, paidAt, raw, posOrderID)
	}
	if err != nil {
		return fmt.Errorf("payment: mark xendit paid: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrXenditPaymentNotFound
	}
	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func (r *PostgresRepository) findItems(ctx context.Context, orderID uuid.UUID) ([]*OrderItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, pos_order_id, product_id, product_name, unit_price, quantity, subtotal
		FROM pos_order_items
		WHERE pos_order_id = $1
		ORDER BY id`, orderID)
	if err != nil {
		return nil, fmt.Errorf("payment: query order items: %w", err)
	}
	defer rows.Close()

	var items []*OrderItem
	for rows.Next() {
		var (
			it      OrderItem
			product uuid.NullUUID
		)
		if err := rows.Scan(&it.ID, &it.PosOrderID, &product, &it.ProductName, &it.UnitPrice, &it.Quantity, &it.Subtotal); err != nil {
			return nil, fmt.Errorf("payment: scan order item: %w", err)
		}
		if product.Valid {
			pid := product.UUID
			it.ProductID = &pid
		}
		items = append(items, &it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("payment: iter order items: %w", err)
	}
	return items, nil
}
