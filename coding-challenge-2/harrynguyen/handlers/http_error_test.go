package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/dotdak/sms-otp/internal/phone"
	"github.com/dotdak/sms-otp/internal/ratelimit"
	"github.com/labstack/echo/v4"
)

func TestAllowTestSMSHeaders(t *testing.T) {
	t.Setenv("ALLOW_TEST_SMS_HEADERS", "true")
	if !allowTestSMSHeaders() {
		t.Fatal("expected true")
	}
	t.Setenv("ALLOW_TEST_SMS_HEADERS", "false")
	if allowTestSMSHeaders() {
		t.Fatal("expected false")
	}
}

func TestTestSMSModeFromRequest(t *testing.T) {
	t.Setenv("ALLOW_TEST_SMS_HEADERS", "true")
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("x-test-sms-mode", "always-fail")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if got := testSMSModeFromRequest(c); got != "always-fail" {
		t.Fatalf("got %q", got)
	}
	t.Setenv("ALLOW_TEST_SMS_HEADERS", "")
	if got := testSMSModeFromRequest(c); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestRespondTooManyRequests_WithLimitedError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := respondTooManyRequests(c, &ratelimit.LimitedError{RetryAfter: 2 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("code=%d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After")
	}
}

func TestRespondTooManyRequests_OtherError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := respondTooManyRequests(c, errors.New("other"))
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestRespondSendSMSError_RateLimit(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := respondSendSMSError(c, &ratelimit.LimitedError{RetryAfter: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestRespondSendSMSError_PhoneError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := respondSendSMSError(c, phone.ErrInvalidNumber)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestRespondSendSMSError_Internal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := respondSendSMSError(c, errors.New("boom"))
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestMaskPhoneSuffix(t *testing.T) {
	if maskPhoneSuffix("12") != "****" {
		t.Fatal(maskPhoneSuffix("12"))
	}
	if !strings.HasSuffix(maskPhoneSuffix("841234567890"), "7890") {
		t.Fatal(maskPhoneSuffix("841234567890"))
	}
}

func TestAuthFlowAttrs(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := obslog.WithRequestID(req.Context(), "rid")
	ctx = obslog.WithCorrelationID(ctx, "cid")
	a := authFlowAttrs(ctx, "register", "parse_body")
	if len(a) < 4 {
		t.Fatal(a)
	}
}

func TestSmsCallbackAttrs(t *testing.T) {
	ctx := obslog.WithCorrelationID(context.Background(), "c")
	a := smsCallbackAttrs(ctx, "bind")
	if len(a) < 4 {
		t.Fatal(a)
	}
}
