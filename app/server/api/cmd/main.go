package main

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/theoneandonlyvabo/qios-web/app/server/api/config"
	"github.com/theoneandonlyvabo/qios-web/app/server/api/core/auth"
	"github.com/theoneandonlyvabo/qios-web/app/server/api/core/dashboard"
	"github.com/theoneandonlyvabo/qios-web/app/server/api/core/operator"
	"github.com/theoneandonlyvabo/qios-web/app/server/api/core/product"
	"github.com/theoneandonlyvabo/qios-web/app/server/api/core/transaction"
	"github.com/theoneandonlyvabo/qios-web/app/server/api/core/user"
	"github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/database"
	appjwt "github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/jwt"
	applogger "github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/logger"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/middleware"
)

type echoValidator struct {
	validate *validator.Validate
}

func (cev *echoValidator) Validate(input any) error {
	return cev.validate.Struct(input)
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

	jwtSvc, err := appjwt.NewService(cfg)
	if err != nil {
		log.Fatalf("failed to init jwt service: %v", err)
	}

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

	authMiddleware := appmiddleware.RequireAuth(jwtSvc)

	authRepo := auth.NewPostgresRepository(db)
	authSvc := auth.NewService(authRepo, jwtSvc)
	auth.RegisterRoutes(e, auth.NewHandler(authSvc))

	user.RegisterRoutes(e, db, authMiddleware)

	operatorRepo := operator.NewPostgresRepository(db)
	operatorPlan := operator.NewPostgresPlanLookup(db)
	operatorSvc := operator.NewService(operatorRepo, operatorPlan, jwtSvc)
	operator.RegisterRoutes(e, operator.NewHandler(operatorSvc), authMiddleware)

	productRepo := product.NewPostgresRepository(db)
	productSvc := product.NewService(productRepo)
	product.RegisterRoutes(e, product.NewHandler(productSvc), authMiddleware)

	transactionRepo := transaction.NewPostgresRepository(db)
	transactionSvc := transaction.NewService(transactionRepo)
	transaction.RegisterRoutes(e, transaction.NewHandler(transactionSvc), authMiddleware)

	dashboard.RegisterRoutes(e, dashboard.NewHandler(dashboard.NewQueries(db)), authMiddleware)

	applogger.Info("server starting on port %s", cfg.AppPort)
	if err := e.Start(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
