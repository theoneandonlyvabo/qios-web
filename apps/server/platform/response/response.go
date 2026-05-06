// platform/response/response.go
//
// Helper untuk mengembalikan response JSON yang konsisten di seluruh aplikasi.
// Semua handler wajib menggunakan fungsi di sini, bukan menulis JSON manual.
//
// Shape standar:
//   { "success": true/false, "data": ..., "error": "..." }
//
// Konvensi penamaan:
//   - Fungsi tanpa suffix  → response generik, pesan error sudah ditentukan.
//   - Fungsi dengan suffix "Msg" → response dengan custom error message dari handler.

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

// ----------------------------------------------------------------
// 2xx
// ----------------------------------------------------------------

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

// ----------------------------------------------------------------
// 4xx — generik (pesan default)
// ----------------------------------------------------------------

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

// ----------------------------------------------------------------
// 4xx — Msg variant (custom message dari handler)
// ----------------------------------------------------------------

// UnauthorizedMsg mengembalikan 401 dengan pesan custom.
func UnauthorizedMsg(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, body{Success: false, Error: message})
}

// ForbiddenMsg mengembalikan 403 dengan pesan custom.
func ForbiddenMsg(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, body{Success: false, Error: message})
}

// NotFoundMsg mengembalikan 404 dengan pesan custom.
func NotFoundMsg(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, body{Success: false, Error: message})
}

// NotImplemented mengembalikan 501 dengan pesan custom.
func NotImplemented(c echo.Context, message string) error {
	return c.JSON(http.StatusNotImplemented, body{Success: false, Error: message})
}

// ----------------------------------------------------------------
// 5xx
// ----------------------------------------------------------------

// Internal mengembalikan 500. Pesan error internal tidak diekspos ke client.
func Internal(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, body{Success: false, Error: "Terjadi kesalahan pada server"})
}

// InternalError mengembalikan 500 dengan pesan custom.
// Gunakan hanya untuk debugging atau logging — jangan ekspos detail teknis ke production client.
func InternalError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, body{Success: false, Error: message})
}
