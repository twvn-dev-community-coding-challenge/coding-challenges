package metrics

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// EchoMiddleware records sms_otp_http_requests_total. Skips /metrics to avoid scrape noise.
func EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			handler := c.Path()
			if handler == "" {
				handler = c.Request().URL.Path
			}
			if handler == "/metrics" {
				return next(c)
			}
			err := next(c)
			code := strconv.Itoa(c.Response().Status)
			if handler == "" {
				handler = "unknown"
			}
			HTTPRequestsTotal.WithLabelValues(c.Request().Method, handler, code).Inc()
			return err
		}
	}
}
