// core/transaction/service.go
//
// Logika bisnis untuk domain transaction.
// Service tidak menyentuh database langsung — semua via Repository.

package transaction

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service mendefinisikan kontrak business logic transaksi.
type Service interface {
	Create(ctx context.Context, businessID uuid.UUID, operatorID *uuid.UUID, req CreateOrderRequest) (*OrderWithItems, error)
	List(ctx context.Context, businessID uuid.UUID, f ListFilter) (*ListResult, error)
	GetByID(ctx context.Context, businessID, orderID uuid.UUID) (*OrderWithItems, error)
	Confirm(ctx context.Context, businessID, orderID uuid.UUID, req ConfirmOrderRequest) (*ConfirmResponse, error)
	Void(ctx context.Context, businessID, orderID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ----------------------------------------------------------------
// Create
// ----------------------------------------------------------------

func (s *service) Create(ctx context.Context, businessID uuid.UUID, operatorID *uuid.UUID, req CreateOrderRequest) (*OrderWithItems, error) {
	// Kumpulkan product IDs unik dari request.
	idSet := make(map[uuid.UUID]struct{}, len(req.Items))
	for _, item := range req.Items {
		id, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrProductNotFound, item.ProductID)
		}
		idSet[id] = struct{}{}
	}

	productIDs := make([]uuid.UUID, 0, len(idSet))
	for id := range idSet {
		productIDs = append(productIDs, id)
	}

	snapshots, err := s.repo.FindProducts(ctx, businessID, productIDs)
	if err != nil {
		return nil, fmt.Errorf("transaction service: find products: %w", err)
	}

	// Index snapshot by id untuk lookup O(1).
	byID := make(map[uuid.UUID]productSnapshot, len(snapshots))
	for _, s := range snapshots {
		byID[s.id] = s
	}

	// Bangun items dengan snapshot + hitung total.
	var totalAmount int64
	items := make([]*OrderItem, 0, len(req.Items))

	for _, inp := range req.Items {
		pid, _ := uuid.Parse(inp.ProductID)
		snap, ok := byID[pid]
		if !ok {
			return nil, ErrProductNotFound
		}

		subtotal := snap.price * int64(inp.Quantity)
		totalAmount += subtotal

		items = append(items, &OrderItem{
			ProductID:   &pid,
			ProductName: snap.name,
			UnitPrice:   snap.price,
			Quantity:    inp.Quantity,
			Subtotal:    subtotal,
		})
	}

	orderID, err := generateOrderID(businessID)
	if err != nil {
		return nil, fmt.Errorf("transaction service: generate order id: %w", err)
	}

	order := &Order{
		BusinessID:  businessID,
		OperatorID:  operatorID,
		OrderID:     orderID,
		TotalAmount: totalAmount,
		Status:      StatusPending,
		Note:        req.Note,
	}

	if err := s.repo.Create(ctx, order, items); err != nil {
		return nil, fmt.Errorf("transaction service: create: %w", err)
	}

	return &OrderWithItems{Order: *order, Items: items}, nil
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

// ----------------------------------------------------------------
// Confirm
// ----------------------------------------------------------------

func (s *service) Confirm(ctx context.Context, businessID, orderID uuid.UUID, req ConfirmOrderRequest) (*ConfirmResponse, error) {
	now := time.Now()
	confirmedAt := &sql.NullTime{Valid: true, Time: now}
	method := req.PaymentMethod

	if err := s.repo.UpdateStatus(ctx, orderID, businessID, StatusConfirmed, &method, confirmedAt); err != nil {
		return nil, err
	}

	result, err := s.repo.FindByID(ctx, orderID, businessID)
	if err != nil {
		return nil, err
	}

	// Populate consumption_log dari recipe produk.
	go s.populateConsumptionLog(context.Background(), result, businessID, now)

	res := &ConfirmResponse{Order: result.Order}

	if method == PaymentQRIS {
		qs, err := s.repo.GetBusinessQrisString(ctx, businessID)
		if err != nil {
			return nil, err
		}
		res.QrisString = qs
	}

	return res, nil
}

func (s *service) populateConsumptionLog(ctx context.Context, order *OrderWithItems, businessID uuid.UUID, confirmedAt time.Time) {
	if len(order.Items) == 0 {
		return
	}
	ids := make([]uuid.UUID, 0, len(order.Items))
	for _, item := range order.Items {
		if item.ProductID != nil {
			ids = append(ids, *item.ProductID)
		}
	}
	if len(ids) == 0 {
		return
	}
	recipes, err := s.repo.FindProductRecipes(ctx, ids)
	if err != nil || len(recipes) == 0 {
		return
	}
	var entries []ConsumptionEntry
	for _, item := range order.Items {
		if item.ProductID == nil {
			continue
		}
		for _, ri := range recipes[*item.ProductID] {
			entries = append(entries, ConsumptionEntry{
				TransactionID: order.ID,
				BusinessID:    businessID,
				ProductID:     item.ProductID,
				ProductName:   item.ProductName,
				Ingredient:    ri.Ingredient,
				QuantityUsed:  ri.Quantity * float64(item.Quantity),
				Unit:          ri.Unit,
				ConfirmedAt:   confirmedAt,
			})
		}
	}
	_ = s.repo.InsertConsumptionLog(ctx, entries)
}

// ----------------------------------------------------------------
// Void
// ----------------------------------------------------------------

func (s *service) Void(ctx context.Context, businessID, orderID uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, orderID, businessID, StatusVoided, nil, nil)
}

// ----------------------------------------------------------------
// Helper
// ----------------------------------------------------------------

// generateOrderID menghasilkan order_id unik: {8char_businessID}-{unix_ts}-{4hex_rand}.
func generateOrderID(businessID uuid.UUID) (string, error) {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%x", businessID.String()[:8], time.Now().Unix(), b), nil
}
