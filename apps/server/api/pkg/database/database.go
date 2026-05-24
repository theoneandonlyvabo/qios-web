// File ini bertanggung jawab untuk membuka dan memverifikasi koneksi ke PostgreSQL.
// Dipanggil sekali saat server start. Kalau koneksi gagal, server langsung berhenti —
// tidak ada gunanya server jalan kalau database tidak bisa diakses.
//
// Mengembalikan *sql.DB yang dipakai di seluruh aplikasi sebagai satu koneksi bersama (connection pool).

package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/config"
)

func Connect(cfg *config.Config) *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to reach database: %v", err)
	}

	log.Println("database connected")
	return db
}
