// domain/user/handler.go
//
// Handler user — profile owner dan info bisnis.
//
// Endpoint:
//   GET    /users/me  — profil owner + bisnis
//   PATCH  /users/me  — update profil owner
//   GET    /business  — info bisnis milik owner
//   PATCH  /business  — update info bisnis
//
// Manajemen operator dipindah ke domain/operator.

package user

import (
	"database/sql"
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/theoneandonlyvabo/qios-web/apps/server/platform/response"
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

