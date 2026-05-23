// domain/user/model.go
//
// Tipe-tipe untuk domain user:
//   - Operator, OperatorResponse, OperatorWithSecret → ported dari domain/operator
//   - CreateOperatorRequest, UpdateOperatorRequest   → request types operator CRUD
//   - UserProfile, BusinessInfo                      → profile + business data
//   - UpdateUserRequest, UpdateBusinessRequest        → request types update

package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ----------------------------------------------------------------
// Errors
// ----------------------------------------------------------------

var (
	ErrNotFound    = errors.New("operator not found")
	ErrCodeTaken   = errors.New("operator code already taken")
	ErrLimitReached = errors.New("operator limit reached for current plan")
)

// ----------------------------------------------------------------
// Operator types (ported from operator domain)
// ----------------------------------------------------------------

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

// OperatorResponse — response GET / list / update. Tanpa secret fields.
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
// QRToken hanya dikembalikan satu kali.
type OperatorWithSecret struct {
	OperatorResponse
	QRToken string `json:"qr_token"`
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

// ----------------------------------------------------------------
// User + Business types
// ----------------------------------------------------------------

// UserProfile — data profil owner dari tabel users.
type UserProfile struct {
	ID           string
	Email        string
	FullName     string
	Phone        *string
	Role         string
	BusinessID   string
	QiosID       string
	BusinessName string
	BizPhone     *string
	Address      *string
	City         *string
	Country      *string
	XenditStatus string
}

// BusinessInfo — data bisnis dari tabel businesses.
type BusinessInfo struct {
	ID           string
	QiosID       string
	BusinessName string
	Phone        *string
	Address      *string
	City         *string
	Country      *string
	XenditStatus string
	QrisString   *string
}

// ----------------------------------------------------------------
// Handler response types (internal)
// ----------------------------------------------------------------

type businessInMe struct {
	ID           string  `json:"id"`
	QiosID       string  `json:"qios_id"`
	BusinessName string  `json:"business_name"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	City         *string `json:"city"`
	Country      *string `json:"country"`
	XenditStatus string  `json:"xendit_status"`
}

type meResponse struct {
	ID       string       `json:"id"`
	Email    string       `json:"email"`
	FullName string       `json:"full_name"`
	Phone    *string      `json:"phone"`
	Role     string       `json:"role"`
	Business businessInMe `json:"business"`
}

type businessResponse struct {
	ID           string  `json:"id"`
	QiosID       string  `json:"qios_id"`
	BusinessName string  `json:"business_name"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	City         *string `json:"city"`
	Country      *string `json:"country"`
	XenditStatus string  `json:"xendit_status"`
	QrisString   *string `json:"qris_string"`
}

// UpdateBusinessRequest — body PATCH /business.
type UpdateBusinessRequest struct {
	BusinessName string  `json:"business_name" validate:"omitempty,min=1,max=255"`
	Phone        string  `json:"phone"         validate:"omitempty,min=1,max=32"`
	Address      string  `json:"address"       validate:"omitempty,min=1,max=1024"`
	City         string  `json:"city"          validate:"omitempty,min=1,max=100"`
	Country      string  `json:"country"       validate:"omitempty,min=2,max=100"`
	QrisString   *string `json:"qris_string"   validate:"omitempty,min=1,max=4096"`
}
