package handlers

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dotdak/sms-otp/internal/phone"
	"github.com/dotdak/sms-otp/internal/ratelimit"
	"github.com/labstack/echo/v4"
)

func allowTestSMSHeaders() bool {
	return strings.EqualFold(os.Getenv("ALLOW_TEST_SMS_HEADERS"), "true")
}

func testSMSModeFromRequest(c echo.Context) string {
	if !allowTestSMSHeaders() {
		return ""
	}
	return c.Request().Header.Get("x-test-sms-mode")
}

func respondTooManyRequests(c echo.Context, err error) error {
	var lim *ratelimit.LimitedError
	if errors.As(err, &lim) {
		sec := int(lim.RetryAfter.Seconds())
		if sec < 1 {
			sec = 1
		}
		c.Response().Header().Set("Retry-After", strconv.Itoa(sec))
	}
	return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "Too many requests"})
}

func respondSendSMSError(c echo.Context, err error) error {
	var lim *ratelimit.LimitedError
	if errors.As(err, &lim) {
		return respondTooManyRequests(c, err)
	}
	if errors.Is(err, phone.ErrInvalidCountry) || errors.Is(err, phone.ErrInvalidNumber) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
