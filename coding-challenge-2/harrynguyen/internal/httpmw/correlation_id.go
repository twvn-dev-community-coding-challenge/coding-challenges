package httpmw

import (
	"strings"

	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/labstack/echo/v4"
)

// HeaderXCorrelationID is the canonical HTTP header for a client- or edge-provided trace id.
const HeaderXCorrelationID = "X-Correlation-Id"

// CorrelationID ensures X-Correlation-ID on the response and stores it on the request context
// for structured logs (distinct from X-Request-ID, which identifies a single HTTP request).
// Run before RequestID so both ids are available downstream.
func CorrelationID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cid := strings.TrimSpace(c.Request().Header.Get(HeaderXCorrelationID))
			if cid == "" {
				cid = randomID()
			}
			c.Response().Header().Set(HeaderXCorrelationID, cid)
			ctx := obslog.WithCorrelationID(c.Request().Context(), cid)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
