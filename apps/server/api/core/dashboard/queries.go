// core/dashboard/queries.go
//
// SQL aggregate queries untuk dashboard.
// Semua query hanya menyentuh order dengan status = 'paid'.

package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Queries membungkus semua aggregate query dashboard.
type Queries struct {
	db *sql.DB
}

func NewQueries(db *sql.DB) *Queries {
	return &Queries{db: db}
}

// ----------------------------------------------------------------
// Summary
// ----------------------------------------------------------------

func (q *Queries) Summary(ctx context.Context, businessID uuid.UUID, start, end time.Time) (*Summary, error) {
	var s Summary
	err := q.db.QueryRowContext(ctx,
		`SELECT
		    COALESCE(SUM(total_amount), 0),
		    COUNT(*),
		    COALESCE(AVG(total_amount), 0)
		 FROM pos_orders
		 WHERE business_id = $1 AND status = 'paid'
		   AND created_at >= $2 AND created_at <= $3`,
		businessID, start, end,
	).Scan(&s.TotalRevenue, &s.TotalTransactions, &s.AvgOrderValue)
	if err != nil {
		return nil, fmt.Errorf("dashboard: summary: %w", err)
	}
	return &s, nil
}

// ----------------------------------------------------------------
// Trend
// ----------------------------------------------------------------

func (q *Queries) Trend(ctx context.Context, businessID uuid.UUID, start, end time.Time) ([]*DailyTrend, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT
		    TO_CHAR(DATE(created_at), 'YYYY-MM-DD'),
		    COUNT(*),
		    COALESCE(SUM(total_amount), 0)
		 FROM pos_orders
		 WHERE business_id = $1 AND status = 'paid'
		   AND DATE(created_at) >= DATE($2) AND DATE(created_at) <= DATE($3)
		 GROUP BY DATE(created_at)
		 ORDER BY DATE(created_at) ASC`,
		businessID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("dashboard: trend: %w", err)
	}
	defer rows.Close()

	var out []*DailyTrend
	for rows.Next() {
		var t DailyTrend
		if err := rows.Scan(&t.Date, &t.TransactionCount, &t.Revenue); err != nil {
			return nil, fmt.Errorf("dashboard: trend scan: %w", err)
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

// ----------------------------------------------------------------
// TopProducts
// ----------------------------------------------------------------

func (q *Queries) TopProducts(ctx context.Context, businessID uuid.UUID, start, end time.Time, limit int) ([]*TopProduct, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT
		    oi.product_name,
		    SUM(oi.quantity)  AS quantity_sold,
		    SUM(oi.subtotal)  AS revenue
		 FROM pos_order_items oi
		 JOIN pos_orders o ON o.id = oi.pos_order_id
		 WHERE o.business_id = $1 AND o.status = 'paid'
		   AND o.created_at >= $2 AND o.created_at <= $3
		 GROUP BY oi.product_name
		 ORDER BY quantity_sold DESC
		 LIMIT $4`,
		businessID, start, end, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("dashboard: top products: %w", err)
	}
	defer rows.Close()

	var out []*TopProduct
	for rows.Next() {
		var p TopProduct
		if err := rows.Scan(&p.ProductName, &p.QuantitySold, &p.Revenue); err != nil {
			return nil, fmt.Errorf("dashboard: top products scan: %w", err)
		}
		out = append(out, &p)
	}
	return out, rows.Err()
}

// ----------------------------------------------------------------
// PeakHours
// ----------------------------------------------------------------

func (q *Queries) PeakHours(ctx context.Context, businessID uuid.UUID, start, end time.Time) ([]*HourlyDistribution, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT
		    EXTRACT(HOUR FROM created_at)::int AS hour,
		    COUNT(*),
		    COALESCE(SUM(total_amount), 0)
		 FROM pos_orders
		 WHERE business_id = $1 AND status = 'paid'
		   AND created_at >= $2 AND created_at <= $3
		 GROUP BY hour
		 ORDER BY hour ASC`,
		businessID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("dashboard: peak hours: %w", err)
	}
	defer rows.Close()

	var out []*HourlyDistribution
	for rows.Next() {
		var h HourlyDistribution
		if err := rows.Scan(&h.Hour, &h.TransactionCount, &h.Revenue); err != nil {
			return nil, fmt.Errorf("dashboard: peak hours scan: %w", err)
		}
		out = append(out, &h)
	}
	return out, rows.Err()
}
