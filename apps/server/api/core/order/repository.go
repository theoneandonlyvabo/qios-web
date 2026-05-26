// core/order/repository.go
//
// Layer akses data untuk domain order.
// Semua interaksi langsung ke tabel orders, order_items, order_sessions ada di sini.

package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Repository mendefinisikan kontrak akses data pos.
type Repository interface {
	// Products
	FindProducts(ctx context.Context, businessID uuid.UUID, ids []uuid.UUID) ([]productSnapshot, error)

	// Orders
	Create(ctx context.Context, order *Order, items []*OrderItem) error
	FindByID(ctx context.Context, id, businessID uuid.UUID) (*OrderWithItems, error)
	FindByOperatorToday(ctx context.Context, operatorID, businessID uuid.UUID) ([]*Order, error)
	UpdateItems(ctx context.Context, orderID, businessID uuid.UUID, items []*OrderItem) error
	BeginCheckout(ctx context.Context, orderID, businessID uuid.UUID) error
	GetCheckoutStartedAt(ctx context.Context, orderID, businessID uuid.UUID) (*time.Time, error)
	Confirm(ctx context.Context, orderID, businessID uuid.UUID, method PaymentMethod, confirmedAt time.Time) error
	// ConfirmAtomic performs the timing guard and the status transition atomically
	// inside a single transaction using SELECT FOR UPDATE. This prevents a race
	// where two concurrent confirms both pass the 800ms guard then both fire
	// downstream side-effects before one of them gets ErrNotPending.
	ConfirmAtomic(ctx context.Context, orderID, businessID uuid.UUID, method PaymentMethod, confirmedAt time.Time, minElapsed time.Duration) error
	Void(ctx context.Context, orderID, businessID uuid.UUID, callerOperatorID *uuid.UUID) error

	// Recipe + consumption
	FindProductRecipes(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID][]RecipeItem, error)
	GetBusinessQrisString(ctx context.Context, businessID uuid.UUID) (*string, error)
	InsertConsumptionLog(ctx context.Context, entries []ConsumptionEntry) error

	// Sessions
	CreateSession(ctx context.Context, s *Session) error
	EndSession(ctx context.Context, sessionID, businessID uuid.UUID) error
	ListActiveSessions(ctx context.Context, businessID uuid.UUID) ([]*SessionWithOperator, error)
	TouchSession(ctx context.Context, sessionID uuid.UUID) error
}

// PostgresRepository implementasi Repository di atas database/sql.
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
		return nil, fmt.Errorf("order: find products: %w", err)
	}
	defer rows.Close()

	var out []productSnapshot
	for rows.Next() {
		var s productSnapshot
		if err := rows.Scan(&s.id, &s.name, &s.price); err != nil {
			return nil, fmt.Errorf("order: scan product: %w", err)
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
		return fmt.Errorf("order: begin: %w", err)
	}

	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders
		 (business_id, operator_id, order_id, total_amount, status, note)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		order.BusinessID, order.OperatorID, order.OrderID,
		order.TotalAmount, order.Status, order.Note,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("order: insert order: %w", err)
	}

	for _, item := range items {
		err = tx.QueryRowContext(ctx,
			`INSERT INTO order_items (order_id, product_id, product_name, unit_price, quantity)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id`,
			order.ID, item.ProductID, item.ProductName, item.UnitPrice, item.Quantity,
		).Scan(&item.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("order: insert item: %w", err)
		}
		item.Subtotal = item.UnitPrice * int64(item.Quantity)
	}

	return tx.Commit()
}

// ----------------------------------------------------------------
// FindByID
// ----------------------------------------------------------------

const orderColumns = `id, business_id, operator_id, order_id, total_amount,
       payment_method, status, note, checkout_started_at, confirmed_at, created_at, updated_at`

func scanOrder(row interface{ Scan(...any) error }) (*Order, error) {
	var o Order
	var operatorID uuid.NullUUID
	var paymentMethod sql.NullString
	var note sql.NullString
	var checkoutStartedAt sql.NullTime
	var confirmedAt sql.NullTime

	err := row.Scan(
		&o.ID, &o.BusinessID, &operatorID, &o.OrderID, &o.TotalAmount,
		&paymentMethod, &o.Status, &note, &checkoutStartedAt, &confirmedAt,
		&o.CreatedAt, &o.UpdatedAt,
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
	if checkoutStartedAt.Valid {
		o.CheckoutStartedAt = &checkoutStartedAt.Time
	}
	if confirmedAt.Valid {
		o.ConfirmedAt = &confirmedAt.Time
	}
	return &o, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id, businessID uuid.UUID) (*OrderWithItems, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+orderColumns+` FROM orders WHERE id = $1 AND business_id = $2`,
		id, businessID,
	)
	order, err := scanOrder(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("order: find by id: %w", err)
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
		 FROM order_items WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("order: find items: %w", err)
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
			return nil, fmt.Errorf("order: scan item: %w", err)
		}
		if productID.Valid {
			item.ProductID = &productID.UUID
		}
		out = append(out, &item)
	}
	return out, rows.Err()
}

// ----------------------------------------------------------------
// FindByOperatorToday
// ----------------------------------------------------------------

func (r *PostgresRepository) FindByOperatorToday(ctx context.Context, operatorID, businessID uuid.UUID) ([]*Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+orderColumns+`
		 FROM orders
		 WHERE operator_id = $1 AND business_id = $2
		   AND DATE(created_at) = CURRENT_DATE
		 ORDER BY created_at DESC`,
		operatorID, businessID,
	)
	if err != nil {
		return nil, fmt.Errorf("order: list my orders: %w", err)
	}
	defer rows.Close()

	var out []*Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("order: scan order: %w", err)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// ----------------------------------------------------------------
// UpdateItems
// ----------------------------------------------------------------

func (r *PostgresRepository) UpdateItems(ctx context.Context, orderID, businessID uuid.UUID, items []*OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("order: begin: %w", err)
	}

	// Delete old items
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM order_items WHERE order_id = $1`,
		orderID,
	); err != nil {
		tx.Rollback()
		return fmt.Errorf("order: delete old items: %w", err)
	}

	// Insert new items and recalculate total
	var totalAmount int64
	for _, item := range items {
		item.Subtotal = item.UnitPrice * int64(item.Quantity)
		totalAmount += item.Subtotal
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO order_items (order_id, product_id, product_name, unit_price, quantity)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id`,
			orderID, item.ProductID, item.ProductName, item.UnitPrice, item.Quantity,
		).Scan(&item.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("order: insert item: %w", err)
		}
	}

	// Update total_amount
	res, err := tx.ExecContext(ctx,
		`UPDATE orders
		 SET total_amount = $1, updated_at = NOW()
		 WHERE id = $2 AND business_id = $3 AND status = 'DRAFT'`,
		totalAmount, orderID, businessID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("order: update total: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		tx.Rollback()
		return ErrNotDraft
	}

	return tx.Commit()
}

// ----------------------------------------------------------------
// BeginCheckout
// ----------------------------------------------------------------

func (r *PostgresRepository) BeginCheckout(ctx context.Context, orderID, businessID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders
		 SET checkout_started_at = NOW(), status = 'PENDING', updated_at = NOW()
		 WHERE id = $1 AND business_id = $2 AND status = 'DRAFT'`,
		orderID, businessID,
	)
	if err != nil {
		return fmt.Errorf("order: begin checkout: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotDraft
	}
	return nil
}

// ----------------------------------------------------------------
// GetCheckoutStartedAt
// ----------------------------------------------------------------

func (r *PostgresRepository) GetCheckoutStartedAt(ctx context.Context, orderID, businessID uuid.UUID) (*time.Time, error) {
	var t sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT checkout_started_at FROM orders WHERE id = $1 AND business_id = $2`,
		orderID, businessID,
	).Scan(&t)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("order: get checkout_started_at: %w", err)
	}
	if !t.Valid {
		return nil, nil
	}
	return &t.Time, nil
}

// ----------------------------------------------------------------
// Confirm
// ----------------------------------------------------------------

func (r *PostgresRepository) Confirm(ctx context.Context, orderID, businessID uuid.UUID, method PaymentMethod, confirmedAt time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders
		 SET status         = 'CONFIRMED',
		     payment_method = $1,
		     confirmed_at   = $2,
		     updated_at     = NOW()
		 WHERE id = $3 AND business_id = $4 AND status = 'PENDING'`,
		string(method), confirmedAt, orderID, businessID,
	)
	if err != nil {
		return fmt.Errorf("order: confirm: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotPending
	}
	return nil
}

// ----------------------------------------------------------------
// ConfirmAtomic
// ----------------------------------------------------------------

func (r *PostgresRepository) ConfirmAtomic(ctx context.Context, orderID, businessID uuid.UUID, method PaymentMethod, confirmedAt time.Time, minElapsed time.Duration) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("order: confirm begin: %w", err)
	}

	var status string
	var startedAt sql.NullTime
	err = tx.QueryRowContext(ctx,
		`SELECT status, checkout_started_at FROM orders
		 WHERE id = $1 AND business_id = $2
		 FOR UPDATE`,
		orderID, businessID,
	).Scan(&status, &startedAt)
	if errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return ErrNotFound
	}
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("order: confirm select: %w", err)
	}
	if status != "PENDING" {
		tx.Rollback()
		return ErrNotPending
	}
	if !startedAt.Valid {
		tx.Rollback()
		return ErrCheckoutNotStarted
	}
	if time.Since(startedAt.Time) < minElapsed {
		tx.Rollback()
		return ErrGestureTooFast
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE orders
		 SET status         = 'CONFIRMED',
		     payment_method = $1,
		     confirmed_at   = $2,
		     updated_at     = NOW()
		 WHERE id = $3 AND business_id = $4`,
		string(method), confirmedAt, orderID, businessID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("order: confirm update: %w", err)
	}
	return tx.Commit()
}

// ----------------------------------------------------------------
// Void
// ----------------------------------------------------------------

func (r *PostgresRepository) Void(ctx context.Context, orderID, businessID uuid.UUID, callerOperatorID *uuid.UUID) error {
	var res sql.Result
	var err error
	if callerOperatorID != nil {
		// Operator: can only void their own orders.
		res, err = r.db.ExecContext(ctx,
			`UPDATE orders
			 SET status = 'VOIDED', updated_at = NOW()
			 WHERE id = $1 AND business_id = $2 AND operator_id = $3 AND status IN ('DRAFT', 'PENDING')`,
			orderID, businessID, *callerOperatorID,
		)
	} else {
		// Owner: can void any order in their business.
		res, err = r.db.ExecContext(ctx,
			`UPDATE orders
			 SET status = 'VOIDED', updated_at = NOW()
			 WHERE id = $1 AND business_id = $2 AND status IN ('DRAFT', 'PENDING')`,
			orderID, businessID,
		)
	}
	if err != nil {
		return fmt.Errorf("order: void: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// ----------------------------------------------------------------
// FindProductRecipes (copied from transaction/repository.go)
// ----------------------------------------------------------------

func (r *PostgresRepository) FindProductRecipes(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID][]RecipeItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, recipe FROM products WHERE id = ANY($1) AND recipe IS NOT NULL`,
		pq.Array(productIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("order: find recipes: %w", err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID][]RecipeItem)
	for rows.Next() {
		var id uuid.UUID
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			return nil, err
		}
		var items []RecipeItem
		if err := json.Unmarshal(raw, &items); err == nil {
			out[id] = items
		}
	}
	return out, rows.Err()
}

// ----------------------------------------------------------------
// GetBusinessQrisString (copied from transaction/repository.go)
// ----------------------------------------------------------------

func (r *PostgresRepository) GetBusinessQrisString(ctx context.Context, businessID uuid.UUID) (*string, error) {
	var qs sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT qris_string FROM businesses WHERE id = $1`,
		businessID,
	).Scan(&qs)
	if err != nil {
		return nil, fmt.Errorf("order: get business qris_string: %w", err)
	}
	if qs.Valid {
		return &qs.String, nil
	}
	return nil, nil
}

// ----------------------------------------------------------------
// InsertConsumptionLog (copied from transaction/repository.go)
// ----------------------------------------------------------------

func (r *PostgresRepository) InsertConsumptionLog(ctx context.Context, entries []ConsumptionEntry) error {
	if len(entries) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO consumption_log
			 (transaction_id, business_id, product_id, product_name, ingredient, quantity_used, unit, confirmed_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			e.TransactionID, e.BusinessID, e.ProductID, e.ProductName,
			e.Ingredient, e.QuantityUsed, e.Unit, e.ConfirmedAt,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("order: insert consumption_log: %w", err)
		}
	}
	return tx.Commit()
}

// ----------------------------------------------------------------
// Session methods
// ----------------------------------------------------------------

func (r *PostgresRepository) CreateSession(ctx context.Context, s *Session) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO order_sessions (operator_id, business_id)
		 VALUES ($1, $2)
		 RETURNING id, started_at, last_seen_at`,
		s.OperatorID, s.BusinessID,
	).Scan(&s.ID, &s.StartedAt, &s.LastSeenAt)
	if err != nil {
		return fmt.Errorf("order: create session: %w", err)
	}
	return nil
}

func (r *PostgresRepository) EndSession(ctx context.Context, sessionID, businessID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE order_sessions
		 SET ended_at = NOW()
		 WHERE id = $1 AND business_id = $2 AND ended_at IS NULL`,
		sessionID, businessID,
	)
	if err != nil {
		return fmt.Errorf("order: end session: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (r *PostgresRepository) ListActiveSessions(ctx context.Context, businessID uuid.UUID) ([]*SessionWithOperator, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT ps.id, ps.operator_id, ps.business_id, ps.started_at, ps.ended_at, ps.last_seen_at,
		        o.name, o.operator_code
		 FROM order_sessions ps
		 JOIN operators o ON o.id = ps.operator_id
		 WHERE ps.business_id = $1 AND ps.ended_at IS NULL
		 ORDER BY ps.started_at DESC`,
		businessID,
	)
	if err != nil {
		return nil, fmt.Errorf("order: list sessions: %w", err)
	}
	defer rows.Close()

	var out []*SessionWithOperator
	for rows.Next() {
		var sw SessionWithOperator
		var endedAt sql.NullTime
		if err := rows.Scan(
			&sw.ID, &sw.OperatorID, &sw.BusinessID, &sw.StartedAt,
			&endedAt, &sw.LastSeenAt, &sw.OperatorName, &sw.OperatorCode,
		); err != nil {
			return nil, fmt.Errorf("order: scan session: %w", err)
		}
		if endedAt.Valid {
			sw.EndedAt = &endedAt.Time
		}
		out = append(out, &sw)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) TouchSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE order_sessions SET last_seen_at = NOW() WHERE id = $1 AND ended_at IS NULL`,
		sessionID,
	)
	return err
}
