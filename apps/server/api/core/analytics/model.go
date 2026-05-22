// core/analytics/model.go
//
// Response types untuk domain analytics.
// Analytics = dashboard tapi dengan custom date range dan period comparison.

package analytics

// AnalyticsOverview — response untuk GET /analytics/overview.
type AnalyticsOverview struct {
	CurrentPeriod PeriodData `json:"current_period"`
	Comparison    *CompData  `json:"comparison"`
}

// PeriodData — metrik untuk satu periode.
type PeriodData struct {
	StartDate           string       `json:"start_date"`
	EndDate             string       `json:"end_date"`
	TotalRevenue        int64        `json:"total_revenue"`
	TotalTransactions   int          `json:"total_transactions"`
	AvgTransactionValue int64        `json:"avg_transaction_value"`
	DailyBreakdown      []*DailyTrend `json:"daily_breakdown"`
}

// CompData — metrik periode pembanding + persentase perubahan.
type CompData struct {
	StartDate            string  `json:"start_date"`
	EndDate              string  `json:"end_date"`
	TotalRevenue         int64   `json:"total_revenue"`
	TotalTransactions    int     `json:"total_transactions"`
	RevenueChangePct     float64 `json:"revenue_change_pct"`
	TransactionChangePct float64 `json:"transaction_change_pct"`
}

// DailyTrend — satu titik data harian.
type DailyTrend struct {
	Date             string `json:"date"`
	TransactionCount int    `json:"transaction_count"`
	Revenue          int64  `json:"revenue"`
}

// periodSummary — internal aggregate result.
type periodSummary struct {
	TotalRevenue        int64
	TotalTransactions   int
	AvgTransactionValue int64
}
