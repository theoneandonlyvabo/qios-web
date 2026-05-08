// platform/qmid/generator.go
//
// Generator QM ID untuk businesses. Format: QM-NNNNNN (6 digit, zero-padded).
//
// Race condition safety:
//   Generator wajib dipanggil di dalam *sql.Tx yang sama dengan INSERT businesses,
//   dan transaksi tersebut harus pakai isolation level minimal SERIALIZABLE atau
//   pakai SELECT ... FOR UPDATE pattern. Implementasi di sini pakai
//   "SELECT MAX ... FOR UPDATE" supaya row-level lock mencegah dua tx
//   meng-generate angka yang sama bersamaan.
//
// Catatan: businesses adalah tabel append-only-on-create (qm_id tidak pernah
// di-update setelah create), jadi MAX selalu monotonic.

package qmid

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	prefix     = "QM-"
	digitWidth = 6
)

// Generate menghasilkan QM ID baru di dalam transaksi yang diberikan.
// Caller bertanggung jawab commit/rollback.
func Generate(tx *sql.Tx) (string, error) {
	if tx == nil {
		return "", errors.New("qmid: tx is required")
	}

	// FOR UPDATE memastikan row terbesar di-lock hingga akhir tx.
	// Pakai LIMIT 1 + ORDER BY DESC karena MAX() tidak bisa pakai FOR UPDATE.
	var current sql.NullString
	err := tx.QueryRow(
		`SELECT qm_id FROM businesses
		 ORDER BY qm_id DESC
		 LIMIT 1
		 FOR UPDATE`,
	).Scan(&current)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("qmid: failed to read last qm_id: %w", err)
	}

	next := 1
	if current.Valid && current.String != "" {
		parsed, parseErr := parseSequence(current.String)
		if parseErr != nil {
			return "", parseErr
		}
		next = parsed + 1
	}

	return Format(next), nil
}

// Format mengembalikan QM ID dari angka sequence.
// Exported supaya bisa dipakai di test fixture atau seed data.
func Format(n int) string {
	return fmt.Sprintf("%s%0*d", prefix, digitWidth, n)
}

// parseSequence mem-parse angka di belakang prefix "QM-".
// Tetap berhasil meskipun zero-padding berbeda (mis. QM-1, QM-000001).
func parseSequence(qmID string) (int, error) {
	if !strings.HasPrefix(qmID, prefix) {
		return 0, fmt.Errorf("qmid: invalid format %q (missing prefix)", qmID)
	}
	suffix := strings.TrimPrefix(qmID, prefix)
	if suffix == "" {
		return 0, fmt.Errorf("qmid: invalid format %q (empty suffix)", qmID)
	}
	n, err := strconv.Atoi(suffix)
	if err != nil {
		return 0, fmt.Errorf("qmid: invalid format %q: %w", qmID, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("qmid: invalid format %q (negative)", qmID)
	}
	return n, nil
}
