// platform/qiosid/generator.go
//
// Generator QIOS ID untuk businesses. Format: QIOS-NNNNNN (6 digit, zero-padded).
//
// Race condition safety:
//   Generator wajib dipanggil di dalam *sql.Tx yang sama dengan INSERT businesses,
//   dan transaksi tersebut harus pakai isolation level minimal SERIALIZABLE atau
//   pakai SELECT ... FOR UPDATE pattern. Implementasi di sini pakai
//   "SELECT MAX ... FOR UPDATE" supaya row-level lock mencegah dua tx
//   meng-generate angka yang sama bersamaan.
//
// Catatan: businesses adalah tabel append-only-on-create (qios_id tidak pernah
// di-update setelah create), jadi MAX selalu monotonic.

package qiosid

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	prefix     = "QIOS-"
	digitWidth = 6
)

// Generate menghasilkan QIOS ID baru di dalam transaksi yang diberikan.
// Caller bertanggung jawab commit/rollback.
func Generate(tx *sql.Tx) (string, error) {
	if tx == nil {
		return "", errors.New("qiosid: tx is required")
	}

	// FOR UPDATE memastikan row terbesar di-lock hingga akhir tx.
	// Pakai LIMIT 1 + ORDER BY DESC karena MAX() tidak bisa pakai FOR UPDATE.
	var current sql.NullString
	err := tx.QueryRow(
		`SELECT qios_id FROM businesses
		 ORDER BY qios_id DESC
		 LIMIT 1
		 FOR UPDATE`,
	).Scan(&current)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("qiosid: failed to read last qios_id: %w", err)
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

// Format mengembalikan QIOS ID dari angka sequence.
// Exported supaya bisa dipakai di test fixture atau seed data.
func Format(n int) string {
	return fmt.Sprintf("%s%0*d", prefix, digitWidth, n)
}

// parseSequence mem-parse angka di belakang prefix "QIOS-".
// Tetap berhasil meskipun zero-padding berbeda (mis. QIOS-1, QIOS-0000000001).
func parseSequence(qiosID string) (int, error) {
	if !strings.HasPrefix(qiosID, prefix) {
		return 0, fmt.Errorf("qiosid: invalid format %q (missing prefix)", qiosID)
	}
	suffix := strings.TrimPrefix(qiosID, prefix)
	if suffix == "" {
		return 0, fmt.Errorf("qiosid: invalid format %q (empty suffix)", qiosID)
	}
	n, err := strconv.Atoi(suffix)
	if err != nil {
		return 0, fmt.Errorf("qiosid: invalid format %q: %w", qiosID, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("qiosid: invalid format %q (negative)", qiosID)
	}
	return n, nil
}
