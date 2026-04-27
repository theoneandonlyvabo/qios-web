// File ini bertanggung jawab untuk membaca semua environment variable dan
// mengumpulkannya dalam satu struct Config yang bisa dipakai di seluruh aplikasi.
//
// Dipanggil sekali saat server start via config.Load().
// Semua bagian aplikasi yang butuh konfigurasi menerima *Config sebagai parameter,
// bukan membaca os.Getenv() langsung — supaya perubahan config cukup di satu tempat.

package config

import (
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

	JWTSecret         string
	JWTAccessExpiry   string
	JWTRefreshExpiry  string

	MidtransServerKey string
	MidtransEnv       string
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

		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransEnv:       getEnv("MIDTRANS_ENV", "sandbox"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}