package operator

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/theoneandonlyvabo/qios-web/apps/server/api/config"
	appjwt "github.com/theoneandonlyvabo/qios-web/apps/server/api/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

// ----------------------------------------------------------------
// Mock helpers
// ----------------------------------------------------------------

type mockRepo struct {
	createFn                  func(ctx context.Context, op *Operator) error
	findByIDFn                func(ctx context.Context, id uuid.UUID) (*Operator, error)
	findByBusinessIDFn        func(ctx context.Context, businessID uuid.UUID) ([]*Operator, error)
	findByOperatorCodeFn      func(ctx context.Context, businessID uuid.UUID, code string) (*Operator, error)
	findByQRTokenFn           func(ctx context.Context, token string) (*Operator, error)
	countActiveByBusinessIDFn func(ctx context.Context, businessID uuid.UUID) (int, error)
	updateFn                  func(ctx context.Context, op *Operator) error
	softDeleteFn              func(ctx context.Context, id uuid.UUID) error
	regenerateQRFn            func(ctx context.Context, id uuid.UUID, newToken string) error
}

func (m *mockRepo) Create(ctx context.Context, op *Operator) error {
	return m.createFn(ctx, op)
}
func (m *mockRepo) FindByID(ctx context.Context, id uuid.UUID) (*Operator, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockRepo) FindByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*Operator, error) {
	return m.findByBusinessIDFn(ctx, businessID)
}
func (m *mockRepo) FindByOperatorCode(ctx context.Context, businessID uuid.UUID, code string) (*Operator, error) {
	return m.findByOperatorCodeFn(ctx, businessID, code)
}
func (m *mockRepo) FindByQRToken(ctx context.Context, token string) (*Operator, error) {
	return m.findByQRTokenFn(ctx, token)
}
func (m *mockRepo) CountActiveByBusinessID(ctx context.Context, businessID uuid.UUID) (int, error) {
	return m.countActiveByBusinessIDFn(ctx, businessID)
}
func (m *mockRepo) Update(ctx context.Context, op *Operator) error {
	return m.updateFn(ctx, op)
}
func (m *mockRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return m.softDeleteFn(ctx, id)
}
func (m *mockRepo) RegenerateQR(ctx context.Context, id uuid.UUID, newToken string) error {
	return m.regenerateQRFn(ctx, id, newToken)
}

type mockPlan struct {
	maxFn func(ctx context.Context, businessID uuid.UUID) (int, error)
}

func (m *mockPlan) MaxOperators(ctx context.Context, businessID uuid.UUID) (int, error) {
	return m.maxFn(ctx, businessID)
}

func newServiceForTest(t *testing.T, repo Repository, plan PlanLookup) Service {
	t.Helper()
	cfg := &config.Config{
		JWTSecret:        "test-secret-operator-service",
		JWTAccessExpiry:  "1h",
		JWTRefreshExpiry: "720h",
	}
	jwtSvc, err := appjwt.NewService(cfg)
	if err != nil {
		t.Fatalf("newServiceForTest jwt: %v", err)
	}
	return NewService(repo, plan, jwtSvc)
}

func hashForTest(t *testing.T, password string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hashForTest: %v", err)
	}
	return string(h)
}

// ----------------------------------------------------------------
// CreateOperator
// ----------------------------------------------------------------

func TestCreateOperator_Success(t *testing.T) {
	bizID := uuid.New()
	repo := &mockRepo{
		countActiveByBusinessIDFn: func(_ context.Context, _ uuid.UUID) (int, error) { return 1, nil },
		createFn: func(_ context.Context, op *Operator) error {
			op.ID = uuid.New()
			return nil
		},
	}
	plan := &mockPlan{
		maxFn: func(_ context.Context, _ uuid.UUID) (int, error) { return 3, nil },
	}
	svc := newServiceForTest(t, repo, plan)

	got, err := svc.CreateOperator(context.Background(), bizID, CreateOperatorRequest{
		Name:         "Kasir Satu",
		OperatorCode: "kasir-01",
		Password:     "secret123",
	})

	if err != nil {
		t.Fatalf("CreateOperator err = %v, want nil", err)
	}
	if got.QRToken == "" {
		t.Error("QRToken empty — should be returned once on create")
	}
	if got.Name != "Kasir Satu" {
		t.Errorf("Name = %q, want Kasir Satu", got.Name)
	}
}

func TestCreateOperator_LimitReached(t *testing.T) {
	bizID := uuid.New()
	repo := &mockRepo{
		countActiveByBusinessIDFn: func(_ context.Context, _ uuid.UUID) (int, error) { return 3, nil },
	}
	plan := &mockPlan{
		maxFn: func(_ context.Context, _ uuid.UUID) (int, error) { return 3, nil },
	}
	svc := newServiceForTest(t, repo, plan)

	_, err := svc.CreateOperator(context.Background(), bizID, CreateOperatorRequest{
		Name: "Kasir Baru", OperatorCode: "kasir-04", Password: "secret123",
	})
	if !errors.Is(err, ErrLimitReached) {
		t.Errorf("err = %v, want ErrLimitReached", err)
	}
}

func TestCreateOperator_CodeTaken(t *testing.T) {
	bizID := uuid.New()
	repo := &mockRepo{
		countActiveByBusinessIDFn: func(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil },
		createFn:                  func(_ context.Context, _ *Operator) error { return ErrCodeTaken },
	}
	plan := &mockPlan{maxFn: func(_ context.Context, _ uuid.UUID) (int, error) { return 3, nil }}
	svc := newServiceForTest(t, repo, plan)

	_, err := svc.CreateOperator(context.Background(), bizID, CreateOperatorRequest{
		Name: "Kasir Dua", OperatorCode: "kasir-01", Password: "secret123",
	})
	if !errors.Is(err, ErrCodeTaken) {
		t.Errorf("err = %v, want ErrCodeTaken", err)
	}
}

func TestCreateOperator_UnlimitedPlan(t *testing.T) {
	bizID := uuid.New()
	repo := &mockRepo{
		countActiveByBusinessIDFn: func(_ context.Context, _ uuid.UUID) (int, error) { return 100, nil },
		createFn:                  func(_ context.Context, op *Operator) error { op.ID = uuid.New(); return nil },
	}
	plan := &mockPlan{maxFn: func(_ context.Context, _ uuid.UUID) (int, error) { return -1, nil }}
	svc := newServiceForTest(t, repo, plan)

	_, err := svc.CreateOperator(context.Background(), bizID, CreateOperatorRequest{
		Name: "Kasir Seratus", OperatorCode: "kasir-100", Password: "secret123",
	})
	if err != nil {
		t.Errorf("err = %v, want nil (unlimited plan)", err)
	}
}

// ----------------------------------------------------------------
// LoginWithCredentials
// ----------------------------------------------------------------

func TestLoginWithCredentials_Success(t *testing.T) {
	bizID := uuid.New()
	opID := uuid.New()
	pass := "correct-password"
	repo := &mockRepo{
		findByOperatorCodeFn: func(_ context.Context, _ uuid.UUID, _ string) (*Operator, error) {
			return &Operator{ID: opID, BusinessID: bizID, PasswordHash: hashForTest(t, pass), IsActive: true}, nil
		},
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	got, err := svc.LoginWithCredentials(context.Background(), bizID, OperatorLoginRequest{
		OperatorCode: "kasir-01", Password: pass,
	})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if got.AccessToken == "" {
		t.Error("AccessToken empty")
	}
	if got.Operator.ID != opID {
		t.Errorf("Operator.ID = %v, want %v", got.Operator.ID, opID)
	}
}

func TestLoginWithCredentials_WrongPassword(t *testing.T) {
	bizID := uuid.New()
	repo := &mockRepo{
		findByOperatorCodeFn: func(_ context.Context, _ uuid.UUID, _ string) (*Operator, error) {
			return &Operator{ID: uuid.New(), BusinessID: bizID, PasswordHash: hashForTest(t, "correct"), IsActive: true}, nil
		},
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	_, err := svc.LoginWithCredentials(context.Background(), bizID, OperatorLoginRequest{
		OperatorCode: "kasir-01", Password: "wrong",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginWithCredentials_OperatorNotFound(t *testing.T) {
	bizID := uuid.New()
	repo := &mockRepo{
		findByOperatorCodeFn: func(_ context.Context, _ uuid.UUID, _ string) (*Operator, error) {
			return nil, ErrNotFound
		},
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	_, err := svc.LoginWithCredentials(context.Background(), bizID, OperatorLoginRequest{
		OperatorCode: "ghost", Password: "any",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials (not ErrNotFound)", err)
	}
}

func TestLoginWithCredentials_InactiveOperator(t *testing.T) {
	bizID := uuid.New()
	repo := &mockRepo{
		findByOperatorCodeFn: func(_ context.Context, _ uuid.UUID, _ string) (*Operator, error) {
			return &Operator{ID: uuid.New(), BusinessID: bizID, PasswordHash: hashForTest(t, "correct"), IsActive: false}, nil
		},
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	_, err := svc.LoginWithCredentials(context.Background(), bizID, OperatorLoginRequest{
		OperatorCode: "kasir-01", Password: "correct",
	})
	if !errors.Is(err, ErrInactive) {
		t.Errorf("err = %v, want ErrInactive", err)
	}
}

// ----------------------------------------------------------------
// LoginWithQR
// ----------------------------------------------------------------

func TestLoginWithQR_Success(t *testing.T) {
	bizID := uuid.New()
	opID := uuid.New()
	qrToken := "valid-qr-token"
	repo := &mockRepo{
		findByQRTokenFn: func(_ context.Context, token string) (*Operator, error) {
			if token != qrToken {
				return nil, ErrNotFound
			}
			return &Operator{ID: opID, BusinessID: bizID, IsActive: true}, nil
		},
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	got, err := svc.LoginWithQR(context.Background(), QRLoginRequest{QRToken: qrToken})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if got.AccessToken == "" {
		t.Error("AccessToken empty")
	}
}

func TestLoginWithQR_InvalidToken(t *testing.T) {
	repo := &mockRepo{
		findByQRTokenFn: func(_ context.Context, _ string) (*Operator, error) { return nil, ErrNotFound },
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	_, err := svc.LoginWithQR(context.Background(), QRLoginRequest{QRToken: "bad-token"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginWithQR_InactiveOperator(t *testing.T) {
	repo := &mockRepo{
		findByQRTokenFn: func(_ context.Context, _ string) (*Operator, error) {
			return &Operator{ID: uuid.New(), IsActive: false}, nil
		},
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	_, err := svc.LoginWithQR(context.Background(), QRLoginRequest{QRToken: "some-token"})
	if !errors.Is(err, ErrInactive) {
		t.Errorf("err = %v, want ErrInactive", err)
	}
}

// ----------------------------------------------------------------
// RegenerateQR
// ----------------------------------------------------------------

func TestRegenerateQR_Success(t *testing.T) {
	bizID := uuid.New()
	opID := uuid.New()
	var capturedNewToken string
	repo := &mockRepo{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*Operator, error) {
			return &Operator{ID: opID, BusinessID: bizID, IsActive: true}, nil
		},
		regenerateQRFn: func(_ context.Context, _ uuid.UUID, newToken string) error {
			capturedNewToken = newToken
			return nil
		},
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	got, err := svc.RegenerateQR(context.Background(), bizID, opID)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if got.QRToken == "" {
		t.Error("QRToken empty on regenerate")
	}
	if got.QRToken != capturedNewToken {
		t.Errorf("returned QRToken %q != stored token %q", got.QRToken, capturedNewToken)
	}
}

func TestRegenerateQR_WrongBusiness(t *testing.T) {
	opID := uuid.New()
	repo := &mockRepo{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*Operator, error) {
			return &Operator{ID: opID, BusinessID: uuid.New()}, nil
		},
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	_, err := svc.RegenerateQR(context.Background(), uuid.New(), opID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestRegenerateQR_NotFound(t *testing.T) {
	repo := &mockRepo{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*Operator, error) { return nil, ErrNotFound },
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	_, err := svc.RegenerateQR(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// ----------------------------------------------------------------
// DeleteOperator
// ----------------------------------------------------------------

func TestDeleteOperator_Success(t *testing.T) {
	bizID := uuid.New()
	opID := uuid.New()
	repo := &mockRepo{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*Operator, error) {
			return &Operator{ID: opID, BusinessID: bizID}, nil
		},
		softDeleteFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	if err := svc.DeleteOperator(context.Background(), bizID, opID); err != nil {
		t.Errorf("err = %v, want nil", err)
	}
}

func TestDeleteOperator_WrongBusiness(t *testing.T) {
	opID := uuid.New()
	repo := &mockRepo{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*Operator, error) {
			return &Operator{ID: opID, BusinessID: uuid.New()}, nil
		},
	}
	svc := newServiceForTest(t, repo, &mockPlan{})

	err := svc.DeleteOperator(context.Background(), uuid.New(), opID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
