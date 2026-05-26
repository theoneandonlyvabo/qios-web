package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// Rate limit values are package-level constants (MVP).
// Move to env vars (RATE_LIMIT_OWNER_RPM, etc.) post-MVP once production
// traffic patterns are observed.
func newUserIDRateLimiter(rps float64, burst int) echo.MiddlewareFunc {
	return echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
		IdentifierExtractor: func(c echo.Context) (string, error) {
			id, _ := c.Get("user_id").(string)
			if id == "" {
				id = c.RealIP()
			}
			return id, nil
		},
		Store: echomiddleware.NewRateLimiterMemoryStoreWithConfig(
			echomiddleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(rps),
				Burst:     burst,
				ExpiresIn: 5 * time.Minute,
			},
		),
	})
}

// RateLimitOwner: 100 req/min per user_id for owner-scoped routes.
var RateLimitOwner = newUserIDRateLimiter(100.0/60.0, 20)

// RateLimitOperator: 60 req/min per user_id for operator-scoped routes.
var RateLimitOperator = newUserIDRateLimiter(60.0/60.0, 15)

// RateLimitAdmin: 200 req/min per user_id for admin routes.
var RateLimitAdmin = newUserIDRateLimiter(200.0/60.0, 30)
