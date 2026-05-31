// core/admin/repository.go
//
// Layer akses data untuk domain admin.
// Semua operasi DB dilakukan di sini — service dan handler tidak menyentuh DB.

package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	// Owner
	ListOwners(ctx context.Context, page, limit int) ([]*OwnerSummary, int, error)
	FindOwnerByID(ctx context.Context, businessID uuid.UUID) (*OwnerDetail, error)
	SetOwnerStatus(ctx context.Context, businessID uuid.UUID, suspended bool) error
	SetOwnerCredential(ctx context.Context, businessID uuid.UUID, email *string, passwordHash string) error

	// Business (create/update tetap pakai business-centric)
	CreateBusiness(ctx context.Context, req CreateBusinessRequest) (*Business, error)
	FindBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error)
	UpdateBusiness(ctx context.Context, b *Business) error

	// Product
	ListProductsByBusiness(ctx context.Context, businessID uuid.UUID) ([]*AdminProduct, error)
	FindProductByID(ctx context.Context, id uuid.UUID) (*AdminProduct, error)
	FindProductDetailByID(ctx context.Context, id uuid.UUID) (*AdminProductDetail, error)
	CreateProduct(ctx context.Context, businessID uuid.UUID, req AdminCreateProductRequest) (*AdminProduct, error)
	UpdateProduct(ctx context.Context, p *AdminProduct) error
	UpdateProductRecipe(ctx context.Context, id uuid.UUID, recipe json.RawMessage) error
	SoftDeleteProduct(ctx context.Context, id uuid.UUID) error

	// Operator
	DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error

	// Transaction
	ListTransactions(ctx context.Context, f AdminListTransactionsFilter) ([]*AdminTransaction, int, error)
	FindTransactionByID(ctx context.Context, id uuid.UUID) (*AdminTransaction, error)
	VoidTransaction(ctx context.Context, id uuid.UUID) error

	// Audit
	WriteAuditLog(ctx context.Context, adminID uuid.UUID, targetType string, targetID *uuid.UUID, action string) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ----------------------------------------------------------------
// Owner
// ----------------------------------------------------------------

func (r *PostgresRepository) ListOwners(ctx context.Context, page, limit int) ([]*OwnerSummary, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM businesses`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("admin: count owners: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx,
		`SELECT b.id, u.id, b.qios_id, u.email, u.full_name,
		        b.business_name, b.merchant_status,
		        u.is_active, u.is_suspended, b.created_at
		 FROM businesses b JOIN users u ON u.id = b.user_id
		 ORDER BY b.created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("admin: list owners: %w", err)
	}
	defer rows.Close()

	var out []*OwnerSummary
	for rows.Next() {
		var o OwnerSummary
		if err := rows.Scan(
			&o.BusinessID, &o.UserID, &o.QiosID, &o.Email, &o.FullName,
			&o.BusinessName, &o.MerchantStatus,
			&o.IsActive, &o.IsSuspended, &o.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("admin: scan owner: %w", err)
		}
		out = append(out, &o)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepository) FindOwnerByID(ctx context.Context, businessID uuid.UUID) (*OwnerDetail, error) {
	var o OwnerDetail
	var phone, address, city, country, qrisString sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT b.id, u.id, b.qios_id, u.email, u.full_name,
		        b.business_name, b.phone, b.address, b.city, b.country,
		        b.qris_string, b.merchant_status,
		        u.is_active, u.is_suspended, b.created_at, b.updated_at
		 FROM businesses b JOIN users u ON u.id = b.user_id
		 WHERE b.id = $1`,
		businessID,
	).Scan(
		&o.BusinessID, &o.UserID, &o.QiosID, &o.Email, &o.FullName,
		&o.BusinessName, &phone, &address, &city, &country,
		&qrisString, &o.MerchantStatus,
		&o.IsActive, &o.IsSuspended, &o.CreatedAt, &o.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOwnerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("admin: find owner: %w", err)
	}
	if phone.Valid {
		o.Phone = &phone.String
	}
	if address.Valid {
		o.Address = &address.String
	}
	if city.Valid {
		o.City = &city.String
	}
	if country.Valid {
		o.Country = &country.String
	}
	if qrisString.Valid {
		o.QrisString = &qrisString.String
	}
	return &o, nil
}

// SetOwnerStatus update is_suspended di users dan merchant_status di businesses secara atomik.
func (r *PostgresRepository) SetOwnerStatus(ctx context.Context, businessID uuid.UUID, suspended bool) error {
	merchantStatus := "ACTIVE"
	if suspended {
		merchantStatus = "SUSPENDED"
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("admin: begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE users u SET is_suspended = $1, updated_at = NOW()
		 FROM businesses b WHERE b.id = $2 AND u.id = b.user_id`,
		suspended, businessID,
	)
	if err != nil {
		return fmt.Errorf("admin: update user status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrOwnerNotFound
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE businesses SET merchant_status = $1, updated_at = NOW() WHERE id = $2`,
		merchantStatus, businessID,
	); err != nil {
		return fmt.Errorf("admin: update business status: %w", err)
	}

	return tx.Commit()
}

// SetOwnerCredential update email (opsional) dan password_hash di tabel users.
func (r *PostgresRepository) SetOwnerCredential(ctx context.Context, businessID uuid.UUID, email *string, passwordHash string) error {
	var err error
	if email != nil {
		normalized := strings.ToLower(strings.TrimSpace(*email))
		_, err = r.db.ExecContext(ctx,
			`UPDATE users u SET email = $1, password_hash = $2, updated_at = NOW()
			 FROM businesses b WHERE b.id = $3 AND u.id = b.user_id`,
			normalized, passwordHash, businessID,
		)
	} else {
		_, err = r.db.ExecContext(ctx,
			`UPDATE users u SET password_hash = $1, updated_at = NOW()
			 FROM businesses b WHERE b.id = $2 AND u.id = b.user_id`,
			passwordHash, businessID,
		)
	}
	if err != nil {
		if isUniqueViolation(err) {
			return ErrEmailTaken
		}
		return fmt.Errorf("admin: set credential: %w", err)
	}
	return nil
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
	if err := tx.QueryRowContext(ctx, `SELECT nextval('qios_id_seq')`).Scan(&nextSeq); err != nil {
		return nil, fmt.Errorf("admin: generate qios_id: %w", err)
	}
	qiosID := fmt.Sprintf("QIOS-%06d", nextSeq)

	b := &Business{
		UserID:         userID,
		QiosID:         qiosID,
		BusinessName:   req.BusinessName,
		Phone:          req.Phone,
		Address:        req.Address,
		City:           req.City,
		Country:        req.Country,
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

func (r *PostgresRepository) FindProductDetailByID(ctx context.Context, id uuid.UUID) (*AdminProductDetail, error) {
	var p AdminProductDetail
	var category, description sql.NullString
	var recipeRaw []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT id, business_id, name, price, category, description,
		        is_available, recipe, created_at, updated_at
		 FROM products WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(
		&p.ID, &p.BusinessID, &p.Name, &p.Price,
		&category, &description, &p.IsAvailable,
		&recipeRaw, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("admin: find product detail: %w", err)
	}
	if category.Valid {
		p.Category = &category.String
	}
	if description.Valid {
		p.Description = &description.String
	}
	if len(recipeRaw) > 0 && string(recipeRaw) != "null" {
		_ = json.Unmarshal(recipeRaw, &p.Recipe)
	}
	if p.Recipe == nil {
		p.Recipe = []Ingredient{}
	}
	return &p, nil
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

func (r *PostgresRepository) UpdateProductRecipe(ctx context.Context, id uuid.UUID, recipe json.RawMessage) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET recipe = $1::jsonb, updated_at = NOW()
		 WHERE id = $2 AND deleted_at IS NULL`,
		[]byte(recipe), id,
	)
	if err != nil {
		return fmt.Errorf("admin: update recipe: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrProductNotFound
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
		`SELECT COUNT(*) FROM orders `+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("admin: count transactions: %w", err)
	}

	offset := (f.Page - 1) * f.Limit
	listArgs := append(args, f.Limit, offset)

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, business_id, operator_id, order_id, total_amount, payment_method, status, note, paid_at, created_at, updated_at
		 FROM orders `+where+
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
		 FROM orders WHERE id = $1`, id,
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
		`UPDATE orders SET status='VOIDED', updated_at=NOW()
		 WHERE id=$1 AND status='CONFIRMED'`, id,
	)
	if err != nil {
		return fmt.Errorf("admin: void transaction: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		var exists bool
		_ = r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE id=$1)`, id).Scan(&exists)
		if !exists {
			return ErrTransactionNotFound
		}
		return ErrTransactionNotPending
	}
	return nil
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

