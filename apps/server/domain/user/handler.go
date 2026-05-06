// domain/user/handler.go
//
// Handler user — profile owner, info bisnis, dan manajemen operator.
//
// Endpoint:
//   GET    /users/me                          — profil owner + bisnis
//   PATCH  /users/me                          — update profil owner
//   GET    /business                          — info bisnis milik owner
//   PATCH  /business                          — update info bisnis
//   GET    /business/operators                — list operator + slot info dari plan
//   POST   /business/operators                — buat operator baru (cek slot dari plan)
//   PATCH  /business/operators/:operator_id   — update credentials operator
//   DELETE /business/operators/:operator_id   — hapus operator
//   GET    /business/operators/:operator_id/qr — generate QR login plaintext

package user

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
	"golang.org/x/crypto/bcrypt"
)

// ----------------------------------------------------------------
// Helper
// ----------------------------------------------------------------

func businessIDFromCtx(c echo.Context) string {
	id, _ := c.Get("business_id").(string)
	return id
}

func userIDFromCtx(c echo.Context) string {
	id, _ := c.Get("user_id").(string)
	return id
}

// maxOperatorsFromPlan mengambil batas max_operators dari plan aktif user.
// Fallback ke 3 kalau tidak ada subscription aktif.
func maxOperatorsFromPlan(db *sql.DB, userID string) (int, error) {
	var maxOps int
	err := db.QueryRow(
		`SELECT p.max_operators
		 FROM subscriptions s
		 JOIN plans p ON p.id = s.plan_id
		 WHERE s.user_id = $1
		   AND s.status = 'active'
		   AND (s.expires_at IS NULL OR s.expires_at > NOW())
		 ORDER BY s.started_at DESC
		 LIMIT 1`,
		userID,
	).Scan(&maxOps)

	if errors.Is(err, sql.ErrNoRows) {
		return 3, nil // default free plan
	}
	if err != nil {
		return 0, err
	}
	return maxOps, nil
}

// ----------------------------------------------------------------
// GET /users/me
// ----------------------------------------------------------------

type businessInMe struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Location          *string `json:"location"`
	Category          *string `json:"category"`
	MidtransConnected bool    `json:"midtrans_connected"`
}

type meResponse struct {
	ID       string       `json:"id"`
	Email    string       `json:"email"`
	FullName string       `json:"full_name"`
	Phone    *string      `json:"phone"`
	Role     string       `json:"role"`
	Business businessInMe `json:"business"`
}

func getMe(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := userIDFromCtx(c)
		role, _ := c.Get("role").(string)

		var res meResponse
		res.Role = role

		err := db.QueryRow(
			`SELECT u.id, u.email, u.full_name, u.phone,
			        b.id, b.name, b.location, b.category, b.midtrans_connected
			 FROM users u
			 LEFT JOIN businesses b ON b.user_id = u.id
			 WHERE u.id = $1`,
			userID,
		).Scan(
			&res.ID, &res.Email, &res.FullName, &res.Phone,
			&res.Business.ID, &res.Business.Name, &res.Business.Location,
			&res.Business.Category, &res.Business.MidtransConnected,
		)

		if errors.Is(err, sql.ErrNoRows) {
			return response.NotFound(c)
		}
		if err != nil {
			return response.Internal(c)
		}

		return response.OK(c, res)
	}
}

// ----------------------------------------------------------------
// PATCH /users/me
// ----------------------------------------------------------------

func updateMe(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := userIDFromCtx(c)

		var req struct {
			FullName string  `json:"full_name" validate:"omitempty,min=1,max=255"`
			Phone    *string `json:"phone"     validate:"omitempty,min=1,max=20"`
		}
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "invalid request body")
		}
		if err := c.Validate(&req); err != nil {
			return response.BadRequest(c, err.Error())
		}

		_, err := db.Exec(
			`UPDATE users
			 SET full_name  = COALESCE(NULLIF($1, ''), full_name),
			     phone      = COALESCE($2, phone),
			     updated_at = NOW()
			 WHERE id = $3`,
			req.FullName, req.Phone, userID,
		)
		if err != nil {
			return response.Internal(c)
		}

		return response.NoContent(c)
	}
}

// ----------------------------------------------------------------
// GET /business
// ----------------------------------------------------------------

type businessResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Slug              string  `json:"slug"`
	Location          *string `json:"location"`
	Category          *string `json:"category"`
	Timezone          string  `json:"timezone"`
	Currency          string  `json:"currency"`
	MidtransConnected bool    `json:"midtrans_connected"`
}

func getBusiness(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		businessID := businessIDFromCtx(c)

		var res businessResponse
		err := db.QueryRow(
			`SELECT id, name, slug, location, category, timezone, currency, midtrans_connected
			 FROM businesses WHERE id = $1`,
			businessID,
		).Scan(
			&res.ID, &res.Name, &res.Slug, &res.Location, &res.Category,
			&res.Timezone, &res.Currency, &res.MidtransConnected,
		)

		if errors.Is(err, sql.ErrNoRows) {
			return response.NotFound(c)
		}
		if err != nil {
			return response.Internal(c)
		}

		return response.OK(c, res)
	}
}

// ----------------------------------------------------------------
// PATCH /business
// ----------------------------------------------------------------

func updateBusiness(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		businessID := businessIDFromCtx(c)

		var req struct {
			Name     string `json:"name"     validate:"omitempty,min=1,max=255"`
			Location string `json:"location" validate:"omitempty,min=1,max=255"`
			Category string `json:"category" validate:"omitempty,min=1,max=100"`
			Timezone string `json:"timezone" validate:"omitempty,min=1,max=100"`
			Currency string `json:"currency" validate:"omitempty,min=1,max=10"`
		}
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "invalid request body")
		}
		if err := c.Validate(&req); err != nil {
			return response.BadRequest(c, err.Error())
		}

		_, err := db.Exec(
			`UPDATE businesses
			 SET name     = COALESCE(NULLIF($1, ''), name),
			     location = COALESCE(NULLIF($2, ''), location),
			     category = COALESCE(NULLIF($3, ''), category),
			     timezone = COALESCE(NULLIF($4, ''), timezone),
			     currency = COALESCE(NULLIF($5, ''), currency),
			     updated_at = NOW()
			 WHERE id = $6`,
			req.Name, req.Location, req.Category, req.Timezone, req.Currency, businessID,
		)
		if err != nil {
			return response.Internal(c)
		}

		return response.NoContent(c)
	}
}

// ----------------------------------------------------------------
// GET /business/operators
// ----------------------------------------------------------------

type operatorItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

type operatorsResponse struct {
	Operators []operatorItem `json:"operators"`
	SlotUsed  int            `json:"slot_used"`
	SlotMax   int            `json:"slot_max"`
}

func listOperators(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		businessID := businessIDFromCtx(c)
		userID := userIDFromCtx(c)

		maxOps, err := maxOperatorsFromPlan(db, userID)
		if err != nil {
			return response.Internal(c)
		}

		rows, err := db.Query(
			`SELECT id, name, email, is_active FROM operators WHERE business_id = $1 ORDER BY created_at`,
			businessID,
		)
		if err != nil {
			return response.Internal(c)
		}
		defer rows.Close()

		operators := []operatorItem{}
		for rows.Next() {
			var op operatorItem
			if err := rows.Scan(&op.ID, &op.Name, &op.Email, &op.IsActive); err != nil {
				return response.Internal(c)
			}
			operators = append(operators, op)
		}

		return response.OK(c, operatorsResponse{
			Operators: operators,
			SlotUsed:  len(operators),
			SlotMax:   maxOps,
		})
	}
}

// ----------------------------------------------------------------
// POST /business/operators
// ----------------------------------------------------------------

func createOperator(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		businessID := businessIDFromCtx(c)
		userID := userIDFromCtx(c)

		// Ambil batas slot dari plan aktif.
		maxOps, err := maxOperatorsFromPlan(db, userID)
		if err != nil {
			return response.Internal(c)
		}

		// Cek slot terpakai.
		var count int
		if err := db.QueryRow(
			`SELECT COUNT(*) FROM operators WHERE business_id = $1`,
			businessID,
		).Scan(&count); err != nil {
			return response.Internal(c)
		}
		if maxOps != -1 && count >= maxOps {
			return response.Conflict(c, "slot operator penuh, upgrade plan untuk menambah lebih banyak kasir")
		}

		var req struct {
			Name     string `json:"name"     validate:"required,min=1,max=255"`
			Email    string `json:"email"    validate:"required,email"`
			Password string `json:"password" validate:"required,min=6"`
		}
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "invalid request body")
		}
		if err := c.Validate(&req); err != nil {
			return response.BadRequest(c, err.Error())
		}

		// Cek email duplikat.
		var exists bool
		if err := db.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM operators WHERE email = $1)`,
			req.Email,
		).Scan(&exists); err != nil {
			return response.Internal(c)
		}
		if exists {
			return response.Conflict(c, "email sudah digunakan")
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return response.Internal(c)
		}

		var operatorID string
		if err := db.QueryRow(
			`INSERT INTO operators (business_id, name, email, password_hash)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id`,
			businessID, req.Name, req.Email, string(hash),
		).Scan(&operatorID); err != nil {
			return response.Internal(c)
		}

		return response.Created(c, map[string]string{"id": operatorID})
	}
}

// ----------------------------------------------------------------
// PATCH /business/operators/:operator_id
// ----------------------------------------------------------------

func updateOperator(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		businessID := businessIDFromCtx(c)
		operatorID := c.Param("operator_id")

		var req struct {
			Name     string `json:"name"     validate:"omitempty,min=1,max=255"`
			Email    string `json:"email"    validate:"omitempty,email"`
			Password string `json:"password" validate:"omitempty,min=6"`
		}
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "invalid request body")
		}
		if err := c.Validate(&req); err != nil {
			return response.BadRequest(c, err.Error())
		}

		// Pastikan operator milik bisnis ini.
		var exists bool
		if err := db.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM operators WHERE id = $1 AND business_id = $2)`,
			operatorID, businessID,
		).Scan(&exists); err != nil {
			return response.Internal(c)
		}
		if !exists {
			return response.NotFound(c)
		}

		if req.Name != "" || req.Email != "" {
			if _, err := db.Exec(
				`UPDATE operators
				 SET name  = COALESCE(NULLIF($1, ''), name),
				     email = COALESCE(NULLIF($2, ''), email)
				 WHERE id = $3 AND business_id = $4`,
				req.Name, req.Email, operatorID, businessID,
			); err != nil {
				return response.Internal(c)
			}
		}

		if req.Password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				return response.Internal(c)
			}
			if _, err := db.Exec(
				`UPDATE operators SET password_hash = $1 WHERE id = $2 AND business_id = $3`,
				string(hash), operatorID, businessID,
			); err != nil {
				return response.Internal(c)
			}
		}

		return response.NoContent(c)
	}
}

// ----------------------------------------------------------------
// DELETE /business/operators/:operator_id
// ----------------------------------------------------------------

func deleteOperator(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		businessID := businessIDFromCtx(c)
		operatorID := c.Param("operator_id")

		res, err := db.Exec(
			`DELETE FROM operators WHERE id = $1 AND business_id = $2`,
			operatorID, businessID,
		)
		if err != nil {
			return response.Internal(c)
		}

		rows, _ := res.RowsAffected()
		if rows == 0 {
			return response.NotFound(c)
		}

		return response.NoContent(c)
	}
}

// ----------------------------------------------------------------
// GET /business/operators/:operator_id/qr
// ----------------------------------------------------------------

func getOperatorQR(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		businessID := businessIDFromCtx(c)
		operatorID := c.Param("operator_id")

		var email string
		err := db.QueryRow(
			`SELECT email FROM operators WHERE id = $1 AND business_id = $2`,
			operatorID, businessID,
		).Scan(&email)

		if errors.Is(err, sql.ErrNoRows) {
			return response.NotFound(c)
		}
		if err != nil {
			return response.Internal(c)
		}

		payload, _ := json.Marshal(map[string]string{
			"operator_id": operatorID,
			"email":       email,
		})
		qrData := base64.StdEncoding.EncodeToString(payload)

		return response.OK(c, map[string]string{"qr_data": qrData})
	}
}
