package metrics

import (
	"net/http"

	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metric names use namespace sms_otp_* for Grafana dashboards in this repo.

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "sms_otp",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "HTTP requests by method, Echo route template, and response status code.",
		},
		[]string{"method", "handler", "code"},
	)

	SMSMessagesCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "sms_otp",
			Subsystem: "sms",
			Name:      "messages_created_total",
			Help:      "SMS rows created in the pipeline (before provider send).",
		},
	)

	SMSRecoveryEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "sms_otp",
			Subsystem: "sms",
			Name:      "recovery_events_total",
			Help:      "Async callback recovery/mitigation paths (resilience testing and production webhooks).",
		},
		[]string{"kind"},
	)

	SMSMessageStatusTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "sms_otp",
			Subsystem: "sms",
			Name:      "message_status_updates_total",
			Help:      "SMS lifecycle transitions observed after persistence updates.",
		},
		[]string{"status"},
	)
)

// Handler exposes /metrics for Prometheus scraping.
func Handler() http.Handler {
	return promhttp.Handler()
}

// WireSMSTelemetry connects sms package hooks to Prometheus counters (call once at startup).
func WireSMSTelemetry() {
	sms.SetTelemetryHooks(
		func() { SMSMessagesCreatedTotal.Inc() },
		func(kind string) { SMSRecoveryEventsTotal.WithLabelValues(kind).Inc() },
	)
}

// WireCostTelemetry exposes in-memory estimated vs actual SMS cost gauges (call once at startup).
func WireCostTelemetry(tracker *sms.InMemoryCostTracker) {
	if tracker == nil {
		return
	}
	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "sms_otp",
			Subsystem: "sms",
			Name:      "total_estimated_cost",
			Help:      "Sum of estimated SMS cost when messages enter the queue (in-memory aggregate by provider).",
		},
		func() float64 { return tracker.Totals().TotalEstimatedCost },
	)
	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "sms_otp",
			Subsystem: "sms",
			Name:      "total_actual_cost",
			Help:      "Sum of incremental actual SMS cost from successful delivery callbacks (in-memory aggregate by provider).",
		},
		func() float64 { return tracker.Totals().TotalActualCost },
	)
}
