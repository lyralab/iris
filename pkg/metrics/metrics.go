// Package metrics exposes Iris business metrics on the default Prometheus registry,
// which the existing /metrics endpoint already serves.
//
// Label values are bounded: status and severity are mapped to a fixed set, and
// provider and channel come from the admin-managed providers table.
package metrics

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const other = "other"

var (
	alertsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "iris_alerts_received_total",
		Help: "Alerts saved from Alertmanager webhooks, by status and severity. Alertmanager re-sends still-firing alerts, so repeats count again.",
	}, []string{"status", "severity"})

	notifications = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "iris_notifications_total",
		Help: "Per-recipient notifications attempted, by provider, channel and result.",
	}, []string{"provider", "channel", "result"})

	silenced = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "iris_alerts_silenced_total",
		Help: "Alert state changes not notified because an admin silenced the alert, by status.",
	}, []string{"status"})

	queueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "iris_alert_queue_depth",
		Help: "Alerts waiting in the scheduler's in-memory queue, sampled after each poll and each dequeue.",
	})

	unsentAlerts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "iris_alerts_unsent",
		Help: "Alerts not yet marked sent (including queued and in-flight ones), sampled on each successful scheduler poll.",
	})

	knownStatuses   = map[string]struct{}{"firing": {}, "resolved": {}}
	knownSeverities = map[string]struct{}{
		"critical": {}, "warning": {}, "info": {}, "error": {}, "page": {}, "none": {},
	}
)

func bounded(v string, known map[string]struct{}) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if _, ok := known[v]; ok {
		return v
	}
	return other
}

// AlertReceived counts one alert saved from Alertmanager.
func AlertReceived(status, severity string) {
	alertsReceived.WithLabelValues(bounded(status, knownStatuses), bounded(severity, knownSeverities)).Inc()
}

// Notification counts one per-recipient delivery attempt.
func Notification(provider, channel string, sent bool) {
	result := "failed"
	if sent {
		result = "sent"
	}
	notifications.WithLabelValues(provider, channel, result).Inc()
}

// NotificationSilenced counts one alert state change skipped by a silence.
func NotificationSilenced(status string) {
	silenced.WithLabelValues(bounded(status, knownStatuses)).Inc()
}

// SetQueueDepth records the scheduler queue length.
func SetQueueDepth(n int) { queueDepth.Set(float64(n)) }

// SetUnsentAlerts records how many alerts are waiting for delivery.
func SetUnsentAlerts(n int) { unsentAlerts.Set(float64(n)) }
