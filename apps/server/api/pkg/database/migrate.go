// platform/database/migrate.go
//
// Simple migration runner berbasis file .sql numbered.
// Cara kerja:
//   1. Buat tabel schema_migrations kalau belum ada.
//   2. Scan semua file di folder migrations/, urutkan ascending.
//   3. Jalankan file yang belum tercatat di schema_migrations.
//   4. Catat setiap file yang berhasil — idempoten, aman dijalankan ulang.
//
// Dipanggil dari main: database.Migrate(db)

package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    filename   VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`

func Migrate(db *sql.DB) error {
	// Pastikan tabel tracking ada
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("migrate: failed to create schema_migrations: %w", err)
	}

	// Ambil daftar migration yang sudah dijalankan
	applied, err := appliedMigrations(db)
	if err != nil {
		return fmt.Errorf("migrate: failed to fetch applied migrations: %w", err)
	}

	// Cari semua file .sql di folder migrations/
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return fmt.Errorf("migrate: failed to glob migration files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		filename := filepath.Base(file)
		if applied[filename] {
			continue
		}

		log.Printf("migrate: applying %s", filename)
		if err := runMigration(db, file, filename); err != nil {
			return fmt.Errorf("migrate: failed on %s: %w", filename, err)
		}
		log.Printf("migrate: applied %s", filename)
	}

	log.Println("migrate: all migrations up to date")
	return nil
}

func appliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT filename FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

func runMigration(db *sql.DB, path, filename string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Jalankan dalam satu transaksi — kalau gagal di tengah, rollback semua
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Jalankan tiap statement dipisah semicolon
	// (exec langsung tidak support multi-statement di lib/pq)
	statements := splitStatements(string(content))
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := tx.Exec(stmt); err != nil {
			tx.Rollback()
			return fmt.Errorf("statement failed: %w\nSQL: %.200s", err, stmt)
		}
	}

	// Catat migration sebagai applied
	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (filename) VALUES ($1)`, filename,
	); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// splitStatements memisahkan SQL berdasarkan semicolon,
// tapi mengabaikan semicolon di dalam string literal sederhana.
func splitStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false

	for i, ch := range sql {
		switch {
		case ch == '\'' && (i == 0 || sql[i-1] != '\\'):
			inSingleQuote = !inSingleQuote
			current.WriteRune(ch)
		case ch == ';' && !inSingleQuote:
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if stmt := strings.TrimSpace(current.String()); stmt != "" {
		statements = append(statements, stmt)
	}
	return statements
}
