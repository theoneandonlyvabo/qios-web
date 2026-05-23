// core/admin/service.go
//
// Logika bisnis untuk domain admin.
// Service tidak menyentuh database langsung — semua via Repository.
// Service tidak menyentuh HTTP — handler yang menerjemahkan ke response.

package admin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	appjwt "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/jwt"
)

type Service interface {
	Login(ctx context.Context, email, password string) (*LoginResult, error)
	Refresh(ctx context.Context, refreshTokenPlain string) (*RefreshResult, error)
	Logout(ctx context.Context, refreshTokenPlain string) error
	Me(ctx context.Context, adminID uuid.UUID) (*AdminResponse, error)

	ListBusinesses(ctx context.Context) ([]*Business, error)
	CreateBusiness(ctx context.Context, req CreateBusinessRequest) (*Business, error)
	GetBusiness(ctx context.Context, id uuid.UUID) (*Business, error)
	UpdateBusiness(ctx context.Context, id uuid.UUID, req UpdateBusinessRequest) (*Business, error)

	ListProducts(ctx context.Context, businessID uuid.UUID) ([]*AdminProduct, error)
	CreateProduct(ctx context.Context, businessID uuid.UUID, req AdminCreateProductRequest) (*AdminProduct, error)
	UpdateProduct(ctx context.Context, productID uuid.UUID, req AdminUpdateProductRequest) (*AdminProduct, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID) error

	DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error

	ListTransactions(ctx context.Context, f AdminListTransactionsFilter) (*AdminListTransactionsResult, error)
	VoidTransaction(ctx context.Context, transactionID uuid.UUID) error
}

type service struct {
	repo   Repository
	jwtSvc *appjwt.Service
}

func NewService(repo Repository, jwtSvc *appjwt.Service) Service {
	return &service{repo: repo, jwtSvc: jwtSvc}
}

func (s *service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	admin, err := s.repo.FindAdminByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if !admin.IsActive {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.jwtSvc.IssueAccessToken(admin.ID.String(), "", appjwt.RoleAdmin)
	if err != nil {
		return nil, fmt.Errorf("admin service: issue token: %w", err)
	}

	plain, hashed, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("admin service: generate refresh: %w", err)
	}
	if err := s.repo.StoreAdminRefreshToken(ctx, admin.ID, hashed, s.jwtSvc.RefreshExpiry()); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:   accessToken,
		RefreshToken:  plain,
		RefreshExpiry: s.jwtSvc.RefreshExpiry(),
	}, nil
}

func (s *service) Refresh(ctx context.Context, plain string) (*RefreshResult, error) {
	tokenHash := hashToken(plain)

	adminID, expiresAt, err := s.repo.FindAdminRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if time.Now().After(expiresAt) {
		_ = s.repo.DeleteAdminRefreshToken(ctx, tokenHash)
		return nil, ErrSessionExpired
	}

	if err := s.repo.DeleteAdminRefreshToken(ctx, tokenHash); err != nil {
		return nil, err
	}

	accessToken, err := s.jwtSvc.IssueAccessToken(adminID.String(), "", appjwt.RoleAdmin)
	if err != nil {
		return nil, fmt.Errorf("admin service: issue token: %w", err)
	}

	newPlain, newHashed, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("admin service: generate refresh: %w", err)
	}
	if err := s.repo.StoreAdminRefreshToken(ctx, adminID, newHashed, s.jwtSvc.RefreshExpiry()); err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:   accessToken,
		RefreshToken:  newPlain,
		RefreshExpiry: s.jwtSvc.RefreshExpiry(),
	}, nil
}

func (s *service) Logout(ctx context.Context, plain string) error {
	if plain == "" {
		return nil
	}
	return s.repo.DeleteAdminRefreshToken(ctx, hashToken(plain))
}

func (s *service) Me(ctx context.Context, adminID uuid.UUID) (*AdminResponse, error) {
	a, err := s.repo.FindAdminByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return &AdminResponse{
		ID:        a.ID,
		Email:     a.Email,
		FullName:  a.FullName,
		IsActive:  a.IsActive,
		CreatedAt: a.CreatedAt,
	}, nil
}

func (s *service) ListBusinesses(ctx context.Context) ([]*Business, error) {
	return s.repo.ListBusinesses(ctx)
}

func (s *service) CreateBusiness(ctx context.Context, req CreateBusinessRequest) (*Business, error) {
	return s.repo.CreateBusiness(ctx, req)
}

func (s *service) GetBusiness(ctx context.Context, id uuid.UUID) (*Business, error) {
	return s.repo.FindBusinessByID(ctx, id)
}

func (s *service) UpdateBusiness(ctx context.Context, id uuid.UUID, req UpdateBusinessRequest) (*Business, error) {
	b, err := s.repo.FindBusinessByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.BusinessName != nil {
		b.BusinessName = *req.BusinessName
	}
	if req.Phone != nil {
		b.Phone = req.Phone
	}
	if req.Address != nil {
		b.Address = req.Address
	}
	if req.City != nil {
		b.City = req.City
	}
	if req.Country != nil {
		b.Country = req.Country
	}
	if req.XenditStatus != nil {
		b.XenditStatus = *req.XenditStatus
	}

	if err := s.repo.UpdateBusiness(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *service) ListProducts(ctx context.Context, businessID uuid.UUID) ([]*AdminProduct, error) {
	return s.repo.ListProductsByBusiness(ctx, businessID)
}

func (s *service) CreateProduct(ctx context.Context, businessID uuid.UUID, req AdminCreateProductRequest) (*AdminProduct, error) {
	return s.repo.CreateProduct(ctx, businessID, req)
}

func (s *service) UpdateProduct(ctx context.Context, productID uuid.UUID, req AdminUpdateProductRequest) (*AdminProduct, error) {
	p, err := s.repo.FindProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.Category != nil {
		p.Category = req.Category
	}
	if req.Description != nil {
		p.Description = req.Description
	}
	if req.IsAvailable != nil {
		p.IsAvailable = *req.IsAvailable
	}

	if err := s.repo.UpdateProduct(ctx, p); err != nil {
		return nil, fmt.Errorf("admin service: update product: %w", err)
	}
	return p, nil
}

func (s *service) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	return s.repo.SoftDeleteProduct(ctx, productID)
}

func (s *service) DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error {
	return s.repo.DeleteOperator(ctx, businessID, operatorID)
}

func (s *service) ListTransactions(ctx context.Context, f AdminListTransactionsFilter) (*AdminListTransactionsResult, error) {
	txs, total, err := s.repo.ListTransactions(ctx, f)
	if err != nil {
		return nil, err
	}
	if txs == nil {
		txs = []*AdminTransaction{}
	}
	return &AdminListTransactionsResult{
		Transactions: txs,
		Total:        total,
		Page:         f.Page,
		Limit:        f.Limit,
	}, nil
}

func (s *service) VoidTransaction(ctx context.Context, id uuid.UUID) error {
	return s.repo.VoidTransaction(ctx, id)
}

func generateToken() (plain, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	plain = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(plain))
	hashed = hex.EncodeToString(sum[:])
	return plain, hashed, nil
}

func hashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
