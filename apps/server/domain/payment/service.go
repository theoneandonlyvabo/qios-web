// domain/payment/service.go
//
// Business logic untuk payment domain.
// Service tidak menyentuh database langsung dan tidak menyentuh HTTP.
// Semua operasi yang butuh konsistensi cross-table dilakukan dalam DB transaction.

package payment

import (
	"context"

	"github.com/google/uuid"
)

// Service mendefinisikan kontrak business logic payment domain.
type Service interface {
	// CreateOrder membuat order baru dari kasir.
	// Untuk payment_method CASH: order langsung PENDING, kasir akan call CompleteCashOrder.
	// Untuk payment_method QRIS: order PENDING, lalu generate QR Xendit.
	CreateOrder(ctx context.Context, businessID, operatorID uuid.UUID, req CreateOrderRequest) (*OrderResponse, error)

	// GetOrder mengambil detail order by UUID.
	GetOrder(ctx context.Context, businessID uuid.UUID, orderID uuid.UUID) (*OrderResponse, error)

	// ListOrders mengambil list order dengan filter dan pagination.
	ListOrders(ctx context.Context, businessID uuid.UUID, filter ListOrdersFilter) ([]*OrderResponse, int, error)

	// CompleteCashOrder menyelesaikan order cash — update status ke PAID.
	// Tidak ada row xendit_payments yang dibuat.
	CompleteCashOrder(ctx context.Context, businessID uuid.UUID, orderID uuid.UUID) (*OrderResponse, error)
}

type service struct {
	repo      Repository
	xenditSvc *XenditService // untuk QRIS flow — lihat xendit_service.go
}

// NewService merangkai Repository dan XenditService.
// xenditSvc boleh nil saat testing cash flow saja.
func NewService(repo Repository, xenditSvc *XenditService) Service {
	return &service{repo: repo, xenditSvc: xenditSvc}
}

func (s *service) CreateOrder(ctx context.Context, businessID, operatorID uuid.UUID, req CreateOrderRequest) (*OrderResponse, error) {
	// TODO: implement
	// 1. Validate req.Items (product IDs exist, quantity > 0)
	// 2. Fetch product prices dari DB (snapshot)
	// 3. Hitung total_amount
	// 4. Generate order_id (QIOS-YYYYMMDD-xxxx)
	// 5. Begin tx
	// 6. CreateWithItems
	// 7. Kalau QRIS: call XenditService.CreateQRCode (sprint berikutnya)
	// 8. Commit
	panic("not implemented")
}

func (s *service) GetOrder(ctx context.Context, businessID uuid.UUID, orderID uuid.UUID) (*OrderResponse, error) {
	// TODO: implement
	// FindByID → validate businessID match → map ke OrderResponse
	panic("not implemented")
}

func (s *service) ListOrders(ctx context.Context, businessID uuid.UUID, filter ListOrdersFilter) ([]*OrderResponse, int, error) {
	// TODO: implement
	panic("not implemented")
}

func (s *service) CompleteCashOrder(ctx context.Context, businessID uuid.UUID, orderID uuid.UUID) (*OrderResponse, error) {
	// TODO: implement
	// 1. FindByID → validate business + status == PENDING + payment_method == CASH
	// 2. Begin tx
	// 3. UpdateStatus(PAID, paidAt=NOW)
	// 4. Commit
	// 5. Return updated order
	panic("not implemented")
}
