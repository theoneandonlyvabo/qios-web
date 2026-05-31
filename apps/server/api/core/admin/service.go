// core/admin/service.go
//
// Logika bisnis untuk domain admin.
// Service tidak menyentuh database langsung — semua via Repository.
// Service tidak menyentuh HTTP — handler yang menerjemahkan ke response.

package admin

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
)

type Service interface {
	// Owner management
	ListOwners(ctx context.Context, page, limit int) (*OwnerListResult, error)
	CreateOwner(ctx context.Context, req CreateBusinessRequest) (*Business, error)
	GetOwner(ctx context.Context, businessID uuid.UUID) (*OwnerDetail, error)
	UpdateOwner(ctx context.Context, businessID uuid.UUID, req UpdateBusinessRequest) (*Business, error)
	SetOwnerStatus(ctx context.Context, businessID uuid.UUID, enabled bool) error
	SetOwnerCredential(ctx context.Context, businessID uuid.UUID, req SetOwnerCredentialRequest) error

	// Product management
	ListProducts(ctx context.Context, businessID uuid.UUID) ([]*AdminProduct, error)
	CreateProduct(ctx context.Context, businessID uuid.UUID, req AdminCreateProductRequest) (*AdminProduct, error)
	GetProduct(ctx context.Context, productID uuid.UUID) (*AdminProductDetail, error)
	UpdateProduct(ctx context.Context, productID uuid.UUID, req AdminUpdateProductRequest) (*AdminProduct, error)
	UpdateProductRecipe(ctx context.Context, productID uuid.UUID, req UpdateRecipeRequest) error
	DeleteProduct(ctx context.Context, productID uuid.UUID) error

	// Operator management
	DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error

	// Transaction management
	ListTransactions(ctx context.Context, f AdminListTransactionsFilter) (*AdminListTransactionsResult, error)
	VoidTransaction(ctx context.Context, transactionID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ----------------------------------------------------------------
// Owner
// ----------------------------------------------------------------

func (s *service) ListOwners(ctx context.Context, page, limit int) (*OwnerListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	owners, total, err := s.repo.ListOwners(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	if owners == nil {
		owners = []*OwnerSummary{}
	}
	return &OwnerListResult{Owners: owners, Total: total, Page: page, Limit: limit}, nil
}

func (s *service) CreateOwner(ctx context.Context, req CreateBusinessRequest) (*Business, error) {
	return s.repo.CreateBusiness(ctx, req)
}

func (s *service) GetOwner(ctx context.Context, businessID uuid.UUID) (*OwnerDetail, error) {
	return s.repo.FindOwnerByID(ctx, businessID)
}

func (s *service) UpdateOwner(ctx context.Context, businessID uuid.UUID, req UpdateBusinessRequest) (*Business, error) {
	b, err := s.repo.FindBusinessByID(ctx, businessID)
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
	if req.MerchantStatus != nil {
		b.MerchantStatus = *req.MerchantStatus
	}

	if err := s.repo.UpdateBusiness(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *service) SetOwnerStatus(ctx context.Context, businessID uuid.UUID, enabled bool) error {
	return s.repo.SetOwnerStatus(ctx, businessID, !enabled) // suspended = !enabled
}

func (s *service) SetOwnerCredential(ctx context.Context, businessID uuid.UUID, req SetOwnerCredentialRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("admin service: hash password: %w", err)
	}
	return s.repo.SetOwnerCredential(ctx, businessID, req.Email, string(hash))
}

// ----------------------------------------------------------------
// Product
// ----------------------------------------------------------------

func (s *service) ListProducts(ctx context.Context, businessID uuid.UUID) ([]*AdminProduct, error) {
	return s.repo.ListProductsByBusiness(ctx, businessID)
}

func (s *service) CreateProduct(ctx context.Context, businessID uuid.UUID, req AdminCreateProductRequest) (*AdminProduct, error) {
	return s.repo.CreateProduct(ctx, businessID, req)
}

func (s *service) GetProduct(ctx context.Context, productID uuid.UUID) (*AdminProductDetail, error) {
	return s.repo.FindProductDetailByID(ctx, productID)
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

func (s *service) UpdateProductRecipe(ctx context.Context, productID uuid.UUID, req UpdateRecipeRequest) error {
	raw, err := json.Marshal(req.Recipe)
	if err != nil {
		return fmt.Errorf("admin service: marshal recipe: %w", err)
	}
	return s.repo.UpdateProductRecipe(ctx, productID, json.RawMessage(raw))
}

func (s *service) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	return s.repo.SoftDeleteProduct(ctx, productID)
}

// ----------------------------------------------------------------
// Operator
// ----------------------------------------------------------------

func (s *service) DeleteOperator(ctx context.Context, businessID, operatorID uuid.UUID) error {
	return s.repo.DeleteOperator(ctx, businessID, operatorID)
}

// ----------------------------------------------------------------
// Transaction
// ----------------------------------------------------------------

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
