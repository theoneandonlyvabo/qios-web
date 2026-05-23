// core/admin/repository.go
//
// Layer akses data untuk domain admin.
// Semua operasi DB dilakukan di sini — service dan handler tidak menyentuh DB.

package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	// Auth
	FindAdminByEmail(ctx context.Context, email string) (*AdminUser, error)
	FindAdminByID(ctx context.Context, id uuid.UUID) (*AdminUser, error)
	StoreAdminRefreshToken(ctx context.Context, adminID uuid.UUID, tokenHash string, expiry time.Duration) error
	FindAdminRefreshToken(ctx context.Context, tokenHash string) (adminID uuid.UUID, expiresAt time.Time, err error)
	DeleteAdminRefreshToken(ctx context.Context, tokenHash string) error

	// Business
	ListBusinesses(ctx context.Context) ([]*Business, error)
	FindBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error)
	CreateBusiness(ctx context.Context, req CreateBusinessRequest) (*Business, error)
	UpdateBusiness(ctx context.Context, b *Business) error

	// Product
	ListProductsByBusiness(ctx context.Context, businessID uuid.UUID) ([]*AdminProduct, error)
	FindProductByID(ctx context.Context, id uuid.UUID) (*AdminProduct, error)
	CreateProduct(ctx context.Context, businessID uuid.UUID, req AdminCreateProductRequest) (*AdminProduct, error)
	UpdateProduct(ctx context.Context, p *AdminProduct) error
	SoftDeleteProduct(ctx context.Context, id uuid.UUID) error

	// Operator
	DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error

	// Transaction
	ListTransactions(ctx context.Context, f AdminListTransactionsFilter) ([]*AdminTransaction, int, error)
	FindTransactionByID(ctx context.Context, id uuid.UUID) (*AdminTransaction, error)
	VoidTransaction(ctx context.Context, id uuid.UUID) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ----------------------------------------------------------------
// Auth
// ----------------------------------------------------------------

func (r *PostgresRepository) FindAdminByEmail(ctx context.Context, email string) (*AdminUser, error) {
	var a AdminUser
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, full_name, is_active, created_at, updated_at
		 FROM admin_users WHERE email = $1`,
		email,
	).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.FullName, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("admin: find by email: %w", err)
	}
	return &a, nil
}

func (r *PostgresRepository) FindAdminByID(ctx context.Context, id uuid.UUID) (*AdminUser, error) {
	var a AdminUser
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, full_name, is_active, created_at, updated_at
		 FROM admin_users WHERE id = $1`,
		id,
	).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.FullName, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAdminNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("admin: find by id: %w", err)
	}
	return &a, nil
}

func (r *PostgresRepository) StoreAdminRefreshToken(ctx context.Context, adminID uuid.UUID, tokenHash string, expiry time.Duration) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO admin_refresh_tokens (admin_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		adminID, tokenHash, time.Now().Add(expiry),
	)
	if err != nil {
		return fmt.Errorf("admin: store refresh token: %w", err)
	}
	return nil
}

func (r *PostgresRepository) FindAdminRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, time.Time, error) {
	var (
		adminID   uuid.UUID
		expiresAt time.Time
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT admin_id, expires_at FROM admin_refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&adminID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, time.Time{}, ErrRefreshNotFound
	}
	if err != nil {
		return uuid.Nil, time.Time{}, fmt.Errorf("admin: find refresh token: %w", err)
	}
	return adminID, expiresAt, nil
}

func (r *PostgresRepository) DeleteAdminRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM admin_refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	return err
}

// ----------------------------------------------------------------
// Business
// ----------------------------------------------------------------

const businessCols = `id, qios_id, user_id, business_name, phone, address, city, country, merchant_status, qris_string, created_at, updated_at`

func scanBusiness(row interface{ Scan(...any) error }) (*Business, error) {
	var b Business
	var phone, address, city, country, qrisString sql.NullString
	err := row.Scan(
		&b.ID, &b.QiosID, &b.UserID, &b.BusinessName,
		&phone, &address, &city, &country,
		&b.MerchantStatus, &qrisString,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if phone.Valid {
		b.Phone = &phone.String
	}
	if address.Valid {
		b.Address = &address.String
	}
	if city.Valid {
		b.City = &city.String
	}
	if country.Valid {
		b.Country = &country.String
	}
	if qrisString.Valid {
		b.QrisString = &qrisString.String
	}
	return &b, nil
}

func (r *PostgresRepository) ListBusinesses(ctx context.Context) ([]*Business, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+businessCols+` FROM businesses ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("admin: list businesses: %w", err)
	}
	defer rows.Close()

	var out []*Business
	for rows.Next() {
		b, err := scanBusiness(rows)
		if err != nil {
			return nil, fmt.Errorf("admin: scan business: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) FindBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error) {
	b, err := scanBusiness(r.db.QueryRowContext(ctx,
		`SELECT `+businessCols+` FROM businesses WHERE id = $1`, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBusinessNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("admin: find business: %w", err)
	}
	return b, nil
}

// CreateBusiness membuat user owner + business secara atomik.
// qios_id di-generate dari sequence MAX + 1 di dalam transaksi DB.
func (r *PostgresRepository) CreateBusiness(ctx context.Context, req CreateBusinessRequest) (*Business, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("admin: hash password: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("admin: begin tx: %w", err)
	}
	defer tx.Rollback()

	var userID uuid.UUID
	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, full_name, is_active)
		 VALUES ($1, $2, $3, TRUE)
		 RETURNING id`,
		strings.ToLower(strings.TrimSpace(req.Email)), string(hash), req.FullName,
	).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("admin: create user: %w", err)
	}

	var nextSeq int
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(CAST(SUBSTRING(qios_id FROM 6) AS INTEGER)), 0) + 1 FROM businesses`,
	).Scan(&nextSeq); err != nil {
		nextSeq = 1
	}
	qiosID := fmt.Sprintf("QIOS-%06d", nextSeq)

	b := &Business{
		UserID:       userID,
		QiosID:       qiosID,
		BusinessName: req.BusinessName,
		Phone:        req.Phone,
		Address:      req.Address,
		City:         req.City,
		Country:      req.Country,
		MerchantStatus: "PENDING",
	}

	err = tx.QueryRowContext(ctx,
		`INSERT INTO businesses (qios_id, user_id, business_name, phone, address, city, country, merchant_status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'PENDING')
		 RETURNING id, created_at, updated_at`,
		b.QiosID, b.UserID, b.BusinessName, b.Phone, b.Address, b.City, b.Country,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("admin: create business: %w", err)
	}

	return b, tx.Commit()
}

func (r *PostgresRepository) UpdateBusiness(ctx context.Context, b *Business) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE businesses
		 SET business_name=$1, phone=$2, address=$3, city=$4, country=$5, merchant_status=$6, updated_at=NOW()
		 WHERE id=$7`,
		b.BusinessName, b.Phone, b.Address, b.City, b.Country, b.MerchantStatus, b.ID,
	)
	if err != nil {
		return fmt.Errorf("admin: update business: %w", err)
	}
	return nil
}

// ----------------------------------------------------------------
// Product
// ----------------------------------------------------------------

const adminProductCols = `id, business_id, name, price, category, description, is_available, total_sold, created_at, updated_at`

func scanAdminProduct(row interface{ Scan(...any) error }) (*AdminProduct, error) {
	var p AdminProduct
	var category, description sql.NullString
	err := row.Scan(
		&p.ID, &p.BusinessID, &p.Name, &p.Price,
		&category, &description, &p.IsAvailable, &p.TotalSold,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if category.Valid {
		p.Category = &category.String
	}
	if description.Valid {
		p.Description = &description.String
	}
	return &p, nil
}

func (r *PostgresRepository) ListProductsByBusiness(ctx context.Context, businessID uuid.UUID) ([]*AdminProduct, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+adminProductCols+`
		 FROM products
		 WHERE business_id = $1 AND deleted_at IS NULL
		 ORDER BY name ASC`, businessID,
	)
	if err != nil {
		return nil, fmt.Errorf("admin: list products: %w", err)
	}
	defer rows.Close()

	var out []*AdminProduct
	for rows.Next() {
		p, err := scanAdminProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("admin: scan product: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) FindProductByID(ctx context.Context, id uuid.UUID) (*AdminProduct, error) {
	p, err := scanAdminProduct(r.db.QueryRowContext(ctx,
		`SELECT `+adminProductCols+`
		 FROM products WHERE id = $1 AND deleted_at IS NULL`, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("admin: find product: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) CreateProduct(ctx context.Context, businessID uuid.UUID, req AdminCreateProductRequest) (*AdminProduct, error) {
	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	p := &AdminProduct{
		BusinessID:  businessID,
		Name:        req.Name,
		Price:       req.Price,
		Category:    req.Category,
		Description: req.Description,
		IsAvailable: isAvailable,
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO products (business_id, name, price, category, description, is_available)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, total_sold, created_at, updated_at`,
		p.BusinessID, p.Name, p.Price, p.Category, p.Description, p.IsAvailable,
	).Scan(&p.ID, &p.TotalSold, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("admin: create product: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) UpdateProduct(ctx context.Context, p *AdminProduct) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE products
		 SET name=$1, price=$2, category=$3, description=$4, is_available=$5, updated_at=NOW()
		 WHERE id=$6 AND deleted_at IS NULL`,
		p.Name, p.Price, p.Category, p.Description, p.IsAvailable, p.ID,
	)
	if err != nil {
		return fmt.Errorf("admin: update product: %w", err)
	}
	return nil
}

func (r *PostgresRepository) SoftDeleteProduct(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("admin: delete product: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrProductNotFound
	}
	return nil
}

// ----------------------------------------------------------------
// Operator
// ----------------------------------------------------------------

func (r *PostgresRepository) DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE operators SET deleted_at=NOW()
		 WHERE id=$1 AND business_id=$2 AND deleted_at IS NULL`,
		operatorID, businessID,
	)
	if err != nil {
		return fmt.Errorf("admin: delete operator: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrOperatorNotFound
	}
	return nil
}

// ----------------------------------------------------------------
// Transaction
// ----------------------------------------------------------------

func scanAdminTransaction(row interface{ Scan(...any) error }) (*AdminTransaction, error) {
	var t AdminTransaction
	var operatorID uuid.NullUUID
	var paymentMethod, note sql.NullString
	var paidAt sql.NullTime
	err := row.Scan(
		&t.ID, &t.BusinessID, &operatorID, &t.OrderID,
		&t.TotalAmount, &paymentMethod, &t.Status, &note, &paidAt,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if operatorID.Valid {
		t.OperatorID = &operatorID.UUID
	}
	if paymentMethod.Valid {
		t.PaymentMethod = &paymentMethod.String
	}
	if note.Valid {
		t.Note = &note.String
	}
	if paidAt.Valid {
		t.PaidAt = &paidAt.Time
	}
	return &t, nil
}

func (r *PostgresRepository) ListTransactions(ctx context.Context, f AdminListTransactionsFilter) ([]*AdminTransaction, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	args := []any{}
	where := `WHERE 1=1`
	idx := 1

	if f.BusinessID != nil {
		where += fmt.Sprintf(` AND business_id = $%d`, idx)
		args = append(args, *f.BusinessID)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, f.Status)
		idx++
	}

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pos_orders `+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("admin: count transactions: %w", err)
	}

	offset := (f.Page - 1) * f.Limit
	listArgs := append(args, f.Limit, offset)

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, business_id, operator_id, order_id, total_amount, payment_method, status, note, paid_at, created_at, updated_at
		 FROM pos_orders `+where+
			fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1),
		listArgs...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("admin: list transactions: %w", err)
	}
	defer rows.Close()

	var out []*AdminTransaction
	for rows.Next() {
		t, err := scanAdminTransaction(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("admin: scan transaction: %w", err)
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepository) FindTransactionByID(ctx context.Context, id uuid.UUID) (*AdminTransaction, error) {
	t, err := scanAdminTransaction(r.db.QueryRowContext(ctx,
		`SELECT id, business_id, operator_id, order_id, total_amount, payment_method, status, note, paid_at, created_at, updated_at
		 FROM pos_orders WHERE id = $1`, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTransactionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("admin: find transaction: %w", err)
	}
	return t, nil
}

func (r *PostgresRepository) VoidTransaction(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE pos_orders SET status='cancelled', updated_at=NOW()
		 WHERE id=$1 AND status='pending'`, id,
	)
	if err != nil {
		return fmt.Errorf("admin: void transaction: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		var exists bool
		_ = r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pos_orders WHERE id=$1)`, id).Scan(&exists)
		if !exists {
			return ErrTransactionNotFound
		}
		return ErrTransactionNotPending
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
