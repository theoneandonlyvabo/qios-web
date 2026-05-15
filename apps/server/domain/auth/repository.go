// domain/auth/repository.go
//
// Layer akses data untuk domain auth.
// Semua SQL di sini — service dan handler tidak boleh menyentuh DB langsung.
//
// Operasi yang berjalan di dalam transaksi register flow menerima *sql.Tx
// secara eksplisit. Operasi standalone (refresh token, login lookup) pakai
// *sql.DB pool.

package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Repository mendefinisikan kontrak akses data auth.
type Repository interface {
	// Login / refresh lookups (non-transactional).
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserRoleByID(ctx context.Context, userID string) (businessID, role string, err error)

	// Transactional register flow.
	CreateUser(ctx context.Context, tx *sql.Tx, email, passwordHash, fullName, phone string) (userID string, err error)
	CreateBusiness(ctx context.Context, tx *sql.Tx, qiosID, userID, businessName, phone, address, city, country string) (businessID string, err error)
	UpdateBusinessXendit(ctx context.Context, tx *sql.Tx, businessID, xenditAccountID, xenditStatus string) error

	// Refresh token CRUD (non-transactional — single-row writes pakai pool).
	StoreRefreshToken(ctx context.Context, userID, tokenHash string, expiry time.Duration) error
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	FindRefreshToken(ctx context.Context, tokenHash string) (userID string, expiresAt time.Time, err error)
}

// PostgresRepository adalah implementasi Repository di atas database/sql + lib/pq.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ----------------------------------------------------------------
// Login / refresh lookups
// ----------------------------------------------------------------

// FindUserByEmail mengambil user beserta business_id (kalau ada).
// Mengembalikan ErrInvalidCredentials saat email tidak terdaftar agar
// handler tidak perlu membedakan email-tidak-ada dengan password-salah.
func (r *PostgresRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var (
		u          User
		businessID sql.NullString
		passHash   sql.NullString
		phone      sql.NullString
		fullName   sql.NullString
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT u.id, u.email, u.password_hash, u.full_name, u.phone,
		        u.is_active, u.is_suspended, b.id
		 FROM users u
		 LEFT JOIN businesses b ON b.user_id = u.id
		 WHERE u.email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &passHash, &fullName, &phone,
		&u.IsActive, &u.IsSuspended, &businessID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("auth: find user by email: %w", err)
	}
	if passHash.Valid {
		u.PasswordHash = passHash.String
	}
	if fullName.Valid {
		u.FullName = fullName.String
	}
	if phone.Valid {
		u.Phone = phone.String
	}
	if businessID.Valid {
		u.BusinessID = businessID.String
	}
	return &u, nil
}

// FindUserRoleByID mendeteksi apakah id milik owner (users + businesses)
// atau operator (operators). Dipakai saat refresh untuk menerbitkan
// access token sesuai role.
func (r *PostgresRepository) FindUserRoleByID(ctx context.Context, userID string) (string, string, error) {
	var businessID string

	err := r.db.QueryRowContext(ctx,
		`SELECT b.id FROM users u
		 JOIN businesses b ON b.user_id = u.id
		 WHERE u.id = $1`,
		userID,
	).Scan(&businessID)
	if err == nil {
		return businessID, roleOwner, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", fmt.Errorf("auth: find owner role: %w", err)
	}

	err = r.db.QueryRowContext(ctx,
		`SELECT business_id FROM operators WHERE id = $1`,
		userID,
	).Scan(&businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrInvalidCredentials
	}
	if err != nil {
		return "", "", fmt.Errorf("auth: find operator role: %w", err)
	}
	return businessID, roleOperator, nil
}

// ----------------------------------------------------------------
// Transactional register flow
// ----------------------------------------------------------------

// CreateUser menyisipkan baris baru di tabel users. Mengembalikan ErrEmailTaken
// kalau email sudah dipakai.
func (r *PostgresRepository) CreateUser(ctx context.Context, tx *sql.Tx, email, passwordHash, fullName, phone string) (string, error) {
	var userID string
	err := tx.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, full_name, phone)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		email, passwordHash, fullName, nullableString(phone),
	).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			return "", ErrEmailTaken
		}
		return "", fmt.Errorf("auth: insert user: %w", err)
	}
	return userID, nil
}

// CreateBusiness menyisipkan baris baru di tabel businesses dengan
// xendit_status = 'PENDING'. qios_id sudah pre-generated oleh caller
// di dalam tx yang sama (lihat platform/qmid.Generate).
func (r *PostgresRepository) CreateBusiness(ctx context.Context, tx *sql.Tx, qiosID, userID, businessName, phone, address, city, country string) (string, error) {
	var businessID string
	err := tx.QueryRowContext(ctx,
		`INSERT INTO businesses (
			qios_id, user_id, business_name, phone, address, city, country, xendit_status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'PENDING')
		RETURNING id`,
		qiosID, userID, businessName,
		nullableString(phone), nullableString(address),
		nullableString(city), nullableString(country),
	).Scan(&businessID)
	if err != nil {
		if isUniqueViolation(err) {
			return "", ErrQiosIDCollision
		}
		return "", fmt.Errorf("auth: insert business: %w", err)
	}
	return businessID, nil
}

// UpdateBusinessXendit set xendit_account_id dan xendit_status setelah call
// ke Xendit MANAGED account API sukses.
func (r *PostgresRepository) UpdateBusinessXendit(ctx context.Context, tx *sql.Tx, businessID, xenditAccountID, xenditStatus string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE businesses
		 SET xendit_account_id = $1,
		     xendit_status     = $2,
		     updated_at        = NOW()
		 WHERE id = $3`,
		xenditAccountID, xenditStatus, businessID,
	)
	if err != nil {
		return fmt.Errorf("auth: update business xendit: %w", err)
	}
	return nil
}

// ----------------------------------------------------------------
// Refresh token CRUD
// ----------------------------------------------------------------

// StoreRefreshToken menyimpan hash refresh token + expiry ke tabel refresh_tokens.
func (r *PostgresRepository) StoreRefreshToken(ctx context.Context, userID, tokenHash string, expiry time.Duration) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, time.Now().Add(expiry),
	)
	if err != nil {
		return fmt.Errorf("auth: store refresh token: %w", err)
	}
	return nil
}

// DeleteRefreshToken menghapus refresh token by hash. Tidak error kalau row tidak ada.
func (r *PostgresRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("auth: delete refresh token: %w", err)
	}
	return nil
}

// FindRefreshToken mengambil user_id dan expires_at untuk hash yang diberikan.
// Mengembalikan ErrRefreshNotFound kalau tidak ada.
func (r *PostgresRepository) FindRefreshToken(ctx context.Context, tokenHash string) (string, time.Time, error) {
	var (
		userID    string
		expiresAt time.Time
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", time.Time{}, ErrRefreshNotFound
	}
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth: find refresh token: %w", err)
	}
	return userID, expiresAt, nil
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

// nullableString mengembalikan sql.NullString supaya kolom nullable di DB
// menyimpan NULL kalau input string kosong, bukan empty string.
func nullableString(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// isUniqueViolation cek apakah error berasal dari unique constraint Postgres (23505).
func isUniqueViolation(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
