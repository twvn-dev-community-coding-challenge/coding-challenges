package httpmw

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// InternalAPIKey requires Authorization: Bearer <key> or X-Internal-Key: <key>.
func InternalAPIKey(expected string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if expected == "" {
				return next(c)
			}
			token := bearerOrKey(c)
			if token != expected {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			}
			return next(c)
		}
	}
}

// CallbackSecret requires header X-Callback-Secret matching expected (for provider webhooks).
func CallbackSecret(expected string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if expected == "" {
				return next(c)
			}
			if c.Request().Header.Get("X-Callback-Secret") != expected {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid callback secret"})
			}
			return next(c)
		}
	}
}

func bearerOrKey(c echo.Context) string {
	auth := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return c.Request().Header.Get("X-Internal-Key")
}
