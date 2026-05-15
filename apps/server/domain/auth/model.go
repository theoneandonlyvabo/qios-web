// domain/auth/model.go
//
// Tipe domain untuk auth:
//   - User                 → representasi user yang dipakai login flow
//   - RegisterInput        → input service Register
//   - RegisterResult       → output service Register
//   - LoginResult          → output service Login
//   - RefreshResult        → output service Refresh
//   - Sentinel errors      → distinguish 401 / 403 / 409 / 422 / 500 di handler
//   - xenditCreator        → port kecil untuk dependency Xendit (di-mock di test)

package auth

import (
	"context"
	"errors"
	"time"

	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/payment"
)

// ----------------------------------------------------------------
// Sentinel errors
// ----------------------------------------------------------------

// ErrInvalidCredentials dipakai service Login saat email atau password salah.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrAccountInactive dipakai saat user is_active=false atau is_suspended=true.
var ErrAccountInactive = errors.New("account is inactive or suspended")

// ErrGoogleOnlyAccount dipakai saat user mencoba login dengan password
// tapi akun-nya hanya punya kredensial Google OAuth.
var ErrGoogleOnlyAccount = errors.New("account uses google login")

// ErrEmailTaken dipakai service Register saat email sudah ada di tabel users.
var ErrEmailTaken = errors.New("email already registered")

// ErrQiosIDCollision dipakai service Register kalau qios_id bentrok.
// Sangat jarang — biasanya bisa retry langsung dari client.
var ErrQiosIDCollision = errors.New("qios_id collision")

// ErrXenditCreate dipakai saat call Xendit gagal di tengah register flow.
// Handler memetakannya ke 422 Unprocessable Entity.
var ErrXenditCreate = errors.New("xendit create account failed")

// ErrRefreshNotFound dipakai saat refresh token tidak ada di DB.
var ErrRefreshNotFound = errors.New("refresh token not found")

// ErrSessionExpired dipakai saat refresh token sudah lewat expires_at.
var ErrSessionExpired = errors.New("session expired")

// ----------------------------------------------------------------
// Domain types
// ----------------------------------------------------------------

// User merepresentasikan baris di tabel users beserta business_id terkait.
// PasswordHash dan business_id dipakai service untuk verifikasi login.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	IsActive     bool
	IsSuspended  bool
	BusinessID   string // dari LEFT JOIN businesses — kosong kalau belum punya bisnis.
}

// RegisterInput adalah payload service.Register, sudah ter-normalize.
type RegisterInput struct {
	Email        string
	Password     string
	FullName     string
	Phone        string
	BusinessName string
	Address      string
	City         string
	Country      string
}

// RegisterResult adalah output service.Register.
// RefreshToken adalah token plaintext yang dipasang handler ke cookie.
type RegisterResult struct {
	UserID         string
	BusinessID     string
	QiosID         string
	XenditStatus   string
	AccessToken    string
	RefreshToken   string
	RefreshExpiry  time.Duration
}

// LoginResult adalah output service.Login.
type LoginResult struct {
	AccessToken   string
	RefreshToken  string
	RefreshExpiry time.Duration
}

// RefreshResult adalah output service.Refresh.
type RefreshResult struct {
	AccessToken   string
	RefreshToken  string
	RefreshExpiry time.Duration
}

// ----------------------------------------------------------------
// External ports
// ----------------------------------------------------------------

// xenditCreator memungkinkan service di-test tanpa hit jaringan.
// Production-nya dipenuhi oleh *payment.XenditService.
type xenditCreator interface {
	CreateSubAccount(ctx context.Context, in payment.ManagedAccountInput) (*payment.ManagedAccountResult, error)
}

// Role constants — keep handler logic readable.
const (
	roleOwner    = "owner"
	roleOperator = "operator"
)
