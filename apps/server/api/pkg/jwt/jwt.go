// platform/jwt/jwt.go
//
// Service untuk issue dan verify JWT.
// Access token payload: user_id, business_id, role, exp.
// Dipanggil dari auth handler (issue) dan auth middleware (verify).

package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/config"
)

type Service struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// Role constants — dipakai oleh middleware untuk gating.
const (
	RoleOwner    = "owner"
	RoleOperator = "operator"
	RoleAdmin    = "admin"
)

type Claims struct {
	UserID     string `json:"user_id"`
	OperatorID string `json:"operator_id,omitempty"`
	BusinessID string `json:"business_id"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

func NewService(cfg *config.Config) (*Service, error) {
	accessExpiry, err := time.ParseDuration(cfg.JWTAccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("jwt: invalid JWT_ACCESS_EXPIRY: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(cfg.JWTRefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("jwt: invalid JWT_REFRESH_EXPIRY: %w", err)
	}

	return &Service{
		secret:        []byte(cfg.JWTSecret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

// IssueAccessToken membuat access token JWT untuk user.
// Dipakai owner login dan refresh — role bebas, OperatorID kosong.
func (s *Service) IssueAccessToken(userID, businessID, role string) (string, error) {
	claims := Claims{
		UserID:     userID,
		BusinessID: businessID,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// IssueOperatorAccessToken membuat access token khusus operator.
// UserID diisi operatorID juga — kompatibel dengan refresh handler yang
// melakukan lookup berbasis user_id (mengikuti pola lama).
// OperatorID di-set eksplisit supaya handler bisa baca lewat context.
func (s *Service) IssueOperatorAccessToken(operatorID, businessID string) (string, error) {
	claims := Claims{
		UserID:     operatorID,
		OperatorID: operatorID,
		BusinessID: businessID,
		Role:       RoleOperator,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// VerifyAccessToken memverifikasi dan mem-parse access token JWT.
func (s *Service) VerifyAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt: failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("jwt: invalid token claims")
	}

	return claims, nil
}

// RefreshExpiry mengembalikan durasi refresh token untuk dipakai saat generate token.
func (s *Service) RefreshExpiry() time.Duration {
	return s.refreshExpiry
}
