// core/metrics/queries.go
//
// Merged Queries struct dari dashboard, analytics, report, insight.
// status = 'paid' (bug lama di dashboard/report) dinormalisasi ke 'CONFIRMED'.
// Tabel dirujuk dengan nama baru: orders, order_items (setelah migration 014).

package metrics

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

// periodSummary — internal aggregate result (from analytics).
type periodSummary struct {
	TotalRevenue        int64
	TotalTransactions   int
	AvgTransactionValue int64
}

// ----------------------------------------------------------------
// Dashboard methods
// ----------------------------------------------------------------

func (q *Queries) Summary(ctx context.Context, businessID uuid.UUID, start, end time.Time) (*Summary, error) {
	var s Summary
	err := q.db.QueryRowContext(ctx,
		`SELECT
		    COALESCE(SUM(total_amount), 0),
		    COUNT(*),
		    COALESCE(AVG(total_amount), 0)
		 FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND created_at >= $2 AND created_at <= $3`,
		businessID, start, end,
	).Scan(&s.TotalRevenue, &s.TotalTransactions, &s.AvgOrderValue)
	if err != nil {
		return nil, fmt.Errorf("metrics: summary: %w", err)
	}
	return &s, nil
}

func (q *Queries) Trend(ctx context.Context, businessID uuid.UUID, start, end time.Time) ([]*DailyTrend, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT
		    TO_CHAR(DATE(created_at), 'YYYY-MM-DD'),
		    COUNT(*),
		    COALESCE(SUM(total_amount), 0)
		 FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= DATE($2) AND DATE(created_at) <= DATE($3)
		 GROUP BY DATE(created_at)
		 ORDER BY DATE(created_at) ASC`,
		businessID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("metrics: trend: %w", err)
	}
	defer rows.Close()

	var out []*DailyTrend
	for rows.Next() {
		var t DailyTrend
		if err := rows.Scan(&t.Date, &t.TransactionCount, &t.Revenue); err != nil {
			return nil, fmt.Errorf("metrics: trend scan: %w", err)
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

func (q *Queries) TopProducts(ctx context.Context, businessID uuid.UUID, start, end time.Time, limit int) ([]*TopProduct, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT
		    oi.product_name,
		    SUM(oi.quantity)  AS quantity_sold,
		    SUM(oi.subtotal)  AS revenue
		 FROM order_items oi
		 JOIN orders o ON o.id = oi.order_id
		 WHERE o.business_id = $1 AND o.status = 'CONFIRMED'
		   AND o.created_at >= $2 AND o.created_at <= $3
		 GROUP BY oi.product_name
		 ORDER BY quantity_sold DESC
		 LIMIT $4`,
		businessID, start, end, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("metrics: top products: %w", err)
	}
	defer rows.Close()

	var out []*TopProduct
	for rows.Next() {
		var p TopProduct
		if err := rows.Scan(&p.ProductName, &p.QuantitySold, &p.Revenue); err != nil {
			return nil, fmt.Errorf("metrics: top products scan: %w", err)
		}
		out = append(out, &p)
	}
	return out, rows.Err()
}

func (q *Queries) PeakHours(ctx context.Context, businessID uuid.UUID, start, end time.Time) ([]*HourlyDistribution, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT
		    EXTRACT(HOUR FROM created_at)::int AS hour,
		    COUNT(*),
		    COALESCE(SUM(total_amount), 0)
		 FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND created_at >= $2 AND created_at <= $3
		 GROUP BY hour
		 ORDER BY hour ASC`,
		businessID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("metrics: peak hours: %w", err)
	}
	defer rows.Close()

	var out []*HourlyDistribution
	for rows.Next() {
		var h HourlyDistribution
		if err := rows.Scan(&h.Hour, &h.TransactionCount, &h.Revenue); err != nil {
			return nil, fmt.Errorf("metrics: peak hours scan: %w", err)
		}
		out = append(out, &h)
	}
	return out, rows.Err()
}

// ----------------------------------------------------------------
// Analytics methods
// ----------------------------------------------------------------

func (q *Queries) fetchSummary(ctx context.Context, businessID uuid.UUID, start, end time.Time) (periodSummary, error) {
	var s periodSummary
	err := q.db.QueryRowContext(ctx,
		`SELECT
		    COALESCE(SUM(total_amount), 0),
		    COUNT(*),
		    COALESCE(ROUND(AVG(total_amount)), 0)
		 FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= DATE($2) AND DATE(created_at) <= DATE($3)`,
		businessID, start, end,
	).Scan(&s.TotalRevenue, &s.TotalTransactions, &s.AvgTransactionValue)
	if err != nil {
		return s, fmt.Errorf("metrics: fetch summary: %w", err)
	}
	return s, nil
}

func (q *Queries) fetchDailyBreakdown(ctx context.Context, businessID uuid.UUID, start, end time.Time) ([]*DailyTrend, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT
		    TO_CHAR(DATE(created_at), 'YYYY-MM-DD'),
		    COUNT(*),
		    COALESCE(SUM(total_amount), 0)
		 FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= DATE($2) AND DATE(created_at) <= DATE($3)
		 GROUP BY DATE(created_at)
		 ORDER BY DATE(created_at) ASC`,
		businessID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("metrics: daily breakdown: %w", err)
	}
	defer rows.Close()

	var out []*DailyTrend
	for rows.Next() {
		var t DailyTrend
		if err := rows.Scan(&t.Date, &t.TransactionCount, &t.Revenue); err != nil {
			return nil, fmt.Errorf("metrics: daily breakdown scan: %w", err)
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

	duration := end.Sub(start) + 24*time.Hour
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

// ----------------------------------------------------------------
// Report methods
// ----------------------------------------------------------------

func (q *Queries) DailySales(ctx context.Context, businessID uuid.UUID, date time.Time) (*DailySalesReport, error) {
	dateStr := date.Format("2006-01-02")

	var rep DailySalesReport
	rep.Date = dateStr
	rep.BusinessID = businessID.String()

	err := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		 FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED' AND DATE(created_at) = $2`,
		businessID, dateStr,
	).Scan(&rep.TotalRevenue, &rep.TotalTransactions)
	if err != nil {
		return nil, fmt.Errorf("metrics: daily summary: %w", err)
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
		 FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= $2 AND DATE(created_at) <= $3`,
		businessID, startStr, endStr,
	).Scan(&rep.TotalRevenue, &rep.TotalTransactions)
	if err != nil {
		return nil, fmt.Errorf("metrics: monthly summary: %w", err)
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
		 FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= $2 AND DATE(created_at) <= $3
		 GROUP BY DATE(created_at)
		 ORDER BY DATE(created_at) ASC`,
		businessID, startStr, endStr,
	)
	if err != nil {
		return nil, fmt.Errorf("metrics: daily trend: %w", err)
	}
	defer rows.Close()

	var out []*DailyTrend
	for rows.Next() {
		var t DailyTrend
		if err := rows.Scan(&t.Date, &t.TransactionCount, &t.Revenue); err != nil {
			return nil, fmt.Errorf("metrics: daily trend scan: %w", err)
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
		 FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= $2 AND DATE(created_at) <= $3
		   AND payment_method IS NOT NULL
		 GROUP BY payment_method
		 ORDER BY SUM(total_amount) DESC`,
		businessID, startStr, endStr,
	)
	if err != nil {
		return nil, fmt.Errorf("metrics: method breakdown: %w", err)
	}
	defer rows.Close()

	var out []*MethodBreakdown
	for rows.Next() {
		var m MethodBreakdown
		var method sql.NullString
		if err := rows.Scan(&method, &m.Count, &m.Revenue); err != nil {
			return nil, fmt.Errorf("metrics: method breakdown scan: %w", err)
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
		 FROM order_items oi
		 JOIN orders o ON o.id = oi.order_id
		 WHERE o.business_id = $1 AND o.status = 'CONFIRMED'
		   AND DATE(o.created_at) >= $2 AND DATE(o.created_at) <= $3
		 GROUP BY oi.product_name
		 ORDER BY SUM(oi.quantity) DESC
		 LIMIT 20`,
		businessID, startStr, endStr,
	)
	if err != nil {
		return nil, fmt.Errorf("metrics: product breakdown: %w", err)
	}
	defer rows.Close()

	var out []*TopProduct
	for rows.Next() {
		var p TopProduct
		if err := rows.Scan(&p.ProductName, &p.QuantitySold, &p.Revenue); err != nil {
			return nil, fmt.Errorf("metrics: product breakdown scan: %w", err)
		}
		out = append(out, &p)
	}
	if out == nil {
		out = []*TopProduct{}
	}
	return out, rows.Err()
}

// ----------------------------------------------------------------
// Insight methods
// ----------------------------------------------------------------

func (q *Queries) Generate(ctx context.Context, businessID uuid.UUID) ([]*InsightCard, error) {
	now := time.Now()
	today := now.Format("2006-01-02")
	window30Start := now.AddDate(0, 0, -29).Format("2006-01-02")

	var cards []*InsightCard

	if c := q.revenueTrend(ctx, businessID, now); c != nil {
		cards = append(cards, c)
	}
	if c := q.topProductToday(ctx, businessID, today); c != nil {
		cards = append(cards, c)
	}
	if c := q.peakHour(ctx, businessID, window30Start, today); c != nil {
		cards = append(cards, c)
	}
	if c := q.quietDayWarning(ctx, businessID, today, window30Start); c != nil {
		cards = append(cards, c)
	}

	if cards == nil {
		cards = []*InsightCard{}
	}
	return cards, nil
}

func (q *Queries) revenueTrend(ctx context.Context, businessID uuid.UUID, now time.Time) *InsightCard {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	thisWeekStart := now.AddDate(0, 0, -(weekday - 1))
	lastWeekStart := thisWeekStart.AddDate(0, 0, -7)
	lastWeekEnd := thisWeekStart.AddDate(0, 0, -1)

	var thisRev, lastRev int64
	_ = q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= DATE($2) AND DATE(created_at) <= DATE($3)`,
		businessID, thisWeekStart.Format("2006-01-02"), now.Format("2006-01-02"),
	).Scan(&thisRev)
	_ = q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= DATE($2) AND DATE(created_at) <= DATE($3)`,
		businessID, lastWeekStart.Format("2006-01-02"), lastWeekEnd.Format("2006-01-02"),
	).Scan(&lastRev)

	if lastRev == 0 {
		return nil
	}

	changePct := float64(thisRev-lastRev) / float64(lastRev) * 100
	var direction, insightType string
	if changePct >= 0 {
		direction = fmt.Sprintf("naik %.1f%%", changePct)
		insightType = "trend"
	} else {
		direction = fmt.Sprintf("turun %.1f%%", -changePct)
		insightType = "warning"
	}

	return newInsightCard(
		"revenue-trend", insightType,
		"Revenue minggu ini "+direction,
		fmt.Sprintf("Revenue minggu ini %s dibanding minggu lalu. Tetap pantau tren ini untuk menjaga momentum bisnis.", direction),
		thisWeekStart.Format("2006-01-02"), now.Format("2006-01-02"),
	)
}

func (q *Queries) topProductToday(ctx context.Context, businessID uuid.UUID, today string) *InsightCard {
	var name string
	var qty int
	err := q.db.QueryRowContext(ctx,
		`SELECT oi.product_name, SUM(oi.quantity)
		 FROM order_items oi
		 JOIN orders o ON o.id = oi.order_id
		 WHERE o.business_id = $1 AND o.status = 'CONFIRMED' AND DATE(o.created_at) = $2
		 GROUP BY oi.product_name
		 ORDER BY SUM(oi.quantity) DESC
		 LIMIT 1`,
		businessID, today,
	).Scan(&name, &qty)
	if err != nil || name == "" {
		return nil
	}

	return newInsightCard(
		"top-product-today", "opportunity",
		fmt.Sprintf("%s paling laris hari ini", name),
		fmt.Sprintf("%s terjual %d porsi hari ini. Pastikan stok cukup untuk sisa hari ini.", name, qty),
		today, today,
	)
}

func (q *Queries) peakHour(ctx context.Context, businessID uuid.UUID, startStr, endStr string) *InsightCard {
	var hour, count int
	err := q.db.QueryRowContext(ctx,
		`SELECT EXTRACT(HOUR FROM created_at)::int, COUNT(*)
		 FROM orders
		 WHERE business_id = $1 AND status = 'CONFIRMED'
		   AND DATE(created_at) >= $2 AND DATE(created_at) <= $3
		 GROUP BY EXTRACT(HOUR FROM created_at)
		 ORDER BY COUNT(*) DESC
		 LIMIT 1`,
		businessID, startStr, endStr,
	).Scan(&hour, &count)
	if err != nil || count == 0 {
		return nil
	}

	return newInsightCard(
		"peak-hour", "trend",
		fmt.Sprintf("Jam puncak sekitar pukul %02d.00", hour),
		fmt.Sprintf("Berdasarkan 30 hari terakhir, transaksi paling ramai terjadi sekitar pukul %02d.00–%02d.00. Siapkan staf dan stok sebelum jam tersebut.", hour, hour+1),
		startStr, endStr,
	)
}

func (q *Queries) quietDayWarning(ctx context.Context, businessID uuid.UUID, today, window30Start string) *InsightCard {
	if time.Now().Hour() < 12 {
		return nil
	}

	var count int
	_ = q.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM orders
		 WHERE business_id = $1 AND DATE(created_at) = $2`,
		businessID, today,
	).Scan(&count)

	if count > 0 {
		return nil
	}

	var avg sql.NullFloat64
	_ = q.db.QueryRowContext(ctx,
		`SELECT AVG(daily_count) FROM (
		    SELECT DATE(created_at), COUNT(*) AS daily_count
		    FROM orders
		    WHERE business_id = $1 AND DATE(created_at) >= $2 AND DATE(created_at) < $3
		    GROUP BY DATE(created_at)
		) sub`,
		businessID, window30Start, today,
	).Scan(&avg)

	if !avg.Valid || avg.Float64 < 1 {
		return nil
	}

	return newInsightCard(
		"quiet-day-warning", "warning",
		"Belum ada transaksi hari ini",
		fmt.Sprintf("Rata-rata %.0f transaksi per hari dalam 30 hari terakhir, tapi hari ini belum ada aktivitas. Cek apakah sistem POS berjalan normal.", avg.Float64),
		today, today,
	)
}

func newInsightCard(id, insightType, title, narrative, startDate, endDate string) *InsightCard {
	return &InsightCard{
		ID:        id,
		Type:      insightType,
		Title:     title,
		Narrative: narrative,
		SourceDataWindow: DataWindow{
			StartDate: startDate,
			EndDate:   endDate,
		},
		ModelVersion:    "rule-v1",
		ConfidenceScore: nil,
		UpdatedAt:       time.Now(),
	}
}
