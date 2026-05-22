// core/insight/model.go

package insight

import "time"

// InsightCard — satu kartu insight rule-based.
// Schema forward-compat: confidence_score null di MVP, diisi saat switch ke AI engine.
type InsightCard struct {
	ID              string        `json:"id"`
	Type            string        `json:"type"`
	Title           string        `json:"title"`
	Narrative       string        `json:"narrative"`
	SourceDataWindow DataWindow   `json:"source_data_window"`
	ModelVersion    string        `json:"model_version"`
	ConfidenceScore *float64      `json:"confidence_score"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// DataWindow — periode data yang menghasilkan insight ini.
type DataWindow struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
