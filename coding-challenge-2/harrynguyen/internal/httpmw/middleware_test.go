package httpmw

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/labstack/echo/v4"
)

func TestCorrelationID_Generated(t *testing.T) {
	e := echo.New()
	e.Use(CorrelationID())
	e.GET("/", func(c echo.Context) error {
		if obslog.CorrelationID(c.Request().Context()) == "" {
			return echo.NewHTTPError(500, "missing cid")
		}
		return c.NoContent(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d", rec.Code)
	}
	if rec.Header().Get(HeaderXCorrelationID) == "" {
		t.Fatal("expected response correlation id")
	}
}

func TestCorrelationID_FromHeader(t *testing.T) {
	e := echo.New()
	e.Use(CorrelationID())
	e.GET("/", func(c echo.Context) error {
		if obslog.CorrelationID(c.Request().Context()) != "trace-1" {
			return echo.NewHTTPError(500)
		}
		return c.NoContent(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderXCorrelationID, "trace-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Header().Get(HeaderXCorrelationID) != "trace-1" {
		t.Fatal("expected echo of correlation id")
	}
}

func TestRequestID_FromHeader(t *testing.T) {
	e := echo.New()
	e.Use(RequestID())
	e.GET("/", func(c echo.Context) error {
		if obslog.RequestID(c.Request().Context()) != "rid-9" {
			return echo.NewHTTPError(500)
		}
		return c.NoContent(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderXRequestID, "rid-9")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Header().Get(echo.HeaderXRequestID) != "rid-9" {
		t.Fatal("expected request id on response")
	}
}

func TestInternalAPIKey(t *testing.T) {
	e := echo.New()
	g := e.Group("", InternalAPIKey("secret"))
	g.GET("/x", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no key: want 401 got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	req2.Header.Set("Authorization", "Bearer secret")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("bearer: want 200 got %d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/x", nil)
	req3.Header.Set("X-Internal-Key", "secret")
	rec3 := httptest.NewRecorder()
	e.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("x-internal-key: want 200 got %d", rec3.Code)
	}
}

func TestInternalAPIKey_EmptyExpected(t *testing.T) {
	e := echo.New()
	g := e.Group("", InternalAPIKey(""))
	g.GET("/x", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("empty expected should pass: %d", rec.Code)
	}
}

func TestCallbackSecret(t *testing.T) {
	e := echo.New()
	g := e.Group("", CallbackSecret("cb"))
	g.POST("/x", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/x", nil)
	req2.Header.Set("X-Callback-Secret", "cb")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", rec2.Code)
	}
}

func TestCallbackSecret_Empty(t *testing.T) {
	e := echo.New()
	g := e.Group("", CallbackSecret(""))
	g.POST("/x", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", rec.Code)
	}
}

func TestAccessLog(t *testing.T) {
	e := echo.New()
	e.Use(AccessLog())
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusTeapot) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("code=%d", rec.Code)
	}
}
