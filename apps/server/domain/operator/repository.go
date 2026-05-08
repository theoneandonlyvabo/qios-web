// domain/operator/repository.go
//
// Layer akses data untuk domain operator.
// Semua interaksi langsung ke tabel operators dilakukan di sini —
// service dan handler tidak boleh menyentuh database langsung.

package operator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Repository mendefinisikan kontrak akses data operator.
type Repository interface {
	Create(ctx context.Context, op *Operator) error
	FindByID(ctx context.Context, id uuid.UUID) (*Operator, error)
	FindByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*Operator, error)
	FindByOperatorCode(ctx context.Context, businessID uuid.UUID, code string) (*Operator, error)
	FindByQRToken(ctx context.Context, token string) (*Operator, error)
	CountActiveByBusinessID(ctx context.Context, businessID uuid.UUID) (int, error)
	Update(ctx context.Context, op *Operator) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	RegenerateQR(ctx context.Context, id uuid.UUID, newToken string) error
}

// PostgresRepository adalah implementasi Repository di atas database/sql + lib/pq.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

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

// Create menyisipkan operator baru.
// Mengembalikan ErrCodeTaken kalau operator_code sudah dipakai pada bisnis yang sama,
// atau kalau qr_token bentrok secara global (sangat jarang terjadi).
func (r *PostgresRepository) Create(ctx context.Context, op *Operator) error {
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
		return fmt.Errorf("operator: create: %w", err)
	}
	return nil
}

// FindByID mengambil operator by id (termasuk yang inactive, tapi exclude soft-deleted).
func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*Operator, error) {
	query := `SELECT ` + operatorColumns + `
		FROM operators
		WHERE id = $1 AND deleted_at IS NULL`

	op, err := scanOperator(r.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("operator: find by id: %w", err)
	}
	return op, nil
}

// FindByBusinessID mengembalikan semua operator non-deleted milik bisnis,
// urut ascending by created_at.
func (r *PostgresRepository) FindByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*Operator, error) {
	query := `SELECT ` + operatorColumns + `
		FROM operators
		WHERE business_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("operator: find by business: %w", err)
	}
	defer rows.Close()

	operators := make([]*Operator, 0)
	for rows.Next() {
		op, err := scanOperator(rows)
		if err != nil {
			return nil, fmt.Errorf("operator: scan: %w", err)
		}
		operators = append(operators, op)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("operator: rows: %w", err)
	}
	return operators, nil
}

// FindByOperatorCode dipakai handler login — operator_code unik per bisnis.
func (r *PostgresRepository) FindByOperatorCode(ctx context.Context, businessID uuid.UUID, code string) (*Operator, error) {
	query := `SELECT ` + operatorColumns + `
		FROM operators
		WHERE business_id = $1 AND operator_code = $2 AND deleted_at IS NULL`

	op, err := scanOperator(r.db.QueryRowContext(ctx, query, businessID, code))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("operator: find by code: %w", err)
	}
	return op, nil
}

// FindByQRToken dipakai handler QR login — qr_token unik global.
func (r *PostgresRepository) FindByQRToken(ctx context.Context, token string) (*Operator, error) {
	query := `SELECT ` + operatorColumns + `
		FROM operators
		WHERE qr_token = $1 AND deleted_at IS NULL`

	op, err := scanOperator(r.db.QueryRowContext(ctx, query, token))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("operator: find by qr: %w", err)
	}
	return op, nil
}

// CountActiveByBusinessID menghitung jumlah operator aktif (non-deleted) per bisnis,
// dipakai untuk enforcement slot cap dari plan.
func (r *PostgresRepository) CountActiveByBusinessID(ctx context.Context, businessID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM operators
		 WHERE business_id = $1 AND deleted_at IS NULL`,
		businessID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("operator: count: %w", err)
	}
	return count, nil
}

// Update melakukan partial update. Hanya field non-nil yang berubah.
// Kembalikan ErrNotFound kalau row tidak ada / sudah dihapus.
func (r *PostgresRepository) Update(ctx context.Context, op *Operator) error {
	query := `UPDATE operators
	          SET name       = $1,
	              is_active  = $2,
	              updated_at = NOW()
	          WHERE id = $3 AND deleted_at IS NULL
	          RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query, op.Name, op.IsActive, op.ID).Scan(&op.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("operator: update: %w", err)
	}
	return nil
}

// SoftDelete set deleted_at = NOW(). Tidak menghapus data.
func (r *PostgresRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE operators
		 SET deleted_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("operator: soft delete: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// RegenerateQR mengganti qr_token operator. Sesi JWT yang aktif tidak terpengaruh
// karena JWT mengikat ke operator id, bukan ke qr_token.
func (r *PostgresRepository) RegenerateQR(ctx context.Context, id uuid.UUID, newToken string) error {
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
		return fmt.Errorf("operator: regenerate qr: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// isUniqueViolation memeriksa Postgres error code 23505 (unique_violation).
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	// Fallback parse string — beberapa driver tidak expose pq.Error langsung.
	return strings.Contains(err.Error(), "duplicate key value")
}

// isUndefinedTable memeriksa Postgres error code 42P01 (undefined_table).
// Dipakai sebagai fallback saat tabel plans/subscriptions belum dimigrasi.
func isUndefinedTable(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "42P01"
	}
	msg := err.Error()
	return strings.Contains(msg, "does not exist")
}
