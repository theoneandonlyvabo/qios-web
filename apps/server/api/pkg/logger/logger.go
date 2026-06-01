// platform/logger/logger.go
//
// Custom logger untuk QIOS server — output kompak, kolom rata, low-noise.
//
// Format HTTP:    HH:MM:SS  CODE  METHOD  /uri                    latency
// Format log:     HH:MM:SS  LEVEL message
// Format webhook: HH:MM:SS  HOOK  event external_id STATUS amount IDR
//
// Path yang di-skip dari middleware HTTP log:
//   /metrics, /health, /readyz, /favicon.ico

package logger

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	reset = "\033[0m"
	bold  = "\033[1m"
	dim   = "\033[2m"

	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	cyan   = "\033[36m"
	blue   = "\033[34m"
	gray   = "\033[90m"
	white  = "\033[97m"
)

var skipPaths = map[string]struct{}{
	"/metrics":      {},
	"/health":       {},
	"/readyz":       {},
	"/favicon.ico":  {},
}

func ts() string {
	return dim + time.Now().Format("15:04:05") + reset
}

func colorStatus(code int) string {
	switch {
	case code >= 500:
		return red + bold
	case code >= 400:
		return yellow + bold
	case code >= 300:
		return cyan + bold
	default:
		return green + bold
	}
}

func colorMethod(m string) string {
	switch m {
	case "GET":
		return blue + bold
	case "POST":
		return green + bold
	case "PUT", "PATCH":
		return yellow + bold
	case "DELETE":
		return red + bold
	default:
		return cyan + bold
	}
}

func webhookStatusColor(status string) string {
	switch status {
	case "COMPLETED", "SUCCEEDED", "PAID", "SUCCESS", "ACTIVE":
		return green + bold
	case "PENDING", "REGISTERED":
		return yellow + bold
	default:
		return red + bold
	}
}

// Middleware returns Echo middleware yang log tiap HTTP request.
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, skip := skipPaths[c.Request().URL.Path]; skip {
				return next(c)
			}

			start := time.Now()
			err := next(c)

			req := c.Request()
			res := c.Response()
			latencyMs := time.Since(start).Milliseconds()

			status := res.Status
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = 500
				}
			}

			fmt.Printf(
				"%s  %s%3d%s  %s%-6s%s %s%-40s%s %s%5dms%s\n",
				ts(),
				colorStatus(status), status, reset,
				colorMethod(req.Method), req.Method, reset,
				white, req.RequestURI, reset,
				gray, latencyMs, reset,
			)

			return err
		}
	}
}

func levelLine(color, level, format string, args ...any) {
	fmt.Printf("%s  %s%-5s%s %s\n", ts(), color, level, reset, fmt.Sprintf(format, args...))
}

func Info(format string, args ...any)  { levelLine(green+bold, "INFO", format, args...) }
func Warn(format string, args ...any)  { levelLine(yellow+bold, "WARN", format, args...) }
func Error(format string, args ...any) { levelLine(red+bold, "ERROR", format, args...) }

// Webhook logs structured webhook event.
// amount 0 = tidak relevan untuk event ini (mis. account.activated).
func Webhook(event, externalID, status string, amount int64) {
	amountStr := ""
	if amount > 0 {
		amountStr = fmt.Sprintf(" %s%d IDR%s", gray, amount, reset)
	}
	externalStr := ""
	if externalID != "" {
		externalStr = fmt.Sprintf(" %s%s%s", gray, externalID, reset)
	}
	fmt.Printf(
		"%s  %sHOOK%s  %s%s%s%s  %s%s%s%s\n",
		ts(),
		cyan+bold, reset,
		white, event, reset, externalStr,
		webhookStatusColor(status), status, reset,
		amountStr,
	)
}
