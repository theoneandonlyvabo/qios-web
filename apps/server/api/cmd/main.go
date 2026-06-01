package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/theoneandonlyvabo/qios-web/apps/server/api/config"
	adminpkg "github.com/theoneandonlyvabo/qios-web/apps/server/api/core/admin"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/core/auth"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/core/metrics"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/core/order"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/core/product"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/core/transaction"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/core/user"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/database"
	appjwt "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/jwt"
	applogger "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/logger"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/middleware"
)

type echoValidator struct {
	validate *validator.Validate
}

func cookieSameSite(s string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteDefaultMode
	}
}

func (cev *echoValidator) Validate(input any) error {
	return cev.validate.Struct(input)
}

// metricsGuard hanya izinkan akses dari localhost.
func metricsGuard() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if ip != "127.0.0.1" && ip != "::1" {
				return echo.ErrForbidden
			}
			return next(c)
		}
	}
}

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	db := database.Connect(cfg)
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Bersihkan refresh token yang sudah expired setiap 24 jam.
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, _ = db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
				_, _ = db.Exec(`DELETE FROM admin_refresh_tokens WHERE expires_at < NOW()`)
			case <-ctx.Done():
				return
			}
		}
	}()

	jwtSvc, err := appjwt.NewService(cfg)
	if err != nil {
		log.Fatalf("failed to init jwt service: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Validator = &echoValidator{validate: validator.New()}

	// Recover harus pertama — tangkap panic dari semua handler termasuk /health.
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.BodyLimit("2M"))
	e.Use(echomiddleware.TimeoutWithConfig(echomiddleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))
	e.IPExtractor = echo.ExtractIPFromXFFHeader()
	e.Use(applogger.Middleware())
	e.Use(echomiddleware.SecureWithConfig(echomiddleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
	}))
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     strings.Split(cfg.CORSAllowedOrigins, ","),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Admin-Key"},
		AllowCredentials: true,
	}))

	e.Use(echoprometheus.NewMiddleware("qios"))
	e.GET("/metrics", echoprometheus.NewHandler(), metricsGuard())
	e.GET("/livez", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	readyz := func(c echo.Context) error {
		if err := db.PingContext(c.Request().Context()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "db_unreachable"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
	e.GET("/readyz", readyz)
	e.GET("/health", readyz) // alias for /readyz — remove once infra confirms unused

	authMiddleware := appmiddleware.RequireAuth(jwtSvc)

	authRepo := auth.NewPostgresRepository(db)
	authSvc := auth.NewService(authRepo, jwtSvc)
	auth.RegisterRoutes(e, auth.NewHandler(authSvc, auth.CookieConfig{
		Secure:   cfg.CookieSecure,
		SameSite: cookieSameSite(cfg.CookieSameSite),
		Domain:   cfg.CookieDomain,
	}))

	userRepo := user.NewPostgresRepository(db)
	userPlan := user.NewPostgresPlanLookup(db)
	userSvc := user.NewService(userRepo, userPlan)
	user.RegisterRoutes(e, user.NewHandler(userSvc), authMiddleware)

	orderRepo := order.NewPostgresRepository(db)
	orderSvc := order.NewService(orderRepo)
	order.RegisterRoutes(e, order.NewHandler(orderSvc), authMiddleware)

	productRepo := product.NewPostgresRepository(db)
	productSvc := product.NewService(productRepo)
	product.RegisterRoutes(e, product.NewHandler(productSvc), authMiddleware)

	transactionRepo := transaction.NewPostgresRepository(db)
	transactionSvc := transaction.NewService(transactionRepo)
	transaction.RegisterRoutes(e, transaction.NewHandler(transactionSvc), authMiddleware)

	metrics.RegisterRoutes(e, metrics.NewHandler(metrics.NewQueries(db)), authMiddleware)

	requireAdminKey := appmiddleware.RequireAdminKey(cfg.AdminAPIKey)
	adminRepo := adminpkg.NewPostgresRepository(db)
	adminSvc := adminpkg.NewService(adminRepo)
	adminpkg.RegisterRoutes(e, adminpkg.NewHandler(adminSvc), requireAdminKey)

	go func() {
		applogger.Info("server starting on port %s", cfg.AppPort)
		if err := e.Start(":" + cfg.AppPort); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	applogger.Info("server stopped gracefully")
}
