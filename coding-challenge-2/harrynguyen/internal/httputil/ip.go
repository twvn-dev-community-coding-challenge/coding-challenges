package httputil

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// ClientIP returns the first X-Forwarded-For hop when present, otherwise Echo RealIP.
func ClientIP(c echo.Context) string {
	xff := c.Request().Header.Get("X-Forwarded-For")
	if xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	return c.RealIP()
}
