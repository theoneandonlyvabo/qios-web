// cmd/main.go
//
// Entry point server QIOS.
// Urutan startup:
//   1. Load config + validate
//   2. Connect database
//   3. Jalankan migration
//   4. Inisialisasi services (jwt, xendit)
//   5. Setup Echo + middleware global + validator
//   6. Register routes
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
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/operator"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/payment"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/transaction"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/user"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/xendit"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/database"
	appjwt "github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
)

// echoValidator membungkus go-playground/validator supaya kompatibel dengan
// echo.Validator interface. Tanpa ini, c.Validate() akan panic.
type echoValidator struct {
	v *validator.Validate
}

func (ev *echoValidator) Validate(i any) error {
	return ev.v.Struct(i)
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
	e.Validator = &echoValidator{v: validator.New()}

	// Middleware global
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true, // untuk httpOnly cookie refresh token
	}))

	// 6. Register routes
	authMiddleware := appmiddleware.RequireAuth(jwtSvc)

	auth.RegisterRoutes(e, db, cfg, jwtSvc, xenditSvc)
	user.RegisterRoutes(e, db, authMiddleware)
	transaction.RegisterRoutes(e, db, authMiddleware)
	xendit.RegisterRoutes(e, db, cfg, authMiddleware)

	// Operator domain — owner CRUD + kasir login.
	operatorRepo := operator.NewPostgresRepository(db)
	operatorPlan := operator.NewPostgresPlanLookup(db)
	operatorSvc := operator.NewService(operatorRepo, operatorPlan, jwtSvc)
	operator.RegisterRoutes(e, operator.NewHandler(operatorSvc), authMiddleware)

	// 7. Start server
	log.Printf("server starting on port %s", cfg.AppPort)
	if err := e.Start(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
