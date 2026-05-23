// core/analytics/queries.go

package analytics

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

func (q *Queries) fetchSummary(ctx context.Context, businessID uuid.UUID, start, end time.Time) (periodSummary, error) {
	var s periodSummary
	err := q.db.QueryRowContext(ctx,
		`SELECT
		    COALESCE(SUM(total_amount), 0),
		    COUNT(*),
		    COALESCE(ROUND(AVG(total_amount)), 0)
		 FROM pos_orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= DATE($2) AND DATE(created_at) <= DATE($3)`,
		businessID, start, end,
	).Scan(&s.TotalRevenue, &s.TotalTransactions, &s.AvgTransactionValue)
	if err != nil {
		return s, fmt.Errorf("analytics: summary: %w", err)
	}
	return s, nil
}

func (q *Queries) fetchDailyBreakdown(ctx context.Context, businessID uuid.UUID, start, end time.Time) ([]*DailyTrend, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT
		    TO_CHAR(DATE(created_at), 'YYYY-MM-DD'),
		    COUNT(*),
		    COALESCE(SUM(total_amount), 0)
		 FROM pos_orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= DATE($2) AND DATE(created_at) <= DATE($3)
		 GROUP BY DATE(created_at)
		 ORDER BY DATE(created_at) ASC`,
		businessID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("analytics: daily breakdown: %w", err)
	}
	defer rows.Close()

	var out []*DailyTrend
	for rows.Next() {
		var t DailyTrend
		if err := rows.Scan(&t.Date, &t.TransactionCount, &t.Revenue); err != nil {
			return nil, fmt.Errorf("analytics: daily breakdown scan: %w", err)
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

func (q *Queries) Overview(ctx context.Context, businessID uuid.UUID, start, end time.Time, compareWith string) (*AnalyticsOverview, error) {
	curr, err := q.fetchSummary(ctx, businessID, start, end)
	if err != nil {
		return nil, err
	}

	daily, err := q.fetchDailyBreakdown(ctx, businessID, start, end)
	if err != nil {
		return nil, err
	}
	if daily == nil {
		daily = []*DailyTrend{}
	}

	ov := &AnalyticsOverview{
		CurrentPeriod: PeriodData{
			StartDate:           start.Format("2006-01-02"),
			EndDate:             end.Format("2006-01-02"),
			TotalRevenue:        curr.TotalRevenue,
			TotalTransactions:   curr.TotalTransactions,
			AvgTransactionValue: curr.AvgTransactionValue,
			DailyBreakdown:      daily,
		},
	}

	if compareWith == "none" {
		return ov, nil
	}

	duration := end.Sub(start) + 24*time.Hour // inclusive
	var compStart, compEnd time.Time
	switch compareWith {
	case "previous_year":
		compStart = start.AddDate(-1, 0, 0)
		compEnd = end.AddDate(-1, 0, 0)
	default: // previous_period
		compEnd = start.AddDate(0, 0, -1)
		compStart = compEnd.Add(-duration).AddDate(0, 0, 1)
	}

	comp, err := q.fetchSummary(ctx, businessID, compStart, compEnd)
	if err != nil {
		return nil, err
	}

	var revPct, txPct float64
	if comp.TotalRevenue > 0 {
		revPct = float64(curr.TotalRevenue-comp.TotalRevenue) / float64(comp.TotalRevenue) * 100
	}
	if comp.TotalTransactions > 0 {
		txPct = float64(curr.TotalTransactions-comp.TotalTransactions) / float64(comp.TotalTransactions) * 100
	}

	ov.Comparison = &CompData{
		StartDate:            compStart.Format("2006-01-02"),
		EndDate:              compEnd.Format("2006-01-02"),
		TotalRevenue:         comp.TotalRevenue,
		TotalTransactions:    comp.TotalTransactions,
		RevenueChangePct:     revPct,
		TransactionChangePct: txPct,
	}

	return ov, nil
}
