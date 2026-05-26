// core/order/service.go
//
// Logika bisnis untuk domain order.
// Service tidak menyentuh database langsung — semua via Repository.
//
// Order flow: CreateOrder → UpdateItems → BeginCheckout → ConfirmCheckout / VoidOrder
// Session management: ListActiveSessions, ForceEndSession

package order

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service mendefinisikan kontrak business logic pos.
type Service interface {
	CreateOrder(ctx context.Context, businessID uuid.UUID, operatorID *uuid.UUID, req CreateOrderRequest) (*OrderWithItems, error)
	UpdateItems(ctx context.Context, businessID, orderID uuid.UUID, req UpdateItemsRequest) (*OrderWithItems, error)
	BeginCheckout(ctx context.Context, businessID, orderID uuid.UUID) (*Order, error)
	ConfirmCheckout(ctx context.Context, businessID, orderID uuid.UUID, req ConfirmOrderRequest) (*ConfirmResponse, error)
	VoidOrder(ctx context.Context, businessID, orderID uuid.UUID, callerOperatorID *uuid.UUID) error
	ListMyOrders(ctx context.Context, operatorID, businessID uuid.UUID) ([]*Order, error)
	ListActiveSessions(ctx context.Context, businessID uuid.UUID) ([]*SessionWithOperator, error)
	ForceEndSession(ctx context.Context, sessionID, businessID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ----------------------------------------------------------------
// CreateOrder
// ----------------------------------------------------------------

func (s *service) CreateOrder(ctx context.Context, businessID uuid.UUID, operatorID *uuid.UUID, req CreateOrderRequest) (*OrderWithItems, error) {
	if len(req.Items) == 0 {
		return nil, ErrEmptyItems
	}

	// Collect unique product IDs
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
		return nil, fmt.Errorf("pos service: find products: %w", err)
	}

	byID := make(map[uuid.UUID]productSnapshot, len(snapshots))
	for _, snap := range snapshots {
		byID[snap.id] = snap
	}

	var totalAmount int64
	items := make([]*OrderItem, 0, len(req.Items))
	for _, inp := range req.Items {
		pid, _ := uuid.Parse(inp.ProductID)
		snap, ok := byID[pid]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrProductNotFound, inp.ProductID)
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
		return nil, fmt.Errorf("pos service: generate order id: %w", err)
	}

	order := &Order{
		BusinessID:  businessID,
		OperatorID:  operatorID,
		OrderID:     orderID,
		TotalAmount: totalAmount,
		Status:      StatusDraft,
		Note:        req.Note,
	}

	if err := s.repo.Create(ctx, order, items); err != nil {
		return nil, fmt.Errorf("pos service: create: %w", err)
	}

	return &OrderWithItems{Order: *order, Items: items}, nil
}

// ----------------------------------------------------------------
// UpdateItems
// ----------------------------------------------------------------

func (s *service) UpdateItems(ctx context.Context, businessID, orderID uuid.UUID, req UpdateItemsRequest) (*OrderWithItems, error) {
	if len(req.Items) == 0 {
		return nil, ErrEmptyItems
	}

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
		return nil, fmt.Errorf("pos service: find products: %w", err)
	}

	byID := make(map[uuid.UUID]productSnapshot, len(snapshots))
	for _, snap := range snapshots {
		byID[snap.id] = snap
	}

	items := make([]*OrderItem, 0, len(req.Items))
	for _, inp := range req.Items {
		pid, _ := uuid.Parse(inp.ProductID)
		snap, ok := byID[pid]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrProductNotFound, inp.ProductID)
		}
		items = append(items, &OrderItem{
			ProductID:   &pid,
			ProductName: snap.name,
			UnitPrice:   snap.price,
			Quantity:    inp.Quantity,
		})
	}

	if err := s.repo.UpdateItems(ctx, orderID, businessID, items); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, orderID, businessID)
}

// ----------------------------------------------------------------
// BeginCheckout
// ----------------------------------------------------------------

func (s *service) BeginCheckout(ctx context.Context, businessID, orderID uuid.UUID) (*Order, error) {
	if err := s.repo.BeginCheckout(ctx, orderID, businessID); err != nil {
		return nil, err
	}
	result, err := s.repo.FindByID(ctx, orderID, businessID)
	if err != nil {
		return nil, err
	}
	return &result.Order, nil
}

// ----------------------------------------------------------------
// ConfirmCheckout
// ----------------------------------------------------------------

func (s *service) ConfirmCheckout(ctx context.Context, businessID, orderID uuid.UUID, req ConfirmOrderRequest) (*ConfirmResponse, error) {
	// ConfirmAtomic holds a FOR UPDATE row lock, re-checks status and timing,
	// then flips status to CONFIRMED — all in one transaction. This prevents
	// a concurrent confirm from passing the 800ms guard and firing side-effects
	// (consumption log, QRIS lookup) more than once.
	now := time.Now()
	if err := s.repo.ConfirmAtomic(ctx, orderID, businessID, req.PaymentMethod, now, 800*time.Millisecond); err != nil {
		return nil, err
	}

	result, err := s.repo.FindByID(ctx, orderID, businessID)
	if err != nil {
		return nil, err
	}

	// Populate consumption_log asynchronously — pakai timeout 30s, bukan Background()
	// supaya goroutine tidak leak selamanya kalau DB shutdown atau hang.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.populateConsumptionLog(ctx, result, businessID, now)
	}()

	res := &ConfirmResponse{Order: result.Order}

	if req.PaymentMethod == PaymentQRIS {
		qs, err := s.repo.GetBusinessQrisString(ctx, businessID)
		if err != nil {
			return nil, err
		}
		res.QrisString = qs
	}

	return res, nil
}

// ----------------------------------------------------------------
// VoidOrder
// ----------------------------------------------------------------

func (s *service) VoidOrder(ctx context.Context, businessID, orderID uuid.UUID, callerOperatorID *uuid.UUID) error {
	return s.repo.Void(ctx, orderID, businessID, callerOperatorID)
}

// ----------------------------------------------------------------
// ListMyOrders
// ----------------------------------------------------------------

func (s *service) ListMyOrders(ctx context.Context, operatorID, businessID uuid.UUID) ([]*Order, error) {
	orders, err := s.repo.FindByOperatorToday(ctx, operatorID, businessID)
	if err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []*Order{}
	}
	return orders, nil
}

// ----------------------------------------------------------------
// Session management
// ----------------------------------------------------------------

func (s *service) ListActiveSessions(ctx context.Context, businessID uuid.UUID) ([]*SessionWithOperator, error) {
	sessions, err := s.repo.ListActiveSessions(ctx, businessID)
	if err != nil {
		return nil, err
	}
	if sessions == nil {
		sessions = []*SessionWithOperator{}
	}
	return sessions, nil
}

func (s *service) ForceEndSession(ctx context.Context, sessionID, businessID uuid.UUID) error {
	return s.repo.EndSession(ctx, sessionID, businessID)
}

// ----------------------------------------------------------------
// Internal helpers
// ----------------------------------------------------------------

// populateConsumptionLog mengisi consumption_log dari recipe produk.
// Dipanggil sebagai goroutine — tidak memblokir response.
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

// generateOrderID menghasilkan order_id unik: {8char_businessID}-{unix_ts}-{4hex_rand}.
func generateOrderID(businessID uuid.UUID) (string, error) {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%x", businessID.String()[:8], time.Now().Unix(), b), nil
}
