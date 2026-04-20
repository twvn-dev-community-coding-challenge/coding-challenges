package httpmw

import (
	"time"

	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/labstack/echo/v4"
)

// AccessLog writes one structured line per request after handling (use after CorrelationID and RequestID).
func AccessLog() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			obslog.Init()
			start := time.Now()
			err := next(c)
			req := c.Request()
			status := c.Response().Status
			if status == 0 {
				status = 200
			}
			attrs := []any{
				"request_id", obslog.RequestID(req.Context()),
				"method", req.Method,
				"path", req.URL.Path,
				"status", status,
				"latency_ms", time.Since(start).Milliseconds(),
			}
			if cid := obslog.CorrelationID(req.Context()); cid != "" {
				attrs = append([]any{"correlation_id", cid}, attrs...)
			}
			obslog.L.InfoContext(req.Context(), obslog.MsgHTTPRequest, attrs...)
			return err
		}
	}
}
