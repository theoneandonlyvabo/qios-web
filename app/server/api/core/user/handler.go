// domain/user/handler.go
//
// Handler user — profile owner dan info bisnis.
//
// Endpoint:
//   GET    /users/me  — profil owner + bisnis
//   PATCH  /users/me  — update profil owner
//   GET    /business  — info bisnis milik owner
//   PATCH  /business  — update info bisnis (selain xendit_*)
//
// Manajemen operator dipindah ke domain/operator.

package user

import (
	"database/sql"
	"errors"

	"github.com/labstack/echo/v4"

	"github.com/theoneandonlyvabo/qios-web/app/server/api/pkg/response"
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
	ID           string  `json:"id"`
	QiosID       string  `json:"qios_id"`
	BusinessName string  `json:"business_name"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	City         *string `json:"city"`
	Country      *string `json:"country"`
	XenditStatus string  `json:"xendit_status"`
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
			        b.id, b.qios_id, b.business_name, b.phone, b.address, b.city, b.country, b.xendit_status
			 FROM users u
			 LEFT JOIN businesses b ON b.user_id = u.id
			 WHERE u.id = $1`,
			userID,
		).Scan(
			&res.ID, &res.Email, &res.FullName, &res.Phone,
			&res.Business.ID, &res.Business.QiosID, &res.Business.BusinessName,
			&res.Business.Phone, &res.Business.Address, &res.Business.City,
			&res.Business.Country, &res.Business.XenditStatus,
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
			Phone    *string `json:"phone"     validate:"omitempty,min=1,max=32"`
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
	ID           string  `json:"id"`
	QiosID       string  `json:"qios_id"`
	BusinessName string  `json:"business_name"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	City         *string `json:"city"`
	Country      *string `json:"country"`
	XenditStatus string  `json:"xendit_status"`
}

func getBusiness(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		businessID := businessIDFromCtx(c)

		var res businessResponse
		err := db.QueryRow(
			`SELECT id, qios_id, business_name, phone, address, city, country, xendit_status
			 FROM businesses WHERE id = $1`,
			businessID,
		).Scan(
			&res.ID, &res.QiosID, &res.BusinessName, &res.Phone,
			&res.Address, &res.City, &res.Country, &res.XenditStatus,
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
			BusinessName string `json:"business_name" validate:"omitempty,min=1,max=255"`
			Phone        string `json:"phone"         validate:"omitempty,min=1,max=32"`
			Address      string `json:"address"       validate:"omitempty,min=1,max=1024"`
			City         string `json:"city"          validate:"omitempty,min=1,max=100"`
			Country      string `json:"country"       validate:"omitempty,min=2,max=100"`
		}
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "invalid request body")
		}
		if err := c.Validate(&req); err != nil {
			return response.BadRequest(c, err.Error())
		}

		_, err := db.Exec(
			`UPDATE businesses
			 SET business_name = COALESCE(NULLIF($1, ''), business_name),
			     phone         = COALESCE(NULLIF($2, ''), phone),
			     address       = COALESCE(NULLIF($3, ''), address),
			     city          = COALESCE(NULLIF($4, ''), city),
			     country       = COALESCE(NULLIF($5, ''), country),
			     updated_at    = NOW()
			 WHERE id = $6`,
			req.BusinessName, req.Phone, req.Address, req.City, req.Country, businessID,
		)
		if err != nil {
			return response.Internal(c)
		}

		return response.NoContent(c)
	}
}
