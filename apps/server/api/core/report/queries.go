// core/report/queries.go

package report

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Queries struct {
	db *sql.DB
}

func NewQueries(db *sql.DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) DailySales(ctx context.Context, businessID uuid.UUID, date time.Time) (*DailySalesReport, error) {
	dateStr := date.Format("2006-01-02")

	var rep DailySalesReport
	rep.Date = dateStr
	rep.BusinessID = businessID.String()

	err := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		 FROM pos_orders
		 WHERE business_id = $1 AND status = 'paid' AND DATE(created_at) = $2`,
		businessID, dateStr,
	).Scan(&rep.TotalRevenue, &rep.TotalTransactions)
	if err != nil {
		return nil, fmt.Errorf("report: daily summary: %w", err)
	}

	byMethod, err := q.methodBreakdown(ctx, businessID, dateStr, dateStr)
	if err != nil {
		return nil, err
	}
	byProduct, err := q.productBreakdown(ctx, businessID, dateStr, dateStr)
	if err != nil {
		return nil, err
	}

	rep.BreakdownByMethod = byMethod
	rep.BreakdownByProduct = byProduct
	return &rep, nil
}

func (q *Queries) MonthlySales(ctx context.Context, businessID uuid.UUID, year, month int) (*MonthlySalesReport, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")
	monthStr := start.Format("2006-01")

	var rep MonthlySalesReport
	rep.Month = monthStr
	rep.BusinessID = businessID.String()

	err := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		 FROM pos_orders
		 WHERE business_id = $1 AND status = 'paid'
		   AND DATE(created_at) >= $2 AND DATE(created_at) <= $3`,
		businessID, startStr, endStr,
	).Scan(&rep.TotalRevenue, &rep.TotalTransactions)
	if err != nil {
		return nil, fmt.Errorf("report: monthly summary: %w", err)
	}

	daily, err := q.dailyTrend(ctx, businessID, startStr, endStr)
	if err != nil {
		return nil, err
	}
	byMethod, err := q.methodBreakdown(ctx, businessID, startStr, endStr)
	if err != nil {
		return nil, err
	}
	byProduct, err := q.productBreakdown(ctx, businessID, startStr, endStr)
	if err != nil {
		return nil, err
	}

	rep.DailyBreakdown = daily
	rep.BreakdownByMethod = byMethod
	rep.BreakdownByProduct = byProduct
	return &rep, nil
}

func (q *Queries) dailyTrend(ctx context.Context, businessID uuid.UUID, startStr, endStr string) ([]*DailyTrend, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT TO_CHAR(DATE(created_at), 'YYYY-MM-DD'), COUNT(*), COALESCE(SUM(total_amount), 0)
		 FROM pos_orders
		 WHERE business_id = $1 AND status = 'paid'
		   AND DATE(created_at) >= $2 AND DATE(created_at) <= $3
		 GROUP BY DATE(created_at)
		 ORDER BY DATE(created_at) ASC`,
		businessID, startStr, endStr,
	)
	if err != nil {
		return nil, fmt.Errorf("report: daily trend: %w", err)
	}
	defer rows.Close()

	var out []*DailyTrend
	for rows.Next() {
		var t DailyTrend
		if err := rows.Scan(&t.Date, &t.TransactionCount, &t.Revenue); err != nil {
			return nil, fmt.Errorf("report: daily trend scan: %w", err)
		}
		out = append(out, &t)
	}
	if out == nil {
		out = []*DailyTrend{}
	}
	return out, rows.Err()
}

func (q *Queries) methodBreakdown(ctx context.Context, businessID uuid.UUID, startStr, endStr string) ([]*MethodBreakdown, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT payment_method, COUNT(*), COALESCE(SUM(total_amount), 0)
		 FROM pos_orders
		 WHERE business_id = $1 AND status = 'paid'
		   AND DATE(created_at) >= $2 AND DATE(created_at) <= $3
		   AND payment_method IS NOT NULL
		 GROUP BY payment_method
		 ORDER BY SUM(total_amount) DESC`,
		businessID, startStr, endStr,
	)
	if err != nil {
		return nil, fmt.Errorf("report: method breakdown: %w", err)
	}
	defer rows.Close()

	var out []*MethodBreakdown
	for rows.Next() {
		var m MethodBreakdown
		var method sql.NullString
		if err := rows.Scan(&method, &m.Count, &m.Revenue); err != nil {
			return nil, fmt.Errorf("report: method breakdown scan: %w", err)
		}
		m.Method = method.String
		out = append(out, &m)
	}
	if out == nil {
		out = []*MethodBreakdown{}
	}
	return out, rows.Err()
}

func (q *Queries) productBreakdown(ctx context.Context, businessID uuid.UUID, startStr, endStr string) ([]*TopProduct, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT oi.product_name, SUM(oi.quantity), SUM(oi.subtotal)
		 FROM pos_order_items oi
		 JOIN pos_orders o ON o.id = oi.pos_order_id
		 WHERE o.business_id = $1 AND o.status = 'paid'
		   AND DATE(o.created_at) >= $2 AND DATE(o.created_at) <= $3
		 GROUP BY oi.product_name
		 ORDER BY SUM(oi.quantity) DESC
		 LIMIT 20`,
		businessID, startStr, endStr,
	)
	if err != nil {
		return nil, fmt.Errorf("report: product breakdown: %w", err)
	}
	defer rows.Close()

	var out []*TopProduct
	for rows.Next() {
		var p TopProduct
		if err := rows.Scan(&p.ProductName, &p.QuantitySold, &p.Revenue); err != nil {
			return nil, fmt.Errorf("report: product breakdown scan: %w", err)
		}
		out = append(out, &p)
	}
	if out == nil {
		out = []*TopProduct{}
	}
	return out, rows.Err()
}
