package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestClientIP_XForwardedFor(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", " 10.0.0.1 , 192.168.1.1 ")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if got := ClientIP(c); got != "10.0.0.1" {
		t.Fatalf("ClientIP=%q", got)
	}
}

func TestClientIP_RealIP(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.5:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if got := ClientIP(c); got == "" {
		t.Fatal("expected non-empty RealIP fallback")
	}
}
