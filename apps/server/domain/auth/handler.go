// domain/auth/handler.go
//
// Handler auth — implementasi login, operator login, refresh, logout.
// Refresh token disimpan di httpOnly cookie, bukan response body.
// Access token dikembalikan di response body, disimpen di memory client.

package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
	"golang.org/x/crypto/bcrypt"
)

const refreshTokenCookieName = "refresh_token"

func generateRefreshToken() (plain string, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plain = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(plain))
	hashed = hex.EncodeToString(sum[:])
	return
}

func hashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func storeRefreshToken(db *sql.DB, userID, tokenHash string, expiry time.Duration) error {
	_, err := db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, time.Now().Add(expiry),
	)
	return err
}

func deleteRefreshToken(db *sql.DB, tokenHash string) error {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	return err
}

// setRefreshCookie menempel refresh token ke browser via httpOnly cookie.
// Browser nyimpen otomatis, JS nggak bisa baca.
func setRefreshCookie(c echo.Context, plain string, expiry time.Duration) {
	c.SetCookie(&http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    plain,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(expiry),
		Path:     "/",
	})
}

// clearRefreshCookie menghapus cookie refresh token dari browser.
func clearRefreshCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		Path:     "/",
	})
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// login — owner (users table).
// POST /auth/login
func login(db *sql.DB, _ *config.Config, jwtSvc *jwt.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req loginRequest
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "invalid request body")
		}
		if err := c.Validate(&req); err != nil {
			return response.BadRequest(c, err.Error())
		}

		var (
			userID       string
			passwordHash string
			businessID   string
			isActive     bool
			isSuspended  bool
		)
		err := db.QueryRow(
			`SELECT u.id, u.password_hash, b.id, u.is_active, u.is_suspended
			 FROM users u
			 LEFT JOIN businesses b ON b.user_id = u.id
			 WHERE u.email = $1`,
			req.Email,
		).Scan(&userID, &passwordHash, &businessID, &isActive, &isSuspended)

		if errors.Is(err, sql.ErrNoRows) {
			return response.Unauthorized(c)
		}
		if err != nil {
			return response.Internal(c)
		}

		if !isActive || isSuspended {
			return response.ForbiddenMsg(c, "account is inactive or suspended")
		}

		if passwordHash == "" {
			return response.BadRequest(c, "this account uses Google login")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			return response.Unauthorized(c)
		}

		accessToken, err := jwtSvc.IssueAccessToken(userID, businessID, "owner")
		if err != nil {
			return response.Internal(c)
		}

		plain, hashed, err := generateRefreshToken()
		if err != nil {
			return response.Internal(c)
		}

		if err := storeRefreshToken(db, userID, hashed, jwtSvc.RefreshExpiry()); err != nil {
			return response.Internal(c)
		}

		setRefreshCookie(c, plain, jwtSvc.RefreshExpiry())

		return response.OK(c, map[string]string{"access_token": accessToken})
	}
}

// operatorLogin — kasir (operators table).
// POST /auth/operator/login
func operatorLogin(db *sql.DB, _ *config.Config, jwtSvc *jwt.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req loginRequest
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "invalid request body")
		}
		if err := c.Validate(&req); err != nil {
			return response.BadRequest(c, err.Error())
		}

		var (
			operatorID   string
			businessID   string
			passwordHash string
			isActive     bool
		)
		err := db.QueryRow(
			`SELECT id, business_id, password_hash, is_active
			 FROM operators
			 WHERE email = $1`,
			req.Email,
		).Scan(&operatorID, &businessID, &passwordHash, &isActive)

		if errors.Is(err, sql.ErrNoRows) {
			return response.Unauthorized(c)
		}
		if err != nil {
			return response.Internal(c)
		}

		if !isActive {
			return response.ForbiddenMsg(c, "operator account is inactive")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			return response.Unauthorized(c)
		}

		accessToken, err := jwtSvc.IssueAccessToken(operatorID, businessID, "operator")
		if err != nil {
			return response.Internal(c)
		}

		plain, hashed, err := generateRefreshToken()
		if err != nil {
			return response.Internal(c)
		}

		if err := storeRefreshToken(db, operatorID, hashed, jwtSvc.RefreshExpiry()); err != nil {
			return response.Internal(c)
		}

		setRefreshCookie(c, plain, jwtSvc.RefreshExpiry())

		return response.OK(c, map[string]string{"access_token": accessToken})
	}
}

// googleLogin — owner via Google OAuth.
// POST /auth/google
// TODO: implement post-MVP.
func googleLogin(_ *sql.DB, _ *config.Config, _ *jwt.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		return response.NotImplemented(c, "Google login not yet implemented")
	}
}

// refresh — rotate refresh token.
// POST /auth/refresh
// Baca token dari cookie, bukan body.
func refresh(db *sql.DB, jwtSvc *jwt.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(refreshTokenCookieName)
		if err != nil {
			return response.Unauthorized(c)
		}

		tokenHash := hashToken(cookie.Value)

		var (
			userID    string
			expiresAt time.Time
		)
		err = db.QueryRow(
			`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`,
			tokenHash,
		).Scan(&userID, &expiresAt)

		if errors.Is(err, sql.ErrNoRows) {
			return response.Unauthorized(c)
		}
		if err != nil {
			return response.Internal(c)
		}

		if time.Now().After(expiresAt) {
			_ = deleteRefreshToken(db, tokenHash)
			clearRefreshCookie(c)
			return response.UnauthorizedMsg(c, "sesi telah berakhir, silakan login kembali")
		}

		// Detect role: cek users dulu, fallback ke operators.
		var businessID, role string

		err = db.QueryRow(
			`SELECT b.id FROM users u
			 JOIN businesses b ON b.user_id = u.id
			 WHERE u.id = $1`,
			userID,
		).Scan(&businessID)
		if err == nil {
			role = "owner"
		} else {
			err = db.QueryRow(
				`SELECT business_id FROM operators WHERE id = $1`,
				userID,
			).Scan(&businessID)
			if errors.Is(err, sql.ErrNoRows) {
				return response.Unauthorized(c)
			}
			if err != nil {
				return response.Internal(c)
			}
			role = "operator"
		}

		if err := deleteRefreshToken(db, tokenHash); err != nil {
			return response.Internal(c)
		}

		newAccessToken, err := jwtSvc.IssueAccessToken(userID, businessID, role)
		if err != nil {
			return response.Internal(c)
		}

		newPlain, newHashed, err := generateRefreshToken()
		if err != nil {
			return response.Internal(c)
		}

		if err := storeRefreshToken(db, userID, newHashed, jwtSvc.RefreshExpiry()); err != nil {
			return response.Internal(c)
		}

		setRefreshCookie(c, newPlain, jwtSvc.RefreshExpiry())

		return response.OK(c, map[string]string{"access_token": newAccessToken})
	}
}

// logout — hapus refresh token dari DB dan clear cookie.
// POST /auth/logout
func logout(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(refreshTokenCookieName)
		if err != nil {
			// Cookie sudah tidak ada, anggap sudah logout.
			return response.NoContent(c)
		}

		tokenHash := hashToken(cookie.Value)
		_ = deleteRefreshToken(db, tokenHash)
		clearRefreshCookie(c)

		return response.NoContent(c)
	}
}
