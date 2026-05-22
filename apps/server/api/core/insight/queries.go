// core/insight/queries.go
//
// Rule-based insight engine. Menghasilkan InsightCard dari data transaksi.
// Setiap rule = satu fungsi yang return *InsightCard (nil kalau tidak relevan).

package insight

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

func (q *Queries) Generate(ctx context.Context, businessID uuid.UUID) ([]*InsightCard, error) {
	now := time.Now()
	today := now.Format("2006-01-02")

	// Window: 30 hari terakhir untuk konteks historis
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

// revenueTrend — bandingkan revenue minggu ini vs minggu lalu.
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
		`SELECT COALESCE(SUM(total_amount), 0) FROM pos_orders
		 WHERE business_id = $1 AND status = 'paid'
		   AND DATE(created_at) >= DATE($2) AND DATE(created_at) <= DATE($3)`,
		businessID, thisWeekStart.Format("2006-01-02"), now.Format("2006-01-02"),
	).Scan(&thisRev)
	_ = q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM pos_orders
		 WHERE business_id = $1 AND status = 'paid'
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

	return newCard(
		"revenue-trend",
		insightType,
		"Revenue minggu ini "+direction,
		fmt.Sprintf("Revenue minggu ini %s dibanding minggu lalu. Tetap pantau tren ini untuk menjaga momentum bisnis.", direction),
		thisWeekStart.Format("2006-01-02"),
		now.Format("2006-01-02"),
	)
}

// topProductToday — produk terlaris hari ini.
func (q *Queries) topProductToday(ctx context.Context, businessID uuid.UUID, today string) *InsightCard {
	var name string
	var qty int
	err := q.db.QueryRowContext(ctx,
		`SELECT oi.product_name, SUM(oi.quantity)
		 FROM pos_order_items oi
		 JOIN pos_orders o ON o.id = oi.pos_order_id
		 WHERE o.business_id = $1 AND o.status = 'paid' AND DATE(o.created_at) = $2
		 GROUP BY oi.product_name
		 ORDER BY SUM(oi.quantity) DESC
		 LIMIT 1`,
		businessID, today,
	).Scan(&name, &qty)
	if err != nil || name == "" {
		return nil
	}

	return newCard(
		"top-product-today",
		"opportunity",
		fmt.Sprintf("%s paling laris hari ini", name),
		fmt.Sprintf("%s terjual %d porsi hari ini. Pastikan stok cukup untuk sisa hari ini.", name, qty),
		today, today,
	)
}

// peakHour — jam tersibuk berdasarkan 30 hari terakhir.
func (q *Queries) peakHour(ctx context.Context, businessID uuid.UUID, startStr, endStr string) *InsightCard {
	var hour, count int
	err := q.db.QueryRowContext(ctx,
		`SELECT EXTRACT(HOUR FROM created_at)::int, COUNT(*)
		 FROM pos_orders
		 WHERE business_id = $1 AND status = 'paid'
		   AND DATE(created_at) >= $2 AND DATE(created_at) <= $3
		 GROUP BY EXTRACT(HOUR FROM created_at)
		 ORDER BY COUNT(*) DESC
		 LIMIT 1`,
		businessID, startStr, endStr,
	).Scan(&hour, &count)
	if err != nil || count == 0 {
		return nil
	}

	return newCard(
		"peak-hour",
		"trend",
		fmt.Sprintf("Jam puncak sekitar pukul %02d.00", hour),
		fmt.Sprintf("Berdasarkan 30 hari terakhir, transaksi paling ramai terjadi sekitar pukul %02d.00–%02d.00. Siapkan staf dan stok sebelum jam tersebut.", hour, hour+1),
		startStr, endStr,
	)
}

// quietDayWarning — peringatan kalau hari ini masih 0 transaksi dan sudah lewat jam 12.
func (q *Queries) quietDayWarning(ctx context.Context, businessID uuid.UUID, today, window30Start string) *InsightCard {
	if time.Now().Hour() < 12 {
		return nil
	}

	var count int
	_ = q.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pos_orders
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
		    FROM pos_orders
		    WHERE business_id = $1 AND DATE(created_at) >= $2 AND DATE(created_at) < $3
		    GROUP BY DATE(created_at)
		) sub`,
		businessID, window30Start, today,
	).Scan(&avg)

	if !avg.Valid || avg.Float64 < 1 {
		return nil
	}

	return newCard(
		"quiet-day-warning",
		"warning",
		"Belum ada transaksi hari ini",
		fmt.Sprintf("Rata-rata %.0f transaksi per hari dalam 30 hari terakhir, tapi hari ini belum ada aktivitas. Cek apakah sistem POS berjalan normal.", avg.Float64),
		today, today,
	)
}

func newCard(id, insightType, title, narrative, startDate, endDate string) *InsightCard {
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
