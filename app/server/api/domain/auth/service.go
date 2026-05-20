// domain/auth/service.go
//
// Business logic untuk auth: Login, Register, Refresh, Logout.
// Service tidak menyentuh database langsung — semua via Repository.
// Service tidak menyentuh HTTP — handler yang menerjemahkan ke response.
//
// Refresh token disimpan di DB sebagai SHA-256 hash dari plaintext.
// Plaintext-nya dipasang ke httpOnly cookie oleh handler.

package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/payment"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/qiosid"
)

const (
	bcryptCost              = 12
	xenditCreateAccountTime = 12 * time.Second
)

// Service mendefinisikan kontrak auth flow.
type Service interface {
	Login(ctx context.Context, email, password string) (*LoginResult, error)
	Register(ctx context.Context, in RegisterInput) (*RegisterResult, error)
	Refresh(ctx context.Context, refreshTokenPlain string) (*RefreshResult, error)
	Logout(ctx context.Context, refreshTokenPlain string) error
}

type service struct {
	db        *sql.DB
	repo      Repository
	jwtSvc    *jwt.Service
	xenditSvc xenditCreator
}

// NewService merangkai dependency auth service.
// db diperlukan untuk membuka transaksi multi-table di Register.
func NewService(db *sql.DB, repo Repository, jwtSvc *jwt.Service, xenditSvc xenditCreator) Service {
	return &service{
		db:        db,
		repo:      repo,
		jwtSvc:    jwtSvc,
		xenditSvc: xenditSvc,
	}
}

// ----------------------------------------------------------------
// Login
// ----------------------------------------------------------------

func (s *service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if !user.IsActive || user.IsSuspended {
		return nil, ErrAccountInactive
	}
	if user.PasswordHash == "" {
		return nil, ErrGoogleOnlyAccount
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.jwtSvc.IssueAccessToken(user.ID, user.BusinessID, jwt.RoleOwner)
	if err != nil {
		return nil, fmt.Errorf("auth service: issue access token: %w", err)
	}

	plain, hashed, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth service: generate refresh: %w", err)
	}
	if err := s.repo.StoreRefreshToken(ctx, user.ID, hashed, s.jwtSvc.RefreshExpiry()); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:   accessToken,
		RefreshToken:  plain,
		RefreshExpiry: s.jwtSvc.RefreshExpiry(),
	}, nil
}

// ----------------------------------------------------------------
// Register
// ----------------------------------------------------------------

// Register mendaftarkan owner baru dalam satu transaksi:
//  1. INSERT user
//  2. Generate QIOS ID (dalam tx yang sama, FOR UPDATE)
//  3. INSERT business dengan xendit_status = PENDING
//  4. Call Xendit CreateSubAccount (network — di luar DB tapi sebelum commit)
//  5. UPDATE business set xendit_account_id + status REGISTERED
//  6. Commit
//
// Kalau Xendit gagal, seluruh tx di-rollback (user + business tidak terbuat).
// Setelah commit, terbitkan access + refresh token (kegagalan di sini tidak
// rollback karena akun sudah valid — user tinggal login ulang).
func (s *service) Register(ctx context.Context, in RegisterInput) (*RegisterResult, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("auth service: hash password: %w", err)
	}

	txCtx, cancel := context.WithTimeout(ctx, xenditCreateAccountTime+5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(txCtx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("auth service: begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// Step 1 — insert user.
	userID, err := s.repo.CreateUser(txCtx, tx, in.Email, string(passwordHash), in.FullName, in.Phone)
	if err != nil {
		return nil, err
	}

	// Step 2 — generate QIOS ID di dalam tx yang sama.
	qiosIDValue, err := qiosid.Generate(tx)
	if err != nil {
		return nil, fmt.Errorf("auth service: generate qios id: %w", err)
	}

	// Step 3 — insert business dengan status PENDING.
	businessID, err := s.repo.CreateBusiness(txCtx, tx,
		qiosIDValue, userID, in.BusinessName, in.Phone, in.Address, in.City, in.Country,
	)
	if err != nil {
		return nil, err
	}

	// Step 4 — call Xendit MANAGED account API.
	xenditAccountID, xenditStatus, err := s.createXenditAccount(txCtx, in, businessID)
	if err != nil {
		return nil, err
	}

	// Step 5 — persist Xendit info.
	if err := s.repo.UpdateBusinessXendit(txCtx, tx, businessID, xenditAccountID, xenditStatus); err != nil {
		return nil, err
	}

	// Step 6 — commit.
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("auth service: commit register tx: %w", err)
	}
	committed = true

	// Issue tokens (kegagalan di sini tidak rollback akun).
	accessToken, err := s.jwtSvc.IssueAccessToken(userID, businessID, jwt.RoleOwner)
	if err != nil {
		return nil, fmt.Errorf("auth service: issue access token: %w", err)
	}
	plain, hashed, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth service: generate refresh: %w", err)
	}
	if err := s.repo.StoreRefreshToken(ctx, userID, hashed, s.jwtSvc.RefreshExpiry()); err != nil {
		return nil, err
	}

	return &RegisterResult{
		UserID:        userID,
		BusinessID:    businessID,
		QiosID:        qiosIDValue,
		XenditStatus:  xenditStatus,
		AccessToken:   accessToken,
		RefreshToken:  plain,
		RefreshExpiry: s.jwtSvc.RefreshExpiry(),
	}, nil
}

// createXenditAccount memanggil Xendit API atau pakai mock kalau SKIP_XENDIT=true.
// Dipisah supaya Register tetap pendek dan mudah dibaca.
func (s *service) createXenditAccount(ctx context.Context, in RegisterInput, businessID string) (accountID, status string, err error) {
	if os.Getenv("SKIP_XENDIT") == "true" {
		return "dev-mock-" + businessID, string(payment.StatusRegistered), nil
	}

	callCtx, cancelCall := context.WithTimeout(ctx, xenditCreateAccountTime)
	defer cancelCall()

	res, err := s.xenditSvc.CreateSubAccount(callCtx, payment.ManagedAccountInput{
		Email:        in.Email,
		BusinessName: in.BusinessName,
		Country:      countryToISO(in.Country),
	})
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrXenditCreate, err)
	}
	return res.AccountID, string(res.Status), nil
}

// ----------------------------------------------------------------
// Refresh
// ----------------------------------------------------------------

// Refresh memutar refresh token: validate → delete lama → issue baru.
// Kalau token expired, baris di DB juga dibersihkan supaya tidak menumpuk.
func (s *service) Refresh(ctx context.Context, refreshTokenPlain string) (*RefreshResult, error) {
	tokenHash := hashToken(refreshTokenPlain)

	userID, expiresAt, err := s.repo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if time.Now().After(expiresAt) {
		_ = s.repo.DeleteRefreshToken(ctx, tokenHash)
		return nil, ErrSessionExpired
	}

	businessID, role, err := s.repo.FindUserRoleByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteRefreshToken(ctx, tokenHash); err != nil {
		return nil, err
	}

	var newAccessToken string
	if role == roleOperator {
		newAccessToken, err = s.jwtSvc.IssueOperatorAccessToken(userID, businessID)
	} else {
		newAccessToken, err = s.jwtSvc.IssueAccessToken(userID, businessID, role)
	}
	if err != nil {
		return nil, fmt.Errorf("auth service: issue access token: %w", err)
	}

	newPlain, newHashed, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth service: generate refresh: %w", err)
	}
	if err := s.repo.StoreRefreshToken(ctx, userID, newHashed, s.jwtSvc.RefreshExpiry()); err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:   newAccessToken,
		RefreshToken:  newPlain,
		RefreshExpiry: s.jwtSvc.RefreshExpiry(),
	}, nil
}

// ----------------------------------------------------------------
// Logout
// ----------------------------------------------------------------

// Logout menghapus refresh token dari DB. Aman dipanggil tanpa token —
// handler bisa selalu memanggil ini dan mengandalkan idempotency.
func (s *service) Logout(ctx context.Context, refreshTokenPlain string) error {
	if refreshTokenPlain == "" {
		return nil
	}
	return s.repo.DeleteRefreshToken(ctx, hashToken(refreshTokenPlain))
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

// generateRefreshToken menghasilkan token plaintext dan hash SHA-256-nya.
func generateRefreshToken() (plain, hashed string, err error) {
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

// countryToISO mapping minimal nama negara → ISO-3166 alpha-2 yang dipakai Xendit.
// Kalau input sudah 2 huruf, biarkan apa adanya.
func countryToISO(country string) string {
	c := strings.TrimSpace(country)
	if len(c) == 2 {
		return strings.ToUpper(c)
	}
	switch strings.ToLower(c) {
	case "indonesia":
		return "ID"
	case "philippines":
		return "PH"
	case "malaysia":
		return "MY"
	case "thailand":
		return "TH"
	case "vietnam":
		return "VN"
	default:
		return "" // biarkan kosong; Xendit akan default berdasarkan master account
	}
}

// ensure *payment.XenditService memenuhi xenditCreator (compile-time check).
var _ xenditCreator = (*payment.XenditService)(nil)

// Compile-time check juga untuk service.
var _ Service = (*service)(nil)
