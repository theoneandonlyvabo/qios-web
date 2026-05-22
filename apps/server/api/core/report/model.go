// core/report/model.go

package report

// DailySalesReport — response GET /reports/daily-sales.
type DailySalesReport struct {
	Date               string             `json:"date"`
	BusinessID         string             `json:"business_id"`
	TotalRevenue       int64              `json:"total_revenue"`
	TotalTransactions  int                `json:"total_transactions"`
	BreakdownByMethod  []*MethodBreakdown `json:"breakdown_by_method"`
	BreakdownByProduct []*TopProduct      `json:"breakdown_by_product"`
}

// MonthlySalesReport — response GET /reports/monthly-sales.
type MonthlySalesReport struct {
	Month              string             `json:"month"`
	BusinessID         string             `json:"business_id"`
	TotalRevenue       int64              `json:"total_revenue"`
	TotalTransactions  int                `json:"total_transactions"`
	DailyBreakdown     []*DailyTrend      `json:"daily_breakdown"`
	BreakdownByMethod  []*MethodBreakdown `json:"breakdown_by_method"`
	BreakdownByProduct []*TopProduct      `json:"breakdown_by_product"`
}

// ConsumptionReport — response GET /reports/consumption.
// Entries kosong di MVP karena belum ada consumption_log / recipe data.
type ConsumptionReport struct {
	StartDate  string             `json:"start_date"`
	EndDate    string             `json:"end_date"`
	BusinessID string             `json:"business_id"`
	Entries    []*ConsumptionEntry `json:"entries"`
}

// ConsumptionEntry — satu ingredient aggregate.
type ConsumptionEntry struct {
	IngredientName string          `json:"ingredient_name"`
	TotalQty       float64         `json:"total_qty"`
	Unit           string          `json:"unit"`
	UsedInProducts []*ProductUsage `json:"used_in_products"`
}

// ProductUsage — produk yang memakai ingredient ini.
type ProductUsage struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	QtyConsumed float64 `json:"qty_consumed"`
}

// MethodBreakdown — aggregate per payment method.
type MethodBreakdown struct {
	Method  string `json:"method"`
	Count   int    `json:"count"`
	Revenue int64  `json:"revenue"`
}

// TopProduct — aggregate per produk.
type TopProduct struct {
	ProductName  string `json:"product_name"`
	QuantitySold int    `json:"quantity_sold"`
	Revenue      int64  `json:"revenue"`
}

// DailyTrend — satu titik data harian.
type DailyTrend struct {
	Date             string `json:"date"`
	TransactionCount int    `json:"transaction_count"`
	Revenue          int64  `json:"revenue"`
}
