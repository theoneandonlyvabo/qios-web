// platform/response/response.go
//
// Helper untuk mengembalikan response JSON yang konsisten di seluruh aplikasi.
// Semua handler wajib menggunakan fungsi di sini, bukan menulis JSON manual.
//
// Shape standar:
//   { "success": true/false, "data": ..., "error": "..." }

package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type body struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Error   string `json:"error,omitempty"`
}

// OK mengembalikan 200 dengan data.
func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, body{Success: true, Data: data})
}

// Created mengembalikan 201 dengan data.
func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, body{Success: true, Data: data})
}

// NoContent mengembalikan 200 dengan data null.
func NoContent(c echo.Context) error {
	return c.JSON(http.StatusOK, body{Success: true})
}

// BadRequest mengembalikan 400 dengan pesan error.
func BadRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, body{Success: false, Error: message})
}

// Unauthorized mengembalikan 401.
func Unauthorized(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, body{Success: false, Error: "Unauthorized"})
}

// Forbidden mengembalikan 403.
func Forbidden(c echo.Context) error {
	return c.JSON(http.StatusForbidden, body{Success: false, Error: "Akses ditolak"})
}

// NotFound mengembalikan 404.
func NotFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, body{Success: false, Error: "Data tidak ditemukan"})
}

// Conflict mengembalikan 409 dengan pesan error.
func Conflict(c echo.Context, message string) error {
	return c.JSON(http.StatusConflict, body{Success: false, Error: message})
}

// UnprocessableEntity mengembalikan 422 dengan pesan error.
func UnprocessableEntity(c echo.Context, message string) error {
	return c.JSON(http.StatusUnprocessableEntity, body{Success: false, Error: message})
}

// Internal mengembalikan 500. Pesan error internal tidak diekspos ke client.
func Internal(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, body{Success: false, Error: "Terjadi kesalahan pada server"})
}
