// cmd/main.go
//
// Entry point server QIOS.
// Urutan startup:
//   1. Load config + validate
//   2. Init crypto (enkripsi key untuk data sensitif)
//   3. Connect database
//   4. Jalankan migration
//   5. Inisialisasi services (jwt, xendit)
//   6. Setup Echo + middleware global + validator
//   7. Register routes per domain
//   8. Start server

package main

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/auth"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/dashboard"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/operator"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/payment"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/product"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/user"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/crypto"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/database"
	appjwt "github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
	applogger "github.com/theoneandonlyvabo/qios-web/apps/server/platform/logger"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
)

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

	// 2. Init crypto
	if err := crypto.Init(cfg.EncryptionKey); err != nil {
		log.Fatalf("failed to init crypto: %v", err)
	}

	// 3. Connect database
	db := database.Connect(cfg)
	defer db.Close()

	// 4. Jalankan migration
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	// 5. Inisialisasi services
	jwtSvc, err := appjwt.NewService(cfg)
	if err != nil {
		log.Fatalf("failed to init jwt service: %v", err)
	}

	xenditSvc := payment.NewXenditService(
		cfg.XenditBaseURL,
		cfg.XenditSecretKey,
		cfg.XenditCallbackURL,
		http.DefaultClient,
	)

	// 6. Setup Echo
	e := echo.New()
	e.HideBanner = true
	e.Validator = &echoValidator{validate: validator.New()}

	e.Use(applogger.Middleware())
	e.Use(echoprometheus.NewMiddleware("qios"))
	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 7. Register routes per domain
	authMiddleware := appmiddleware.RequireAuth(jwtSvc)

	authRepo := auth.NewPostgresRepository(db)
	authSvc := auth.NewService(db, authRepo, jwtSvc, xenditSvc)
	auth.RegisterRoutes(e, auth.NewHandler(authSvc))

	user.RegisterRoutes(e, db, authMiddleware)

	productRepo := product.NewRepository(db)
	productSvc := product.NewService(productRepo)
	product.RegisterRoutes(e, product.NewHandler(productSvc), authMiddleware, appmiddleware.RequireOwner)

	operatorRepo := operator.NewPostgresRepository(db)
	operatorPlan := operator.NewPostgresPlanLookup(db)
	operatorSvc := operator.NewService(operatorRepo, operatorPlan, jwtSvc)
	operator.RegisterRoutes(e, operator.NewHandler(operatorSvc), authMiddleware)

	paymentRepo := payment.NewPostgresRepository(db)
	paymentSvc := payment.NewService(db, paymentRepo, xenditSvc, productSvc)
	payment.RegisterRoutes(e, payment.NewHandler(paymentSvc), authMiddleware)

	if cfg.XenditWebhookToken != "" {
		webhookHandler := payment.NewWebhookHandler(db, paymentRepo, cfg.XenditWebhookToken)
		payment.RegisterWebhookRoute(e, webhookHandler)
	} else {
		applogger.Warn("XENDIT_WEBHOOK_TOKEN not set — /webhooks/xendit disabled")
	}

	dashboard.RegisterRoutes(e, dashboard.NewHandler(), authMiddleware)

	// 8. Start server
	applogger.Info("server starting on port %s", cfg.AppPort)
	if err := e.Start(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
