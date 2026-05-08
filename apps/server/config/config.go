// File ini bertanggung jawab untuk membaca semua environment variable dan
// mengumpulkannya dalam satu struct Config yang bisa dipakai di seluruh aplikasi.
//
// Dipanggil sekali saat server start via config.Load().
// Semua bagian aplikasi yang butuh konfigurasi menerima *Config sebagai parameter,
// bukan membaca os.Getenv() langsung — supaya perubahan config cukup di satu tempat.

package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret        string
	JWTAccessExpiry  string
	JWTRefreshExpiry string

	// Xendit (xenPlatform). XenditSecretKey adalah master secret QIOS yang dipakai
	// untuk membuat sub-account dan operasi platform-level lain. Setiap sub-account
	// punya api_key/secret_key sendiri yang disimpan di tabel businesses.
	XenditSecretKey string
	XenditEnv       string // "sandbox" atau "production"
	XenditBaseURL   string // override opsional, default https://api.xendit.io
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from system environment")
	}

	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "qios"),

		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTAccessExpiry:  getEnv("JWT_ACCESS_EXPIRY", "15m"),
		JWTRefreshExpiry: getEnv("JWT_REFRESH_EXPIRY", "720h"),

		XenditSecretKey: getEnv("XENDIT_SECRET_KEY", ""),
		XenditEnv:       getEnv("XENDIT_ENV", "sandbox"),
		XenditBaseURL:   getEnv("XENDIT_BASE_URL", "https://api.xendit.io"),
	}
}

// Validate memeriksa env var yang wajib ada saat startup.
// Dipanggil di main setelah Load — gagal cepat lebih baik daripada panic di runtime.
func (c *Config) Validate() error {
	var missing []string
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.DBPassword == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if c.XenditSecretKey == "" {
		missing = append(missing, "XENDIT_SECRET_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("config: required env vars not set: %v", missing)
	}
	if c.XenditEnv != "sandbox" && c.XenditEnv != "production" {
		return errors.New(`config: XENDIT_ENV must be "sandbox" or "production"`)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
