// core/user/service.go
//
// Logika bisnis untuk domain user.
// Service tidak menyentuh database langsung — semua via Repository.
// Service tidak menyentuh HTTP — handler yang menerjemahkan ke response.
//
// User profile: GetMe, UpdateMe
// Business info: GetBusiness, UpdateBusiness
// Operator CRUD: CreateOperator, GetOperators, GetOperatorByID,
//                UpdateOperator, DeleteOperator, RegenerateQR, GetSlotInfo

package user

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service mendefinisikan kontrak business logic domain user.
type Service interface {
	// User profile
	GetMe(ctx context.Context, userID string) (*UserProfile, error)
	UpdateMe(ctx context.Context, userID string, fullName string, phone *string) error

	// Business
	GetBusiness(ctx context.Context, businessID string) (*BusinessInfo, error)
	UpdateBusiness(ctx context.Context, businessID string, req UpdateBusinessRequest) error

	// Operator CRUD
	CreateOperator(ctx context.Context, businessID uuid.UUID, req CreateOperatorRequest) (*OperatorWithSecret, error)
	GetOperators(ctx context.Context, businessID uuid.UUID) ([]OperatorResponse, error)
	GetOperatorByID(ctx context.Context, businessID, operatorID uuid.UUID) (*OperatorResponse, error)
	UpdateOperator(ctx context.Context, businessID, operatorID uuid.UUID, req UpdateOperatorRequest) (*OperatorResponse, error)
	DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error
	RegenerateQR(ctx context.Context, businessID, operatorID uuid.UUID) (*OperatorWithSecret, error)
	GetSlotInfo(ctx context.Context, businessID uuid.UUID) (used, max int, err error)
}

type service struct {
	repo Repository
	plan PlanLookup
}

// NewService merangkai Repository dan PlanLookup.
// Tidak membutuhkan jwtSvc — operator login sudah dipindah ke /auth domain.
func NewService(repo Repository, plan PlanLookup) Service {
	return &service{repo: repo, plan: plan}
}

// ----------------------------------------------------------------
// User profile
// ----------------------------------------------------------------

func (s *service) GetMe(ctx context.Context, userID string) (*UserProfile, error) {
	return s.repo.FindProfile(ctx, userID)
}

func (s *service) UpdateMe(ctx context.Context, userID string, fullName string, phone *string) error {
	return s.repo.UpdateProfile(ctx, userID, fullName, phone)
}

// ----------------------------------------------------------------
// Business
// ----------------------------------------------------------------

func (s *service) GetBusiness(ctx context.Context, businessID string) (*BusinessInfo, error) {
	return s.repo.FindBusiness(ctx, businessID)
}

func (s *service) UpdateBusiness(ctx context.Context, businessID string, req UpdateBusinessRequest) error {
	return s.repo.UpdateBusiness(ctx, businessID, req)
}

// ----------------------------------------------------------------
// Operator CRUD
// ----------------------------------------------------------------

func (s *service) CreateOperator(ctx context.Context, businessID uuid.UUID, req CreateOperatorRequest) (*OperatorWithSecret, error) {
	maxOps, err := s.plan.MaxOperators(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("user service: read plan: %w", err)
	}
	count, err := s.repo.CountActiveOperators(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("user service: count: %w", err)
	}
	if maxOps != -1 && count >= maxOps {
		return nil, ErrLimitReached
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("user service: hash password: %w", err)
	}

	qrToken, err := generateQRToken()
	if err != nil {
		return nil, fmt.Errorf("user service: generate qr: %w", err)
	}

	op := &Operator{
		BusinessID:   businessID,
		Name:         req.Name,
		OperatorCode: req.OperatorCode,
		PasswordHash: string(hash),
		QRToken:      qrToken,
		IsActive:     true,
	}

	if err := s.repo.CreateOperator(ctx, op); err != nil {
		return nil, err
	}

	return &OperatorWithSecret{
		OperatorResponse: op.ToResponse(),
		QRToken:          qrToken,
	}, nil
}

func (s *service) GetOperators(ctx context.Context, businessID uuid.UUID) ([]OperatorResponse, error) {
	ops, err := s.repo.FindOperatorsByBusinessID(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]OperatorResponse, 0, len(ops))
	for _, op := range ops {
		out = append(out, op.ToResponse())
	}
	return out, nil
}

func (s *service) GetOperatorByID(ctx context.Context, businessID, operatorID uuid.UUID) (*OperatorResponse, error) {
	op, err := s.repo.FindOperatorByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if op.BusinessID != businessID {
		return nil, ErrNotFound
	}
	res := op.ToResponse()
	return &res, nil
}

func (s *service) UpdateOperator(ctx context.Context, businessID, operatorID uuid.UUID, req UpdateOperatorRequest) (*OperatorResponse, error) {
	op, err := s.repo.FindOperatorByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if op.BusinessID != businessID {
		return nil, ErrNotFound
	}

	if req.Name != nil {
		op.Name = *req.Name
	}
	if req.IsActive != nil {
		op.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateOperator(ctx, op); err != nil {
		return nil, err
	}
	res := op.ToResponse()
	return &res, nil
}

func (s *service) DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error {
	op, err := s.repo.FindOperatorByID(ctx, operatorID)
	if err != nil {
		return err
	}
	if op.BusinessID != businessID {
		return ErrNotFound
	}
	return s.repo.SoftDeleteOperator(ctx, operatorID)
}

func (s *service) RegenerateQR(ctx context.Context, businessID, operatorID uuid.UUID) (*OperatorWithSecret, error) {
	op, err := s.repo.FindOperatorByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if op.BusinessID != businessID {
		return nil, ErrNotFound
	}

	newToken, err := generateQRToken()
	if err != nil {
		return nil, fmt.Errorf("user service: generate qr: %w", err)
	}

	if err := s.repo.RegenerateOperatorQR(ctx, operatorID, newToken); err != nil {
		return nil, err
	}
	op.QRToken = newToken
	return &OperatorWithSecret{
		OperatorResponse: op.ToResponse(),
		QRToken:          newToken,
	}, nil
}

func (s *service) GetSlotInfo(ctx context.Context, businessID uuid.UUID) (int, int, error) {
	used, err := s.repo.CountActiveOperators(ctx, businessID)
	if err != nil {
		return 0, 0, err
	}
	max, err := s.plan.MaxOperators(ctx, businessID)
	if err != nil {
		return 0, 0, err
	}
	return used, max, nil
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

// generateQRToken menghasilkan url-safe base64 dari 32 byte crypto/rand.
func generateQRToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
