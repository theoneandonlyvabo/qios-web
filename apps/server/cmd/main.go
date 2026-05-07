// cmd/main.go
//
// Entry point server QIOS.
// Urutan startup:
//   1. Load config
//   2. Connect database
//   3. Jalankan migration
//   4. Inisialisasi services
//   5. Setup Echo + middleware global
//   6. Register routes
//   7. Start server

package main

import (
	"log"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/auth"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/payment"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/transaction"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/user"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/database"
	appjwt "github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
)

func main() {
	// 1. Load config
	cfg := config.Load()

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

	// 5. Setup Echo
	e := echo.New()
	e.HideBanner = true

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

	auth.RegisterRoutes(e, db, cfg, jwtSvc)
	user.RegisterRoutes(e, db, authMiddleware)
	transaction.RegisterRoutes(e, db, authMiddleware)
	payment.RegisterRoutes(e, db, cfg, authMiddleware)

	// 7. Start server
	log.Printf("server starting on port %s", cfg.AppPort)
	if err := e.Start(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
