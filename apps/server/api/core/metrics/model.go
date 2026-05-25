// core/metrics/model.go
//
// Tipe-tipe untuk domain metrics (read-only aggregates).
// Menggabungkan dan de-duplikasi tipe dari dashboard, analytics, report, insight.

package metrics

import "time"

// ----------------------------------------------------------------
// Period helper
// ----------------------------------------------------------------

// periodRange mengembalikan start dan end untuk query berdasarkan string period.
func periodRange(period string) (start, end time.Time) {
	now := time.Now()
	switch period {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = now
	case "this_week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		s := now.AddDate(0, 0, -(weekday - 1))
		start = time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, now.Location())
		end = now
	case "last_month":
		firstThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = firstThisMonth.Add(-time.Nanosecond)
		start = time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, now.Location())
	default: // this_month
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = now
	}
	return
}

// ----------------------------------------------------------------
// Shared types (de-duplicated)
// ----------------------------------------------------------------

// DailyTrend — satu titik data harian. Dipakai oleh summary, overview, dan report.
type DailyTrend struct {
	Date             string `json:"date"`
	TransactionCount int    `json:"transaction_count"`
	Revenue          int64  `json:"revenue"`
}

// TopProduct — aggregate per produk. Dipakai oleh summary dan report.
type TopProduct struct {
	ProductName  string `json:"product_name"`
	QuantitySold int    `json:"quantity_sold"`
	Revenue      int64  `json:"revenue"`
}

// ----------------------------------------------------------------
// Summary types (from dashboard)
// ----------------------------------------------------------------

type Summary struct {
	Period            string  `json:"period"`
	TotalRevenue      int64   `json:"total_revenue"`
	TotalTransactions int     `json:"total_transactions"`
	AvgOrderValue     float64 `json:"avg_order_value"`
}

type HourlyDistribution struct {
	Hour             int   `json:"hour"`
	TransactionCount int   `json:"transaction_count"`
	Revenue          int64 `json:"revenue"`
}

// ----------------------------------------------------------------
// Overview types (from analytics)
// ----------------------------------------------------------------

type AnalyticsOverview struct {
	CurrentPeriod PeriodData `json:"current_period"`
	Comparison    *CompData  `json:"comparison"`
}

type PeriodData struct {
	StartDate           string       `json:"start_date"`
	EndDate             string       `json:"end_date"`
	TotalRevenue        int64        `json:"total_revenue"`
	TotalTransactions   int          `json:"total_transactions"`
	AvgTransactionValue int64        `json:"avg_transaction_value"`
	DailyBreakdown      []*DailyTrend `json:"daily_breakdown"`
}

type CompData struct {
	StartDate            string  `json:"start_date"`
	EndDate              string  `json:"end_date"`
	TotalRevenue         int64   `json:"total_revenue"`
	TotalTransactions    int     `json:"total_transactions"`
	RevenueChangePct     float64 `json:"revenue_change_pct"`
	TransactionChangePct float64 `json:"transaction_change_pct"`
}

// ----------------------------------------------------------------
// Report types (from report)
// ----------------------------------------------------------------

type DailySalesReport struct {
	Date               string             `json:"date"`
	BusinessID         string             `json:"business_id"`
	TotalRevenue       int64              `json:"total_revenue"`
	TotalTransactions  int                `json:"total_transactions"`
	BreakdownByMethod  []*MethodBreakdown `json:"breakdown_by_method"`
	BreakdownByProduct []*TopProduct      `json:"breakdown_by_product"`
}

type MonthlySalesReport struct {
	Month              string             `json:"month"`
	BusinessID         string             `json:"business_id"`
	TotalRevenue       int64              `json:"total_revenue"`
	TotalTransactions  int                `json:"total_transactions"`
	DailyBreakdown     []*DailyTrend      `json:"daily_breakdown"`
	BreakdownByMethod  []*MethodBreakdown `json:"breakdown_by_method"`
	BreakdownByProduct []*TopProduct      `json:"breakdown_by_product"`
}

type ConsumptionReport struct {
	StartDate  string             `json:"start_date"`
	EndDate    string             `json:"end_date"`
	BusinessID string             `json:"business_id"`
	Entries    []*ConsumptionEntry `json:"entries"`
}

type ConsumptionEntry struct {
	IngredientName string          `json:"ingredient_name"`
	TotalQty       float64         `json:"total_qty"`
	Unit           string          `json:"unit"`
	UsedInProducts []*ProductUsage `json:"used_in_products"`
}

type ProductUsage struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	QtyConsumed float64 `json:"qty_consumed"`
}

type MethodBreakdown struct {
	Method  string `json:"method"`
	Count   int    `json:"count"`
	Revenue int64  `json:"revenue"`
}

// ----------------------------------------------------------------
// Insight types (from insight)
// ----------------------------------------------------------------

// InsightCard — satu kartu insight rule-based.
// Schema forward-compat: confidence_score null di MVP, diisi saat switch ke AI engine.
type InsightCard struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"`
	Title           string     `json:"title"`
	Narrative       string     `json:"narrative"`
	SourceDataWindow DataWindow `json:"source_data_window"`
	ModelVersion    string     `json:"model_version"`
	ConfidenceScore *float64   `json:"confidence_score"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type DataWindow struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
