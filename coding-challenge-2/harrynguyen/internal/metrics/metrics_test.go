package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/labstack/echo/v4"
)

// scrapeCounter reads an unlabeled counter value from an exposition-format scrape body.
func scrapeCounter(body, metricLinePrefix string) float64 {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.HasPrefix(line, metricLinePrefix) {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			v, err := strconv.ParseFloat(fields[len(fields)-1], 64)
			if err != nil {
				continue
			}
			return v
		}
	}
	return -1
}

// scrapeRecoveryCounter returns 0 if the labeled series has not been created yet (CounterVec).
func scrapeRecoveryCounter(body, kind string) float64 {
	needle := `kind="` + kind + `"`
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.Contains(line, "sms_otp_sms_recovery_events_total") && strings.Contains(line, needle) {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			v, err := strconv.ParseFloat(fields[len(fields)-1], 64)
			if err != nil {
				continue
			}
			return v
		}
	}
	return 0
}

func TestHandler(t *testing.T) {
	h := Handler()
	if h == nil {
		t.Fatal("nil handler")
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics scrape: %d", rec.Code)
	}
}

func TestEchoMiddleware_RecordsNonMetrics(t *testing.T) {
	e := echo.New()
	e.Use(EchoMiddleware())
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestEchoMiddleware_SkipsMetricsPath(t *testing.T) {
	e := echo.New()
	e.Use(EchoMiddleware())
	e.GET("/metrics", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestEchoMiddleware_unregisteredPathUsesRequestURL(t *testing.T) {
	e := echo.New()
	e.Use(EchoMiddleware())
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok {
			_ = c.String(he.Code, he.Message.(string))
			return
		}
		_ = c.String(http.StatusInternalServerError, err.Error())
	}
	req := httptest.NewRequest(http.MethodGet, "/not-registered-path-xyz", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestWireSMSTelemetry(t *testing.T) {
	WireSMSTelemetry()
}

type metricsNoopEnqueue struct{}

func (metricsNoopEnqueue) Publish(context.Context, sms.SendJob) error { return nil }

func TestWireSMSTelemetry_messageCreatedCounterIncrements(t *testing.T) {
	WireSMSTelemetry()
	h := Handler()
	before := scrapePrometheus(t, h, "sms_otp_sms_messages_created_total")
	repo := sms.NewInMemRepository()
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	svc := sms.NewSMSService(repo, resolver, router, nil, metricsNoopEnqueue{})
	ctx := sms.WithSendSource(context.Background(), sms.SendSourceAPI)
	if _, err := svc.SendSMS(ctx, "PH", "6391712346622", "meta"); err != nil {
		t.Fatal(err)
	}
	after := scrapePrometheus(t, h, "sms_otp_sms_messages_created_total")
	if after != before+1 {
		t.Fatalf("SMSMessagesCreatedTotal: before=%v after=%v", before, after)
	}
}

func scrapePrometheus(t *testing.T, h http.Handler, metricLinePrefix string) float64 {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("scrape: status %d", rec.Code)
	}
	v := scrapeCounter(rec.Body.String(), metricLinePrefix)
	if v < 0 {
		t.Fatalf("metric not found: %q", metricLinePrefix)
	}
	return v
}

func TestWireSMSTelemetry_recoveryCounterIncrements(t *testing.T) {
	WireSMSTelemetry()
	kind := obslog.RecoveryKindRecoveredQueueToDelivered
	h := Handler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("scrape: status %d", rec.Code)
	}
	before := scrapeRecoveryCounter(rec.Body.String(), kind)

	ctx := context.Background()
	repo := sms.NewInMemRepository()
	msg := &providers.SMSMessage{
		ID: "m-tele", MessageID: "p-tele", Country: "PH", PhoneNumber: "6391712346623",
		Content: "x", Carrier: providers.CarrierGlobe, Provider: providers.ProviderTwilio,
		Status: providers.StatusQueue,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}
	svc := sms.NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	if err := svc.HandleCallback(ctx, "p-tele", providers.StatusSendSuccess, 0.01); err != nil {
		t.Fatal(err)
	}
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("scrape: status %d", rec2.Code)
	}
	after := scrapeRecoveryCounter(rec2.Body.String(), kind)
	if after != before+1 {
		t.Fatalf("recovery counter for %q: before=%v after=%v", kind, before, after)
	}
}

func TestWireCostTelemetry_Nil(t *testing.T) {
	WireCostTelemetry(nil)
}

var wireCostOnce sync.Once

func TestWireCostTelemetry_TrackerOnce(t *testing.T) {
	wireCostOnce.Do(func() {
		WireCostTelemetry(sms.NewInMemoryCostTracker())
	})
}

func TestWireCostTelemetry_gaugesAppearInMetricsScrape(t *testing.T) {
	wireCostOnce.Do(func() {
		WireCostTelemetry(sms.NewInMemoryCostTracker())
	})
	h := Handler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, "sms_otp_sms_total_estimated_cost") || !strings.Contains(body, "sms_otp_sms_total_actual_cost") {
		t.Fatalf("expected cost gauge family names in scrape body")
	}
}

func TestSMSStatusObserver(t *testing.T) {
	o := NewSMSStatusObserver()
	o.OnMessageUpdated(t.Context(), nil)
	o.OnMessageUpdated(t.Context(), &providers.SMSMessage{Status: providers.StatusQueue})
}
