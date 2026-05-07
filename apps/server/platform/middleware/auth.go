// platform/middleware/auth.go
//
// Middleware autentikasi dan otorisasi untuk semua protected routes.
//
// RequireAuth     — verifikasi Bearer JWT, inject claims ke context
// RequireOwner    — pastikan role == "owner"
// RequireOperator — pastikan role == "owner" atau "operator"
//
// Claims yang diinject ke context:
//   "user_id"      string
//   "operator_id"  string  (kosong kalau role != "operator")
//   "business_id"  string
//   "role"         string  ("owner" | "operator")

package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
)

// RequireAuth memverifikasi Bearer JWT dan inject claims ke context.
func RequireAuth(jwtSvc *jwt.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				return response.Unauthorized(c)
			}

			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := jwtSvc.VerifyAccessToken(token)
			if err != nil {
				return response.Unauthorized(c)
			}

			c.Set("user_id", claims.UserID)
			c.Set("operator_id", claims.OperatorID)
			c.Set("business_id", claims.BusinessID)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}

// RequireOwner memastikan user yang login adalah owner.
// Wajib dipanggil setelah RequireAuth.
func RequireOwner(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role, _ := c.Get("role").(string)
		if role != "owner" {
			return response.Forbidden(c)
		}
		return next(c)
	}
}

// RequireOperator memastikan user yang login adalah owner atau operator.
// Wajib dipanggil setelah RequireAuth.
func RequireOperator(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role, _ := c.Get("role").(string)
		if role != "owner" && role != "operator" {
			return response.Forbidden(c)
		}
		return next(c)
	}
}
