// cmd/main.go
//
// Entry point server QIOS.
// Urutan startup:
//   1. Load config + validate
//   2. Connect database
//   3. Jalankan migration
//   4. Inisialisasi services (jwt, xendit)
//   5. Setup Echo + middleware global + validator
//   6. Register routes per domain
//   7. Start server

package main

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/auth"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/dashboard"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/operator"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/payment"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/product"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/user"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/database"
	appjwt "github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
)

// echoValidator membungkus ( go-playground/validator ) supaya kompatibel dengan
// echo.Validator interface. Tanpa ini, cev.Validate() akan panic.
type echoValidator struct {
	validate *validator.Validate
}

func (cev *echoValidator) Validate(input any) error {
	return cev.validate.Struct(input)
}

func main() {
	// 1. Load config
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	// 2. Connect database
	db := database.Connect(cfg)
	defer db.Close()

	// 3. Jalankan migration
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	// 4. Inisialisasi services
	jwtSvc, err := appjwt.NewService(cfg)
	if err != nil {
		log.Fatalf("failed to init jwt service: %v", err)
	}

	xenditSvc := payment.NewXenditService(cfg.XenditBaseURL, cfg.XenditSecretKey, http.DefaultClient)

	// 5. Setup Echo
	e := echo.New()
	e.HideBanner = true
	e.Validator = &echoValidator{validate: validator.New()}

	// Middleware global
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true, // untuk httpOnly cookie refresh token
	}))

	// 6. Register routes per domain
	authMiddleware := appmiddleware.RequireAuth(jwtSvc)

	// Auth domain — login, register, refresh, logout, Google OAuth.
	auth.RegisterRoutes(e, db, cfg, jwtSvc, xenditSvc)

	// User domain — profile + business info.
	user.RegisterRoutes(e, db, authMiddleware)

	// Product domain — katalog produk.
	productRepo := product.NewRepository(db)
	productSvc := product.NewService(productRepo)
	product.RegisterRoutes(e, product.NewHandler(productSvc), authMiddleware, appmiddleware.RequireOwner)

	// Operator domain — owner CRUD + kasir login.
	operatorRepo := operator.NewPostgresRepository(db)
	operatorPlan := operator.NewPostgresPlanLookup(db)
	operatorSvc := operator.NewService(operatorRepo, operatorPlan, jwtSvc)
	operator.RegisterRoutes(e, operator.NewHandler(operatorSvc), authMiddleware)

	// Payment domain — POS orders, transaksi, Xendit connect/status.
	// Webhook (POST /payment/xendit/webhook) belum di-register karena butuh
	// XENDIT_WEBHOOK_TOKEN config — wire saat implementasi Xendit integration.
	paymentRepo := payment.NewPostgresRepository(db)
	paymentSvc := payment.NewService(db, paymentRepo, xenditSvc, productSvc)
	payment.RegisterRoutes(e, payment.NewHandler(paymentSvc), authMiddleware)

	// Xendit webhook — public route, verified via x-callback-token.
	if cfg.XenditWebhookToken != "" {
		webhookHandler := payment.NewWebhookHandler(db, paymentRepo, cfg.XenditWebhookToken)
		payment.RegisterWebhookRoute(e, webhookHandler)
	} else {
		log.Println("XENDIT_WEBHOOK_TOKEN not set — /webhooks/xendit disabled")
	}

	// Dashboard domain — placeholder stubs untuk summary, trend, peak hours, top products.
	dashboard.RegisterRoutes(e, dashboard.NewHandler(), authMiddleware)

	// 7. Start server
	log.Printf("server starting on port %s", cfg.AppPort)
	if err := e.Start(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
