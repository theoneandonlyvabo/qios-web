// domain/auth/handler.go
//
// Handler auth — implementasi login, operator login, refresh, logout.
// Login owner: cek tabel users + businesses.
// Login operator: cek tabel operators.
// Refresh: rotate refresh token (hash lama dihapus, hash baru disimpan).
// Logout: hapus refresh token dari DB.

package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
	"golang.org/x/crypto/bcrypt"
)

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

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// POST /auth/login
func login(db *sql.DB, cfg *config.Config, jwtSvc *jwt.Service) echo.HandlerFunc {
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
			return response.UnauthorizedMsg(c, "invalid credentials")
		}
		if err != nil {
			return response.InternalError(c, "database error")
		}

		if !isActive || isSuspended {
			return response.ForbiddenMsg(c, "account is inactive or suspended")
		}

		if passwordHash == "" {
			return response.BadRequest(c, "this account uses Google login")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			return response.UnauthorizedMsg(c, "invalid credentials")
		}

		accessToken, err := jwtSvc.IssueAccessToken(userID, businessID, "owner")
		if err != nil {
			return response.InternalError(c, "failed to issue token")
		}

		plain, hashed, err := generateRefreshToken()
		if err != nil {
			return response.InternalError(c, "failed to generate token")
		}

		if err := storeRefreshToken(db, userID, hashed, jwtSvc.RefreshExpiry()); err != nil {
			return response.InternalError(c, "failed to store token")
		}

		return response.OK(c, authResponse{
			AccessToken:  accessToken,
			RefreshToken: plain,
		})
	}
}

// POST /auth/operator/login
func operatorLogin(db *sql.DB, cfg *config.Config, jwtSvc *jwt.Service) echo.HandlerFunc {
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
			return response.UnauthorizedMsg(c, "invalid credentials")
		}
		if err != nil {
			return response.InternalError(c, "database error")
		}

		if !isActive {
			return response.ForbiddenMsg(c, "operator account is inactive")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			return response.UnauthorizedMsg(c, "invalid credentials")
		}

		accessToken, err := jwtSvc.IssueAccessToken(operatorID, businessID, "operator")
		if err != nil {
			return response.InternalError(c, "failed to issue token")
		}

		plain, hashed, err := generateRefreshToken()
		if err != nil {
			return response.InternalError(c, "failed to generate token")
		}

		if err := storeRefreshToken(db, operatorID, hashed, jwtSvc.RefreshExpiry()); err != nil {
			return response.InternalError(c, "failed to store token")
		}

		return response.OK(c, authResponse{
			AccessToken:  accessToken,
			RefreshToken: plain,
		})
	}
}

// POST /auth/google/login
func googleLogin(db *sql.DB, cfg *config.Config, jwtSvc *jwt.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		return response.NotImplemented(c, "Google login not yet implemented")
	}
}

// POST /auth/refresh
func refresh(db *sql.DB, jwtSvc *jwt.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			RefreshToken string `json:"refresh_token" validate:"required"`
		}
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "invalid request body")
		}
		if err := c.Validate(&req); err != nil {
			return response.BadRequest(c, err.Error())
		}

		tokenHash := hashToken(req.RefreshToken)

		var (
			userID    string
			expiresAt time.Time
		)
		err := db.QueryRow(
			`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`,
			tokenHash,
		).Scan(&userID, &expiresAt)

		if errors.Is(err, sql.ErrNoRows) {
			return response.UnauthorizedMsg(c, "invalid refresh token")
		}
		if err != nil {
			return response.InternalError(c, "database error")
		}

		if time.Now().After(expiresAt) {
			_ = deleteRefreshToken(db, tokenHash)
			return response.UnauthorizedMsg(c, "refresh token expired")
		}

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
				return response.UnauthorizedMsg(c, "user not found")
			}
			if err != nil {
				return response.InternalError(c, "database error")
			}
			role = "operator"
		}

		if err := deleteRefreshToken(db, tokenHash); err != nil {
			return response.InternalError(c, "failed to rotate token")
		}

		newAccessToken, err := jwtSvc.IssueAccessToken(userID, businessID, role)
		if err != nil {
			return response.InternalError(c, "failed to issue token")
		}

		newPlain, newHashed, err := generateRefreshToken()
		if err != nil {
			return response.InternalError(c, "failed to generate token")
		}

		if err := storeRefreshToken(db, userID, newHashed, jwtSvc.RefreshExpiry()); err != nil {
			return response.InternalError(c, "failed to store token")
		}

		return response.OK(c, authResponse{
			AccessToken:  newAccessToken,
			RefreshToken: newPlain,
		})
	}
}

// POST /auth/logout
func logout(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			RefreshToken string `json:"refresh_token" validate:"required"`
		}
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "invalid request body")
		}
		if err := c.Validate(&req); err != nil {
			return response.BadRequest(c, err.Error())
		}

		tokenHash := hashToken(req.RefreshToken)

		if err := deleteRefreshToken(db, tokenHash); err != nil {
			return response.InternalError(c, "failed to logout")
		}

		return response.OK(c, nil)
	}
}
