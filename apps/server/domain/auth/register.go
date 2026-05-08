// domain/auth/register.go
//
// Endpoint POST /auth/register.
//
// Flow atomic:
//   1. Begin tx
//   2. INSERT user (email + password_hash + full_name + phone)
//   3. Generate QM ID via platform/qmid (di dalam tx yang sama)
//   4. INSERT business dengan xendit_status = PENDING
//   5. Call Xendit CreateManagedAccount (network — di luar DB tapi sebelum commit)
//   6. UPDATE business set xendit_account_id + credentials + status REGISTERED
//   7. Commit
//
// Kalau step 5 atau 6 gagal, rollback seluruh transaksi.
// Setelah commit sukses, issue access + refresh token (sama seperti login).
//
// Catatan: external call (Xendit) dilakukan saat tx terbuka. Konsekuensi:
//   - Latency tx = network latency Xendit. Acceptable untuk register flow
//     yang frekuensinya rendah; jangan dipakai untuk endpoint hot-path.
//   - Kalau commit DB gagal setelah Xendit sukses, Xendit memiliki orphaned
//     sub-account. Cleanup orphan via reconciliation job — bukan inline.

package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/theoneandonlyvabo/qios-web/apps/server/config"
	"github.com/theoneandonlyvabo/qios-web/apps/server/domain/payment"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/jwt"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/qmid"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
)

const (
	bcryptCost              = 12
	xenditCreateAccountTime = 12 * time.Second
)

// xenditCreator adalah interface kecil supaya test bisa mock tanpa perlu
// inject http.Client. Production-nya dipenuhi oleh *payment.XenditService.
type xenditCreator interface {
	CreateManagedAccount(ctx context.Context, in payment.ManagedAccountInput) (*payment.ManagedAccountResult, error)
}

type registerRequest struct {
	Email        string `json:"email"         validate:"required,email,max=255"`
	Password     string `json:"password"      validate:"required,min=8,max=72"`
	FullName     string `json:"full_name"     validate:"required,min=1,max=255"`
	Phone        string `json:"phone"         validate:"required,min=4,max=32"`
	BusinessName string `json:"business_name" validate:"required,min=1,max=255"`
	Address      string `json:"address"       validate:"required,min=1,max=1024"`
	City         string `json:"city"          validate:"required,min=1,max=100"`
	Country      string `json:"country"       validate:"required,min=2,max=100"`
}

type registerResponse struct {
	AccessToken  string `json:"access_token"`
	UserID       string `json:"user_id"`
	BusinessID   string `json:"business_id"`
	QMID         string `json:"qm_id"`
	XenditStatus string `json:"xendit_status"`
}

// register membuat akun owner + bisnis + sub-account Xendit dalam satu tx.
// POST /auth/register
func register(
	db *sql.DB,
	cfg *config.Config,
	jwtSvc *jwt.Service,
	xenditSvc xenditCreator,
) echo.HandlerFunc {
	_ = cfg // reserved untuk masa depan (mis. country default dari config)

	return func(c echo.Context) error {
		var req registerRequest
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "invalid request body")
		}
		if err := c.Validate(&req); err != nil {
			return response.BadRequest(c, err.Error())
		}
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
		if err != nil {
			return response.Internal(c)
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), xenditCreateAccountTime+5*time.Second)
		defer cancel()

		tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return response.Internal(c)
		}
		// committed jadi true setelah tx.Commit() sukses; kalau false saat defer,
		// rollback dipanggil. Aman untuk dipanggil dua kali via sql package.
		committed := false
		defer func() {
			if !committed {
				_ = tx.Rollback()
			}
		}()

		// Step 2 — insert user.
		var userID string
		err = tx.QueryRowContext(ctx,
			`INSERT INTO users (email, password_hash, full_name, phone)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id`,
			req.Email, string(passwordHash), req.FullName, nullableString(req.Phone),
		).Scan(&userID)
		if err != nil {
			if isUniqueViolation(err) {
				return response.Conflict(c, "email sudah terdaftar")
			}
			return response.Internal(c)
		}

		// Step 3 — generate QM ID di dalam tx yang sama.
		qmIDValue, err := qmid.Generate(tx)
		if err != nil {
			return response.Internal(c)
		}

		// Step 4 — insert business dengan status PENDING.
		var businessID string
		err = tx.QueryRowContext(ctx,
			`INSERT INTO businesses (
				qm_id, user_id, business_name, phone, address, city, country, xendit_status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, 'PENDING')
			RETURNING id`,
			qmIDValue, userID, req.BusinessName,
			nullableString(req.Phone), nullableString(req.Address),
			nullableString(req.City), nullableString(req.Country),
		).Scan(&businessID)
		if err != nil {
			if isUniqueViolation(err) {
				return response.Conflict(c, "qm_id collision, silakan coba lagi")
			}
			return response.Internal(c)
		}

		// Step 5 — call Xendit MANAGED account API.
		xenditCtx, xenditCancel := context.WithTimeout(ctx, xenditCreateAccountTime)
		defer xenditCancel()

		xenditRes, err := xenditSvc.CreateManagedAccount(xenditCtx, payment.ManagedAccountInput{
			Email:        req.Email,
			BusinessName: req.BusinessName,
			Country:      countryToISO(req.Country),
		})
		if err != nil {
			c.Logger().Errorf("xendit create account failed: %v", err)
			return response.UnprocessableEntity(c, "gagal membuat akun pembayaran, coba lagi")
		}

		// Step 6 — update business dengan credentials + status REGISTERED.
		_, err = tx.ExecContext(ctx,
			`UPDATE businesses
			 SET xendit_account_id = $1,
			     xendit_api_key    = $2,
			     xendit_secret_key = $3,
			     xendit_status     = $4,
			     updated_at        = NOW()
			 WHERE id = $5`,
			xenditRes.AccountID,
			nullableString(xenditRes.APIKey),
			nullableString(xenditRes.SecretKey),
			string(xenditRes.Status),
			businessID,
		)
		if err != nil {
			return response.Internal(c)
		}

		// Step 7 — commit.
		if err := tx.Commit(); err != nil {
			return response.Internal(c)
		}
		committed = true

		// Issue access + refresh token. Kegagalan di sini tidak rollback
		// register karena akun sudah valid — user tinggal login ulang.
		accessToken, err := jwtSvc.IssueAccessToken(userID, businessID, jwt.RoleOwner)
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

		return c.JSON(http.StatusCreated, map[string]any{
			"success": true,
			"data": registerResponse{
				AccessToken:  accessToken,
				UserID:       userID,
				BusinessID:   businessID,
				QMID:         qmIDValue,
				XenditStatus: string(xenditRes.Status),
			},
		})
	}
}

// nullableString mengembalikan sql.NullString supaya kolom nullable di DB
// menyimpan NULL kalau input string kosong, bukan empty string.
func nullableString(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// isUniqueViolation cek apakah error berasal dari unique constraint Postgres.
func isUniqueViolation(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
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
