package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/jwt"
)

type Service interface {
	Login(ctx context.Context, email, password string) (*LoginResult, error)
	Refresh(ctx context.Context, refreshTokenPlain string) (*RefreshResult, error)
	Logout(ctx context.Context, refreshTokenPlain string) error

	OperatorLoginWithCredentials(ctx context.Context, businessID uuid.UUID, operatorCode, password string) (*OperatorLoginResult, error)
	OperatorLoginWithQR(ctx context.Context, qrToken string) (*OperatorLoginResult, error)
}

type service struct {
	repo   Repository
	jwtSvc *jwt.Service
}

func NewService(repo Repository, jwtSvc *jwt.Service) Service {
	return &service{repo: repo, jwtSvc: jwtSvc}
}

func (s *service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if !user.IsActive || user.IsSuspended {
		return nil, ErrAccountInactive
	}
	if user.PasswordHash == "" {
		return nil, ErrGoogleOnlyAccount
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.jwtSvc.IssueAccessToken(user.ID, user.BusinessID, jwt.RoleOwner)
	if err != nil {
		return nil, fmt.Errorf("auth service: issue access token: %w", err)
	}

	plain, hashed, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth service: generate refresh: %w", err)
	}
	if err := s.repo.StoreRefreshToken(ctx, user.ID, hashed, s.jwtSvc.RefreshExpiry()); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:   accessToken,
		RefreshToken:  plain,
		RefreshExpiry: s.jwtSvc.RefreshExpiry(),
	}, nil
}

func (s *service) Refresh(ctx context.Context, refreshTokenPlain string) (*RefreshResult, error) {
	tokenHash := hashToken(refreshTokenPlain)

	userID, expiresAt, err := s.repo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if time.Now().After(expiresAt) {
		_ = s.repo.DeleteRefreshToken(ctx, tokenHash)
		return nil, ErrSessionExpired
	}

	businessID, role, err := s.repo.FindUserRoleByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteRefreshToken(ctx, tokenHash); err != nil {
		return nil, err
	}

	var newAccessToken string
	if role == roleOperator {
		newAccessToken, err = s.jwtSvc.IssueOperatorAccessToken(userID, businessID)
	} else {
		newAccessToken, err = s.jwtSvc.IssueAccessToken(userID, businessID, role)
	}
	if err != nil {
		return nil, fmt.Errorf("auth service: issue access token: %w", err)
	}

	newPlain, newHashed, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth service: generate refresh: %w", err)
	}
	if err := s.repo.StoreRefreshToken(ctx, userID, newHashed, s.jwtSvc.RefreshExpiry()); err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:   newAccessToken,
		RefreshToken:  newPlain,
		RefreshExpiry: s.jwtSvc.RefreshExpiry(),
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshTokenPlain string) error {
	if refreshTokenPlain == "" {
		return nil
	}
	return s.repo.DeleteRefreshToken(ctx, hashToken(refreshTokenPlain))
}

func generateRefreshToken() (plain, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plain = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(plain))
	hashed = hex.EncodeToString(sum[:])
	return
}

func hashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// ----------------------------------------------------------------
// Operator-facing login
// ----------------------------------------------------------------

func (s *service) OperatorLoginWithCredentials(ctx context.Context, businessID uuid.UUID, operatorCode, password string) (*OperatorLoginResult, error) {
	op, err := s.repo.FindOperatorByCode(ctx, businessID, operatorCode)
	if errors.Is(err, ErrInvalidCredentials) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !op.IsActive {
		return nil, ErrOperatorInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(op.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issueOperatorToken(op)
}

func (s *service) OperatorLoginWithQR(ctx context.Context, qrToken string) (*OperatorLoginResult, error) {
	op, err := s.repo.FindOperatorByQRToken(ctx, qrToken)
	if errors.Is(err, ErrInvalidCredentials) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !op.IsActive {
		return nil, ErrOperatorInactive
	}
	return s.issueOperatorToken(op)
}

func (s *service) issueOperatorToken(op *OperatorLoginData) (*OperatorLoginResult, error) {
	token, err := s.jwtSvc.IssueOperatorAccessToken(op.ID.String(), op.BusinessID.String())
	if err != nil {
		return nil, fmt.Errorf("auth service: issue operator token: %w", err)
	}
	return &OperatorLoginResult{
		AccessToken: token,
		Operator: OperatorInfo{
			ID:         op.ID,
			BusinessID: op.BusinessID,
			Name:       op.Name,
			IsActive:   op.IsActive,
			CreatedAt:  op.CreatedAt,
			UpdatedAt:  op.UpdatedAt,
		},
	}, nil
}

var _ Service = (*service)(nil)
