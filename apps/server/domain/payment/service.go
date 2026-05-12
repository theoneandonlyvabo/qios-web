// domain/payment/service.go
//
// Business logic untuk payment domain.
// Service tidak menyentuh database langsung dan tidak menyentuh HTTP.
// Semua operasi yang butuh konsistensi cross-table dilakukan dalam DB transaction.

package payment

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/product"
)

// ProductLookup adalah small interface yang dibutuhkan payment service untuk
// resolve harga produk server-side. Diimplementasi oleh product.Service.
//
// Sengaja didefinisikan di sisi consumer (payment) — "accept interfaces, return structs".
type ProductLookup interface {
	FindByID(productID, businessID uuid.UUID) (*product.Product, error)
}

// Service mendefinisikan kontrak business logic payment domain.
type Service interface {
	// CreateOrder membuat order baru dari kasir.
	// Harga produk diambil dari DB — client tidak boleh kirim price.
	CreateOrder(ctx context.Context, businessID, operatorID uuid.UUID, req CreateOrderRequest) (*OrderResponse, error)

	// GetOrder mengambil detail order by UUID.
	GetOrder(ctx context.Context, businessID uuid.UUID, orderID uuid.UUID) (*OrderResponse, error)

	// ListOrders mengambil list order dengan filter dan pagination.
	ListOrders(ctx context.Context, businessID uuid.UUID, filter ListOrdersFilter) ([]*OrderResponse, int, error)

	// CompleteCashOrder menyelesaikan order — operator menandai PAID secara manual.
	// MVP tidak punya Xendit, jadi semua payment_method dikonfirmasi manual via endpoint ini.
	CompleteCashOrder(ctx context.Context, businessID uuid.UUID, orderID uuid.UUID) (*OrderResponse, error)
}

type service struct {
	db         *sql.DB
	repo       Repository
	xenditSvc  *XenditService
	productSvc ProductLookup
}

// NewService merangkai dependency payment service.
// xenditSvc boleh nil saat MVP — manual confirmation path tidak panggil Xendit.
func NewService(db *sql.DB, repo Repository, xenditSvc *XenditService, productSvc ProductLookup) Service {
	return &service{db: db, repo: repo, xenditSvc: xenditSvc, productSvc: productSvc}
}

func (s *service) CreateOrder(ctx context.Context, businessID, operatorID uuid.UUID, req CreateOrderRequest) (*OrderResponse, error) {
	// 1. Resolve product prices server-side. Client-sent prices tidak dipakai.
	items := make([]*OrderItem, 0, len(req.Items))
	var total int64
	for _, in := range req.Items {
		p, err := s.productSvc.FindByID(in.ProductID, businessID)
		if err != nil {
			if errors.Is(err, product.ErrNotFound) {
				return nil, fmt.Errorf("payment service: product %s not found: %w", in.ProductID, ErrProductNotFound)
			}
			return nil, fmt.Errorf("payment service: lookup product: %w", err)
		}
		productID := p.ID
		items = append(items, &OrderItem{
			ProductID:   &productID,
			ProductName: p.Name,
			UnitPrice:   p.Price,
			Quantity:    in.Quantity,
		})
		total += p.Price * int64(in.Quantity)
	}
	if total <= 0 {
		return nil, ErrInvalidTotal
	}

	// 2. Build order. order_id format: {business_id}-{unix_ts}-{random6}
	orderIDStr, err := generateOrderID(businessID)
	if err != nil {
		return nil, fmt.Errorf("payment service: generate order id: %w", err)
	}
	opID := operatorID
	order := &PosOrder{
		BusinessID:    businessID,
		OperatorID:    &opID,
		OrderID:       orderIDStr,
		TotalAmount:   total,
		PaymentMethod: req.PaymentMethod,
		Status:        OrderStatusPending,
		Note:          req.Note,
	}

	// 3. Persist atomically.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("payment service: begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := s.repo.CreateWithItems(ctx, tx, order, items); err != nil {
		return nil, fmt.Errorf("payment service: create order: %w", err)
	}

	// Additive Xendit layer: QRIS triggers QR generation + xendit_payments row.
	// Other payment methods (CASH/EWALLET/VA) keep the prior flow untouched.
	var qrString string
	if req.PaymentMethod == PaymentMethodQRIS && s.xenditSvc != nil {
		qr, err := s.generateAndRecordQR(ctx, tx, businessID, order)
		if err != nil {
			return nil, err
		}
		qrString = qr
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("payment service: commit: %w", err)
	}
	committed = true

	resp := toOrderResponse(order, items)
	resp.QRString = qrString
	return resp, nil
}

// generateAndRecordQR creates the Xendit QR for an order and inserts the
// matching xendit_payments row in the same tx as the order itself.
//
// Wrapped in the order tx — if Xendit network call succeeds but commit fails,
// the QR is orphaned at Xendit (matches the orphaned-sub-account tradeoff
// documented in CLAUDE.md Onboarding Flow).
func (s *service) generateAndRecordQR(ctx context.Context, tx *sql.Tx, businessID uuid.UUID, order *PosOrder) (string, error) {
	accountID, status, err := s.repo.GetBusinessXenditAccount(ctx, businessID)
	if err != nil {
		return "", fmt.Errorf("payment service: lookup xendit account: %w", err)
	}
	if accountID == "" {
		return "", ErrXenditNotActive
	}
	if status != StatusRegistered && status != StatusActive {
		return "", ErrXenditNotActive
	}

	res, err := s.xenditSvc.CreateQRCode(ctx, QRCodeInput{
		AccountID:  accountID,
		ExternalID: order.OrderID,
		Amount:     order.TotalAmount,
	})
	if err != nil {
		return "", fmt.Errorf("payment service: create qr: %w", err)
	}

	row := &XenditPayment{
		PosOrderID:      order.ID,
		XenditAccountID: accountID,
		XenditChargeID:  res.XenditID,
		PaymentMethod:   PaymentMethodQRIS,
		Amount:          order.TotalAmount,
		Status:          "PENDING",
		QRString:        res.QRString,
		RawPayload:      res.RawPayload,
	}
	if err := s.repo.InsertXenditPayment(ctx, tx, row); err != nil {
		return "", fmt.Errorf("payment service: persist xendit_payment: %w", err)
	}
	return res.QRString, nil
}

func (s *service) GetOrder(ctx context.Context, businessID, orderID uuid.UUID) (*OrderResponse, error) {
	order, items, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("payment service: get order: %w", err)
	}
	if order.BusinessID != businessID {
		// Treat cross-business access as not found — jangan bocorkan keberadaan order.
		return nil, ErrOrderNotFound
	}
	return toOrderResponse(order, items), nil
}

func (s *service) ListOrders(ctx context.Context, businessID uuid.UUID, filter ListOrdersFilter) ([]*OrderResponse, int, error) {
	orders, total, err := s.repo.FindByBusinessID(ctx, businessID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("payment service: list orders: %w", err)
	}
	out := make([]*OrderResponse, 0, len(orders))
	for _, o := range orders {
		out = append(out, toOrderResponse(o, nil))
	}
	return out, total, nil
}

func (s *service) CompleteCashOrder(ctx context.Context, businessID, orderID uuid.UUID) (*OrderResponse, error) {
	order, items, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("payment service: load order: %w", err)
	}
	if order.BusinessID != businessID {
		return nil, ErrOrderNotFound
	}
	if order.Status == OrderStatusPaid {
		return nil, ErrOrderAlreadyPaid
	}
	if order.Status != OrderStatusPending {
		return nil, ErrInvalidStatus
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("payment service: begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	if err := s.repo.UpdateStatus(ctx, tx, order.ID, OrderStatusPaid, &now); err != nil {
		return nil, fmt.Errorf("payment service: mark paid: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("payment service: commit: %w", err)
	}
	committed = true

	order.Status = OrderStatusPaid
	order.PaidAt = &now
	return toOrderResponse(order, items), nil
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

// generateOrderID returns "{business_id}-{unix_ts}-{random_hex6}".
// Length: 36 + 1 + 10 + 1 + 12 = 60 chars — fits VARCHAR(100).
func generateOrderID(businessID uuid.UUID) (string, error) {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%s", businessID.String(), time.Now().Unix(), hex.EncodeToString(buf)), nil
}

func toOrderResponse(o *PosOrder, items []*OrderItem) *OrderResponse {
	resp := &OrderResponse{
		ID:            o.ID,
		OrderID:       o.OrderID,
		TotalAmount:   o.TotalAmount,
		PaymentMethod: o.PaymentMethod,
		Status:        o.Status,
		Note:          o.Note,
		PaidAt:        o.PaidAt,
		CreatedAt:     o.CreatedAt,
		Items:         make([]OrderItemResponse, 0, len(items)),
	}
	for _, it := range items {
		resp.Items = append(resp.Items, OrderItemResponse{
			ProductID:   it.ProductID,
			ProductName: it.ProductName,
			UnitPrice:   it.UnitPrice,
			Quantity:    it.Quantity,
			Subtotal:    it.Subtotal,
		})
	}
	return resp
}
