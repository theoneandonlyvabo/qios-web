// platform/logger/logger.go
//
// Custom logger untuk QIOS server.
//
// Exports:
//   - Middleware()         Echo middleware — log tiap HTTP request
//   - Info(msg, args...)   General info log
//   - Warn(msg, args...)   Warning log
//   - Error(msg, args...)  Error log
//   - Webhook(event, externalID, status, amount)  Structured webhook log
//
// Format HTTP:    [TIME] [STATUS] METHOD /uri latency
// Format log:     [TIME] [LEVEL] message
// Format webhook: [TIME] [WEBHOOK] event external_id STATUS amount IDR

package logger

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
	white  = "\033[97m"
)

func ts() string {
	return gray + time.Now().Format("2006-01-02 15:04:05") + reset
}

func httpStatusLabel(code int) string {
	switch {
	case code >= 500:
		return red + bold + "SERVER ERROR" + reset
	case code >= 400:
		return yellow + bold + "CLIENT ERROR" + reset
	case code >= 300:
		return cyan + bold + "REDIRECT" + reset
	default:
		return green + bold + "SUCCESS" + reset
	}
}

func webhookStatusColor(status string) string {
	switch status {
	case "COMPLETED", "SUCCEEDED", "PAID", "SUCCESS", "ACTIVE":
		return green + bold + status + reset
	case "PENDING", "REGISTERED":
		return yellow + bold + status + reset
	default:
		return red + bold + status + reset
	}
}

// Middleware returns Echo middleware yang log tiap HTTP request.
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			req := c.Request()
			res := c.Response()
			latency := time.Since(start)

			status := res.Status
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = 500
				}
			}

			fmt.Printf(
				"[%s] [%s] %s%s%s %s%s%s %s%v%s\n",
				ts(),
				httpStatusLabel(status),
				cyan, req.Method, reset,
				white, req.RequestURI, reset,
				gray, latency.Round(time.Millisecond), reset,
			)

			return err
		}
	}
}

func Info(format string, args ...any) {
	fmt.Printf("[%s] [%s] %s\n", ts(), green+bold+"INFO"+reset, fmt.Sprintf(format, args...))
}

func Warn(format string, args ...any) {
	fmt.Printf("[%s] [%s] %s\n", ts(), yellow+bold+"WARN"+reset, fmt.Sprintf(format, args...))
}

func Error(format string, args ...any) {
	fmt.Printf("[%s] [%s] %s\n", ts(), red+bold+"ERROR"+reset, fmt.Sprintf(format, args...))
}

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
		"[%s] [%s] %s%s %s%s\n",
		ts(),
		cyan+bold+"WEBHOOK"+reset,
		white+event+reset,
		externalStr,
		webhookStatusColor(status),
		amountStr,
	)
}
