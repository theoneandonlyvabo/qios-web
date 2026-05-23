// domain/user/repository.go
//
// Layer akses data untuk domain user.
// Menggabungkan akses ke tabel users, businesses, dan operators.
// Service dan handler tidak boleh menyentuh database langsung.

package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ----------------------------------------------------------------
// Repository interfaces
// ----------------------------------------------------------------

// Repository mendefinisikan kontrak akses data domain user.
type Repository interface {
	// User profile
	FindProfile(ctx context.Context, userID string) (*UserProfile, error)
	UpdateProfile(ctx context.Context, userID string, fullName string, phone *string) error

	// Business
	FindBusiness(ctx context.Context, businessID string) (*BusinessInfo, error)
	UpdateBusiness(ctx context.Context, businessID string, req UpdateBusinessRequest) error

	// Operator CRUD
	CreateOperator(ctx context.Context, op *Operator) error
	FindOperatorByID(ctx context.Context, id uuid.UUID) (*Operator, error)
	FindOperatorsByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*Operator, error)
	CountActiveOperators(ctx context.Context, businessID uuid.UUID) (int, error)
	UpdateOperator(ctx context.Context, op *Operator) error
	SoftDeleteOperator(ctx context.Context, id uuid.UUID) error
	RegenerateOperatorQR(ctx context.Context, id uuid.UUID, newToken string) error
}

// PlanLookup reads max_operators slot from active plan.
type PlanLookup interface {
	MaxOperators(ctx context.Context, businessID uuid.UUID) (int, error)
}

// ----------------------------------------------------------------
// PostgresRepository
// ----------------------------------------------------------------

// PostgresRepository implementasi Repository di atas database/sql.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ----------------------------------------------------------------
// User profile
// ----------------------------------------------------------------

func (r *PostgresRepository) FindProfile(ctx context.Context, userID string) (*UserProfile, error) {
	var p UserProfile
	var bizID, qiosID, bizName, xenditStatus sql.NullString
	var bizPhone, address, city, country sql.NullString
	var phone sql.NullString
	var fullName sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT u.id, u.email, u.full_name, u.phone,
		        b.id, b.qios_id, b.business_name, b.phone, b.address, b.city, b.country, b.xendit_status
		 FROM users u
		 LEFT JOIN businesses b ON b.user_id = u.id
		 WHERE u.id = $1`,
		userID,
	).Scan(
		&p.ID, &p.Email, &fullName, &phone,
		&bizID, &qiosID, &bizName, &bizPhone, &address, &city, &country, &xenditStatus,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user: find profile: %w", err)
	}

	if fullName.Valid {
		p.FullName = fullName.String
	}
	if phone.Valid {
		p.Phone = &phone.String
	}
	if bizID.Valid {
		p.BusinessID = bizID.String
	}
	if qiosID.Valid {
		p.QiosID = qiosID.String
	}
	if bizName.Valid {
		p.BusinessName = bizName.String
	}
	if bizPhone.Valid {
		p.BizPhone = &bizPhone.String
	}
	if address.Valid {
		p.Address = &address.String
	}
	if city.Valid {
		p.City = &city.String
	}
	if country.Valid {
		p.Country = &country.String
	}
	if xenditStatus.Valid {
		p.XenditStatus = xenditStatus.String
	}
	return &p, nil
}

func (r *PostgresRepository) UpdateProfile(ctx context.Context, userID string, fullName string, phone *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users
		 SET full_name  = COALESCE(NULLIF($1, ''), full_name),
		     phone      = COALESCE($2, phone),
		     updated_at = NOW()
		 WHERE id = $3`,
		fullName, phone, userID,
	)
	if err != nil {
		return fmt.Errorf("user: update profile: %w", err)
	}
	return nil
}

// ----------------------------------------------------------------
// Business
// ----------------------------------------------------------------

func (r *PostgresRepository) FindBusiness(ctx context.Context, businessID string) (*BusinessInfo, error) {
	var b BusinessInfo
	var phone, address, city, country, qrisString sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, qios_id, business_name, phone, address, city, country, xendit_status, qris_string
		 FROM businesses WHERE id = $1`,
		businessID,
	).Scan(
		&b.ID, &b.QiosID, &b.BusinessName, &phone,
		&address, &city, &country, &b.XenditStatus, &qrisString,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user: find business: %w", err)
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

func (r *PostgresRepository) UpdateBusiness(ctx context.Context, businessID string, req UpdateBusinessRequest) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE businesses
		 SET business_name = COALESCE(NULLIF($1, ''), business_name),
		     phone         = COALESCE(NULLIF($2, ''), phone),
		     address       = COALESCE(NULLIF($3, ''), address),
		     city          = COALESCE(NULLIF($4, ''), city),
		     country       = COALESCE(NULLIF($5, ''), country),
		     qris_string   = COALESCE($6, qris_string),
		     updated_at    = NOW()
		 WHERE id = $7`,
		req.BusinessName, req.Phone, req.Address, req.City, req.Country, req.QrisString, businessID,
	)
	if err != nil {
		return fmt.Errorf("user: update business: %w", err)
	}
	return nil
}

// ----------------------------------------------------------------
// Operator CRUD
// ----------------------------------------------------------------

const operatorColumns = `id, business_id, name, operator_code, password_hash,
       qr_token, is_active, created_at, updated_at, deleted_at`

func scanOperator(row interface{ Scan(...any) error }) (*Operator, error) {
	var op Operator
	var deletedAt sql.NullTime
	err := row.Scan(
		&op.ID, &op.BusinessID, &op.Name, &op.OperatorCode, &op.PasswordHash,
		&op.QRToken, &op.IsActive, &op.CreatedAt, &op.UpdatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}
	if deletedAt.Valid {
		op.DeletedAt = &deletedAt.Time
	}
	return &op, nil
}

func (r *PostgresRepository) CreateOperator(ctx context.Context, op *Operator) error {
	query := `
		INSERT INTO operators (business_id, name, operator_code, password_hash,
		                       qr_token, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		op.BusinessID, op.Name, op.OperatorCode, op.PasswordHash,
		op.QRToken, op.IsActive,
	).Scan(&op.ID, &op.CreatedAt, &op.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrCodeTaken
		}
		return fmt.Errorf("user: create operator: %w", err)
	}
	return nil
}

func (r *PostgresRepository) FindOperatorByID(ctx context.Context, id uuid.UUID) (*Operator, error) {
	op, err := scanOperator(r.db.QueryRowContext(ctx,
		`SELECT `+operatorColumns+` FROM operators WHERE id = $1 AND deleted_at IS NULL`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user: find operator by id: %w", err)
	}
	return op, nil
}

func (r *PostgresRepository) FindOperatorsByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*Operator, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+operatorColumns+` FROM operators
		 WHERE business_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at ASC`,
		businessID,
	)
	if err != nil {
		return nil, fmt.Errorf("user: list operators: %w", err)
	}
	defer rows.Close()

	var out []*Operator
	for rows.Next() {
		op, err := scanOperator(rows)
		if err != nil {
			return nil, fmt.Errorf("user: scan operator: %w", err)
		}
		out = append(out, op)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) CountActiveOperators(ctx context.Context, businessID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM operators
		 WHERE business_id = $1 AND deleted_at IS NULL`,
		businessID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("user: count operators: %w", err)
	}
	return count, nil
}

func (r *PostgresRepository) UpdateOperator(ctx context.Context, op *Operator) error {
	err := r.db.QueryRowContext(ctx,
		`UPDATE operators
		 SET name       = $1,
		     is_active  = $2,
		     updated_at = NOW()
		 WHERE id = $3 AND deleted_at IS NULL
		 RETURNING updated_at`,
		op.Name, op.IsActive, op.ID,
	).Scan(&op.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("user: update operator: %w", err)
	}
	return nil
}

func (r *PostgresRepository) SoftDeleteOperator(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE operators
		 SET deleted_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("user: soft delete operator: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) RegenerateOperatorQR(ctx context.Context, id uuid.UUID, newToken string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE operators
		 SET qr_token = $1, updated_at = NOW()
		 WHERE id = $2 AND deleted_at IS NULL`,
		newToken, id,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrCodeTaken
		}
		return fmt.Errorf("user: regenerate qr: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
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
	return strings.Contains(err.Error(), "duplicate key value")
}

// ----------------------------------------------------------------
// PostgresPlanLookup
// ----------------------------------------------------------------

// PostgresPlanLookup membaca max_operators dari subscription aktif.
// Fallback ke 3 kalau plan belum di-set atau tabel belum ada.
type PostgresPlanLookup struct {
	db *sql.DB
}

func NewPostgresPlanLookup(db *sql.DB) *PostgresPlanLookup {
	return &PostgresPlanLookup{db: db}
}

const defaultMaxOperators = 3

func (p *PostgresPlanLookup) MaxOperators(ctx context.Context, businessID uuid.UUID) (int, error) {
	var maxOps sql.NullInt64
	err := p.db.QueryRowContext(ctx,
		`SELECT pl.max_operators
		 FROM businesses b
		 JOIN subscriptions s ON s.user_id = b.user_id
		 JOIN plans pl        ON pl.id     = s.plan_id
		 WHERE b.id = $1
		   AND s.status = 'active'
		   AND (s.expires_at IS NULL OR s.expires_at > NOW())
		 ORDER BY s.started_at DESC
		 LIMIT 1`,
		businessID,
	).Scan(&maxOps)

	if errors.Is(err, sql.ErrNoRows) {
		return defaultMaxOperators, nil
	}
	if err != nil {
		if isUndefinedTable(err) {
			return defaultMaxOperators, nil
		}
		return 0, fmt.Errorf("user: plan lookup: %w", err)
	}
	if !maxOps.Valid {
		return defaultMaxOperators, nil
	}
	return int(maxOps.Int64), nil
}

func isUndefinedTable(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "42P01"
	}
	return strings.Contains(err.Error(), "does not exist")
}
