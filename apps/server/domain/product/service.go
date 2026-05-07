package product

import (
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListProducts(businessID uuid.UUID, filter FilterParams) ([]Product, error) {
	products, err := s.repo.FindAllByBusiness(businessID, filter)
	if err != nil {
		return nil, fmt.Errorf("product service: list: %w", err)
	}
	// Nil slice jadi empty array — frontend nggak perlu handle null
	if products == nil {
		products = []Product{}
	}
	return products, nil
}

func (s *Service) CreateProduct(businessID uuid.UUID, input CreateInput) (*Product, error) {
	p := &Product{
		BusinessID:  businessID,
		Name:        input.Name,
		Price:       input.Price,
		Category:    input.Category,
		Description: input.Description,
		IsAvailable: true,
	}
	created, err := s.repo.Create(p)
	if err != nil {
		return nil, fmt.Errorf("product service: create: %w", err)
	}
	return created, nil
}

func (s *Service) UpdateProduct(productID, businessID uuid.UUID, input UpdateInput) (*Product, error) {
	updated, err := s.repo.Update(productID, businessID, input)
	if err != nil {
		return nil, fmt.Errorf("product service: update: %w", err)
	}
	return updated, nil
}

func (s *Service) DeleteProduct(productID, businessID uuid.UUID) error {
	if err := s.repo.SoftDelete(productID, businessID); err != nil {
		return fmt.Errorf("product service: delete: %w", err)
	}
	return nil
}
