package httpmw

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/labstack/echo/v4"
)

// RequestID ensures X-Request-ID on the response and attaches the id to the request context for slog.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rid := c.Request().Header.Get(echo.HeaderXRequestID)
			if rid == "" {
				rid = randomID()
			}
			c.Response().Header().Set(echo.HeaderXRequestID, rid)
			ctx := obslog.WithRequestID(c.Request().Context(), rid)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func randomID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "req-unknown"
	}
	return hex.EncodeToString(b[:])
}
