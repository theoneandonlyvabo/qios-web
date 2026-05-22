package auth

import (
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountInactive    = errors.New("account is inactive or suspended")
	ErrGoogleOnlyAccount  = errors.New("account uses google login")
	ErrEmailTaken         = errors.New("email already registered")
	ErrQiosIDCollision    = errors.New("qios_id collision")
	ErrRefreshNotFound    = errors.New("refresh token not found")
	ErrSessionExpired     = errors.New("session expired")
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

const (
	roleOwner    = "owner"
	roleOperator = "operator"
)
