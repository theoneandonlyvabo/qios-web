package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountInactive    = errors.New("account is inactive or suspended")
	ErrGoogleOnlyAccount  = errors.New("account uses google login")
	ErrEmailTaken         = errors.New("email already registered")
	ErrQiosIDCollision    = errors.New("qios_id collision")
	ErrRefreshNotFound    = errors.New("refresh token not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrOperatorInactive   = errors.New("operator is inactive")
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	IsActive     bool
	IsSuspended  bool
	BusinessID   string
}

type LoginResult struct {
	AccessToken   string
	RefreshToken  string
	RefreshExpiry time.Duration
}

type RefreshResult struct {
	AccessToken   string
	RefreshToken  string
	RefreshExpiry time.Duration
}

// OperatorLoginData — minimal operator data needed during login flow.
type OperatorLoginData struct {
	ID           uuid.UUID
	BusinessID   uuid.UUID
	PasswordHash string
	IsActive     bool
	Name         string
	OperatorCode string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// OperatorInfo — operator data returned to client after login.
// OperatorCode tidak disertakan di sini — itu credential, tidak boleh balik ke client via response body.
type OperatorInfo struct {
	ID         uuid.UUID `json:"id"`
	BusinessID uuid.UUID `json:"business_id"`
	Name       string    `json:"name"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// OperatorLoginResult — response for operator login endpoints.
type OperatorLoginResult struct {
	AccessToken string       `json:"access_token"`
	Operator    OperatorInfo `json:"operator"`
}

const (
	roleOwner    = "owner"
	roleOperator = "operator"
)
