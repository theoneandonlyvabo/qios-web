// core/dashboard/model.go
//
// Response types untuk 4 dashboard endpoints.
// Semua aggregate hanya menghitung order dengan status 'paid'.

package dashboard

import "time"

// ----------------------------------------------------------------
// Period helper
// ----------------------------------------------------------------

// periodRange mengembalikan start dan end untuk query berdasarkan string period.
// Default ke this_month jika period tidak dikenal.
func periodRange(period string) (start, end time.Time) {
	now := time.Now()
	switch period {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = now
	case "this_week":
		// Senin sebagai hari pertama minggu.
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
// Response types
// ----------------------------------------------------------------

// Summary — data untuk GET /dashboard/summary.
type Summary struct {
	Period            string  `json:"period"`
	TotalRevenue      int64   `json:"total_revenue"`
	TotalTransactions int     `json:"total_transactions"`
	AvgOrderValue     float64 `json:"avg_order_value"`
}

// DailyTrend — satu titik data untuk GET /dashboard/transactions/trend.
type DailyTrend struct {
	Date             string `json:"date"`
	TransactionCount int    `json:"transaction_count"`
	Revenue          int64  `json:"revenue"`
}

// TopProduct — satu produk untuk GET /dashboard/products/top.
type TopProduct struct {
	ProductName  string `json:"product_name"`
	QuantitySold int    `json:"quantity_sold"`
	Revenue      int64  `json:"revenue"`
}

// HourlyDistribution — satu slot jam untuk GET /dashboard/transactions/peak-hours.
type HourlyDistribution struct {
	Hour             int   `json:"hour"`
	TransactionCount int   `json:"transaction_count"`
	Revenue          int64 `json:"revenue"`
}
