// domain/operator/service.go
//
// Logika bisnis untuk domain operator.
// Service tidak menyentuh database langsung — semua via Repository.
// Service tidak menyentuh HTTP — handler yang menerjemahkan ke response.

package operator

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

// defaultMaxOperators adalah fallback saat user belum punya plan aktif.
const defaultMaxOperators = 3

// PlanLookup membaca slot operator maksimum dari plan aktif.
// Diambil sebagai dependency supaya service tidak perlu tahu skema subscriptions / plans.
// Implementasi default ada di NewPostgresPlanLookup.
type PlanLookup interface {
	MaxOperators(ctx context.Context, businessID uuid.UUID) (int, error)
}

// Service mendefinisikan kontrak business logic operator.
type Service interface {
	// Owner-facing
	CreateOperator(ctx context.Context, businessID uuid.UUID, req CreateOperatorRequest) (*OperatorWithSecret, error)
	GetOperators(ctx context.Context, businessID uuid.UUID) ([]OperatorResponse, error)
	GetOperatorByID(ctx context.Context, businessID, operatorID uuid.UUID) (*OperatorResponse, error)
	UpdateOperator(ctx context.Context, businessID, operatorID uuid.UUID, req UpdateOperatorRequest) (*OperatorResponse, error)
	DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error
	RegenerateQR(ctx context.Context, businessID, operatorID uuid.UUID) (*OperatorWithSecret, error)
	GetSlotInfo(ctx context.Context, businessID uuid.UUID) (used, max int, err error)

	// Operator-facing
	LoginWithCredentials(ctx context.Context, businessID uuid.UUID, req OperatorLoginRequest) (*LoginResponse, error)
	LoginWithQR(ctx context.Context, req QRLoginRequest) (*LoginResponse, error)
}

type service struct {
	repo   Repository
	plan   PlanLookup
	jwtSvc *jwt.Service
}

// NewService merangkai Repository, PlanLookup, dan jwt.Service.
func NewService(repo Repository, plan PlanLookup, jwtSvc *jwt.Service) Service {
	return &service{repo: repo, plan: plan, jwtSvc: jwtSvc}
}

// ----------------------------------------------------------------
// Owner-facing
// ----------------------------------------------------------------

func (s *service) CreateOperator(ctx context.Context, businessID uuid.UUID, req CreateOperatorRequest) (*OperatorWithSecret, error) {
	// Slot cap dari plan harus dicek sebelum insert.
	maxOps, err := s.plan.MaxOperators(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("operator service: read plan: %w", err)
	}
	count, err := s.repo.CountActiveByBusinessID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("operator service: count: %w", err)
	}
	// -1 berarti unlimited.
	if maxOps != -1 && count >= maxOps {
		return nil, ErrLimitReached
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("operator service: hash password: %w", err)
	}

	qrToken, err := generateQRToken()
	if err != nil {
		return nil, fmt.Errorf("operator service: generate qr: %w", err)
	}

	op := &Operator{
		BusinessID:   businessID,
		Name:         req.Name,
		OperatorCode: req.OperatorCode,
		PasswordHash: string(hash),
		QRToken:      qrToken,
		IsActive:     true,
	}
	if err := s.repo.Create(ctx, op); err != nil {
		return nil, err // ErrCodeTaken sudah di-wrap di repo
	}

	return &OperatorWithSecret{
		OperatorResponse: op.ToResponse(),
		QRToken:          op.QRToken,
	}, nil
}

func (s *service) GetOperators(ctx context.Context, businessID uuid.UUID) ([]OperatorResponse, error) {
	ops, err := s.repo.FindByBusinessID(ctx, businessID)
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
	op, err := s.repo.FindByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if op.BusinessID != businessID {
		return nil, ErrNotFound // jangan bocor info bisnis lain
	}
	res := op.ToResponse()
	return &res, nil
}

func (s *service) UpdateOperator(ctx context.Context, businessID, operatorID uuid.UUID, req UpdateOperatorRequest) (*OperatorResponse, error) {
	op, err := s.repo.FindByID(ctx, operatorID)
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

	if err := s.repo.Update(ctx, op); err != nil {
		return nil, err
	}
	res := op.ToResponse()
	return &res, nil
}

func (s *service) DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error {
	op, err := s.repo.FindByID(ctx, operatorID)
	if err != nil {
		return err
	}
	if op.BusinessID != businessID {
		return ErrNotFound
	}
	return s.repo.SoftDelete(ctx, operatorID)
}

func (s *service) RegenerateQR(ctx context.Context, businessID, operatorID uuid.UUID) (*OperatorWithSecret, error) {
	op, err := s.repo.FindByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if op.BusinessID != businessID {
		return nil, ErrNotFound
	}

	newToken, err := generateQRToken()
	if err != nil {
		return nil, fmt.Errorf("operator service: generate qr: %w", err)
	}

	if err := s.repo.RegenerateQR(ctx, operatorID, newToken); err != nil {
		return nil, err
	}
	op.QRToken = newToken
	return &OperatorWithSecret{
		OperatorResponse: op.ToResponse(),
		QRToken:          newToken,
	}, nil
}

func (s *service) GetSlotInfo(ctx context.Context, businessID uuid.UUID) (int, int, error) {
	used, err := s.repo.CountActiveByBusinessID(ctx, businessID)
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
// Operator-facing
// ----------------------------------------------------------------

func (s *service) LoginWithCredentials(ctx context.Context, businessID uuid.UUID, req OperatorLoginRequest) (*LoginResponse, error) {
	op, err := s.repo.FindByOperatorCode(ctx, businessID, req.OperatorCode)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !op.IsActive {
		return nil, ErrInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(op.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokenFor(op)
}

func (s *service) LoginWithQR(ctx context.Context, req QRLoginRequest) (*LoginResponse, error) {
	op, err := s.repo.FindByQRToken(ctx, req.QRToken)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !op.IsActive {
		return nil, ErrInactive
	}
	return s.issueTokenFor(op)
}

func (s *service) issueTokenFor(op *Operator) (*LoginResponse, error) {
	token, err := s.jwtSvc.IssueOperatorAccessToken(op.ID.String(), op.BusinessID.String())
	if err != nil {
		return nil, fmt.Errorf("operator service: issue token: %w", err)
	}
	return &LoginResponse{
		AccessToken: token,
		Operator:    op.ToResponse(),
	}, nil
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

// generateQRToken menghasilkan url-safe base64 dari 32 byte crypto/rand.
// Tanpa padding agar aman di-encode ke QR code.
func generateQRToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ----------------------------------------------------------------
// PlanLookup default
// ----------------------------------------------------------------

// PostgresPlanLookup membaca max_operators dari subscription aktif user.
// Mengikuti pola yang sudah dipakai di domain/user — fallback 3 kalau row tidak ada,
// fallback 3 kalau plans/subscriptions tabel belum ada (migrasi 004 masih pending).
type PostgresPlanLookup struct {
	db *sql.DB
}

func NewPostgresPlanLookup(db *sql.DB) *PostgresPlanLookup {
	return &PostgresPlanLookup{db: db}
}

func (p *PostgresPlanLookup) MaxOperators(ctx context.Context, businessID uuid.UUID) (int, error) {
	var maxOps sql.NullInt64
	err := p.db.QueryRowContext(ctx,
		`SELECT p.max_operators
		 FROM businesses b
		 JOIN subscriptions s ON s.user_id = b.user_id
		 JOIN plans p         ON p.id      = s.plan_id
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
		// Fallback ketika tabel belum ada (migrasi 004 pending).
		// Cek code 42P01 = undefined_table.
		if isUndefinedTable(err) {
			return defaultMaxOperators, nil
		}
		return 0, fmt.Errorf("plan lookup: %w", err)
	}
	if !maxOps.Valid {
		return defaultMaxOperators, nil
	}
	return int(maxOps.Int64), nil
}
