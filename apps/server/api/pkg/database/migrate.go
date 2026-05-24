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
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("migrate: failed to create schema_migrations: %w", err)
	}

	applied, err := appliedMigrations(db)
	if err != nil {
		return fmt.Errorf("migrate: failed to fetch applied migrations: %w", err)
	}

	migrationsDir, err := findMigrationsDir()
	if err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("migrate: failed to glob migration files: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("migrate: no .sql files found in %s", migrationsDir)
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

// findMigrationsDir mencari direktori migrations dari CWD dulu,
// lalu fallback ke direktori binary (untuk production deployment).
func findMigrationsDir() (string, error) {
	// 1. Coba dari CWD (go run, local dev)
	if matches, _ := filepath.Glob("migrations/*.sql"); len(matches) > 0 {
		return "migrations", nil
	}

	// 2. Coba relatif ke binary (production)
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Join(filepath.Dir(exe), "migrations")
		if matches, _ := filepath.Glob(filepath.Join(dir, "*.sql")); len(matches) > 0 {
			return dir, nil
		}
	}

	return "", fmt.Errorf("migrate: migrations directory not found (tried CWD and binary dir)")
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

	tx, err := db.Begin()
	if err != nil {
		return err
	}

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

	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (filename) VALUES ($1)`, filename,
	); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// splitStatements memisahkan SQL berdasarkan semicolon.
// Mengabaikan semicolon di dalam string literal dan komentar -- / /* */.
func splitStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false
	inLineComment := false
	inBlockComment := false

	runes := []rune(sql)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			current.WriteRune(ch)
			continue
		}

		if inBlockComment {
			if ch == '*' && i+1 < len(runes) && runes[i+1] == '/' {
				inBlockComment = false
				current.WriteRune(ch)
				i++
				current.WriteRune(runes[i])
			} else {
				current.WriteRune(ch)
			}
			continue
		}

		if !inSingleQuote {
			if ch == '-' && i+1 < len(runes) && runes[i+1] == '-' {
				inLineComment = true
				current.WriteRune(ch)
				continue
			}
			if ch == '/' && i+1 < len(runes) && runes[i+1] == '*' {
				inBlockComment = true
				current.WriteRune(ch)
				continue
			}
		}

		switch {
		case ch == '\'' && (i == 0 || runes[i-1] != '\\'):
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
