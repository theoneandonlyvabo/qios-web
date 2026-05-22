package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserRoleByID(ctx context.Context, userID string) (businessID, role string, err error)

	StoreRefreshToken(ctx context.Context, userID, tokenHash string, expiry time.Duration) error
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	FindRefreshToken(ctx context.Context, tokenHash string) (userID string, expiresAt time.Time, err error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

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

func isUniqueViolation(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
