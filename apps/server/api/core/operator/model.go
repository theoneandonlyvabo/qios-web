// domain/operator/model.go
//
// Tipe-tipe untuk domain operator:
//   - Operator              → representasi baris di tabel operators
//   - CreateOperatorRequest → input owner saat tambah operator baru
//   - UpdateOperatorRequest → input owner saat update (partial, pointer fields)
//   - OperatorLoginRequest  → input login operator via operator_code + password
//   - QRLoginRequest        → input login operator via QR scan
//   - OperatorResponse      → bentuk yang dikirim ke client (tanpa secret)
//   - OperatorWithSecret    → bentuk khusus saat create/regenerate QR (sekali kembali)

package operator

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound dipakai service dan handler untuk distinguish 404 dari 500.
var ErrNotFound = errors.New("operator not found")

// ErrCodeTaken dipakai saat operator_code sudah dipakai di bisnis yang sama.
var ErrCodeTaken = errors.New("operator code already taken")

// ErrLimitReached dipakai saat slot operator sudah penuh sesuai plan.
var ErrLimitReached = errors.New("operator limit reached for current plan")

// ErrInvalidCredentials dipakai saat operator_code atau password salah.
var ErrInvalidCredentials = errors.New("invalid operator credentials")

// ErrInactive dipakai saat operator sudah dinonaktifkan.
var ErrInactive = errors.New("operator is inactive")

// Operator merepresentasikan baris di tabel operators.
type Operator struct {
	ID           uuid.UUID  `db:"id"            json:"id"`
	BusinessID   uuid.UUID  `db:"business_id"   json:"business_id"`
	Name         string     `db:"name"          json:"name"`
	OperatorCode string     `db:"operator_code" json:"operator_code"`
	PasswordHash string     `db:"password_hash" json:"-"`
	QRToken      string     `db:"qr_token"      json:"-"`
	IsActive     bool       `db:"is_active"     json:"is_active"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"    json:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"    json:"-"`
}

// CreateOperatorRequest — body POST /business/operators.
type CreateOperatorRequest struct {
	Name         string `json:"name"          validate:"required,min=1,max=255"`
	OperatorCode string `json:"operator_code" validate:"required,min=3,max=64"`
	Password     string `json:"password"      validate:"required,min=6,max=128"`
}

// UpdateOperatorRequest — body PUT /business/operators/:id.
// Pointer untuk membedakan field yang dikirim kosong vs tidak dikirim.
type UpdateOperatorRequest struct {
	Name     *string `json:"name"      validate:"omitempty,min=1,max=255"`
	IsActive *bool   `json:"is_active"`
}

// OperatorLoginRequest — body POST /kasir/auth/login.
type OperatorLoginRequest struct {
	OperatorCode string `json:"operator_code" validate:"required"`
	Password     string `json:"password"      validate:"required"`
}

// QRLoginRequest — body POST /kasir/auth/login/qr.
type QRLoginRequest struct {
	QRToken string `json:"qr_token" validate:"required"`
}

// OperatorResponse — response GET / list / update.
// Tidak include password_hash dan qr_token.
type OperatorResponse struct {
	ID           uuid.UUID `json:"id"`
	BusinessID   uuid.UUID `json:"business_id"`
	Name         string    `json:"name"`
	OperatorCode string    `json:"operator_code"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OperatorWithSecret — response create dan regenerate QR.
// QRToken hanya dikembalikan satu kali pada saat create / regenerate.
type OperatorWithSecret struct {
	OperatorResponse
	QRToken string `json:"qr_token"`
}

// LoginResponse — response sukses login operator.
type LoginResponse struct {
	AccessToken string           `json:"access_token"`
	Operator    OperatorResponse `json:"operator"`
}

// ToResponse memetakan Operator ke OperatorResponse.
func (o *Operator) ToResponse() OperatorResponse {
	return OperatorResponse{
		ID:           o.ID,
		BusinessID:   o.BusinessID,
		Name:         o.Name,
		OperatorCode: o.OperatorCode,
		IsActive:     o.IsActive,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}
}
