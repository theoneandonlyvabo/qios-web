// core/product/service.go
//
// Logika bisnis untuk domain product.
// Service tidak menyentuh database langsung — semua via Repository.
// Service tidak menyentuh HTTP — handler yang menerjemahkan ke response.

package product

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service mendefinisikan kontrak business logic product.
type Service interface {
	Create(ctx context.Context, businessID uuid.UUID, req CreateProductRequest) (*Product, error)
	List(ctx context.Context, businessID uuid.UUID) ([]*Product, error)
	GetByID(ctx context.Context, businessID, productID uuid.UUID) (*Product, error)
	Update(ctx context.Context, businessID, productID uuid.UUID, req UpdateProductRequest) (*Product, error)
	Delete(ctx context.Context, businessID, productID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, businessID uuid.UUID, req CreateProductRequest) (*Product, error) {
	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	p := &Product{
		BusinessID:  businessID,
		Name:        req.Name,
		Price:       req.Price,
		Category:    req.Category,
		Description: req.Description,
		IsAvailable: isAvailable,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("product service: create: %w", err)
	}
	return p, nil
}

func (s *service) List(ctx context.Context, businessID uuid.UUID) ([]*Product, error) {
	products, err := s.repo.FindByBusinessID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("product service: list: %w", err)
	}
	return products, nil
}

func (s *service) GetByID(ctx context.Context, businessID, productID uuid.UUID) (*Product, error) {
	return s.repo.FindByID(ctx, productID, businessID)
}

func (s *service) Update(ctx context.Context, businessID, productID uuid.UUID, req UpdateProductRequest) (*Product, error) {
	p, err := s.repo.FindByID(ctx, productID, businessID)
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

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("product service: update: %w", err)
	}
	return p, nil
}

func (s *service) Delete(ctx context.Context, businessID, productID uuid.UUID) error {
	return s.repo.SoftDelete(ctx, productID, businessID)
}
