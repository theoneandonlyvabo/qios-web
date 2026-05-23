// core/transaction/service.go
//
// Logika bisnis untuk domain transaction — read-only.
// Domain ini hanya menyediakan List dan GetByID untuk history transaksi.
// Write operations sudah dipindah ke domain /pos.

package transaction

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service mendefinisikan kontrak business logic transaksi (read-only).
type Service interface {
	List(ctx context.Context, businessID uuid.UUID, f ListFilter) (*ListResult, error)
	GetByID(ctx context.Context, businessID, orderID uuid.UUID) (*OrderWithItems, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ----------------------------------------------------------------
// List
// ----------------------------------------------------------------

func (s *service) List(ctx context.Context, businessID uuid.UUID, f ListFilter) (*ListResult, error) {
	orders, total, err := s.repo.List(ctx, businessID, f)
	if err != nil {
		return nil, fmt.Errorf("transaction service: list: %w", err)
	}

	if orders == nil {
		orders = []*Order{}
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	limit := f.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return &ListResult{
		Transactions: orders,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}, nil
}

// ----------------------------------------------------------------
// GetByID
// ----------------------------------------------------------------

func (s *service) GetByID(ctx context.Context, businessID, orderID uuid.UUID) (*OrderWithItems, error) {
	return s.repo.FindByID(ctx, orderID, businessID)
}
