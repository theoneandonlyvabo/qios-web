package operator

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	appjwt "github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
	appmiddleware "github.com/theoneandonlyvabo/qios-web/apps/server/platform/middleware"
)

func newTestJWTService(t *testing.T) *appjwt.Service {
	t.Helper()
	cfg := &config.Config{
		JWTSecret:        "test-secret-operator-handler",
		JWTAccessExpiry:  "1h",
		JWTRefreshExpiry: "720h",
	}
	svc, err := appjwt.NewService(cfg)
	if err != nil {
		t.Fatalf("newTestJWTService: %v", err)
	}
	return svc
}

func newTestEcho(t *testing.T) (*echo.Echo, *appjwt.Service) {
	t.Helper()
	e := echo.New()
	jwtSvc := newTestJWTService(t)
	authMw := appmiddleware.RequireAuth(jwtSvc)
	h := NewHandler(nil)
	RegisterRoutes(e, h, authMw)
	return e, jwtSvc
}

func TestLogout_NoAuthHeader_Returns401(t *testing.T) {
	e, _ := newTestEcho(t)
	req := httptest.NewRequest(http.MethodPost, "/kasir/auth/logout", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestLogout_InvalidToken_Returns401(t *testing.T) {
	e, _ := newTestEcho(t)
	req := httptest.NewRequest(http.MethodPost, "/kasir/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestLogout_OwnerToken_Returns403(t *testing.T) {
	e, jwtSvc := newTestEcho(t)
	ownerToken, err := jwtSvc.IssueAccessToken("user-uuid", "biz-uuid", "owner")
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/kasir/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestLogout_ValidOperatorToken_Returns200(t *testing.T) {
	e, jwtSvc := newTestEcho(t)
	token, err := jwtSvc.IssueOperatorAccessToken("550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("IssueOperatorAccessToken: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/kasir/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
