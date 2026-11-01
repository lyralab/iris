package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// sample returns the value of the series with exactly the given labels in the default registry, 0 if absent.
func sample(t *testing.T, name string, labels map[string]string) float64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Gather err = %v", err)
	}
	for _, f := range families {
		if f.GetName() != name {
			continue
		}
		for _, m := range f.GetMetric() {
			if len(m.GetLabel()) != len(labels) {
				continue
			}
			match := true
			for _, l := range m.GetLabel() {
				if labels[l.GetName()] != l.GetValue() {
					match = false
				}
			}
			if !match {
				continue
			}
			if m.GetCounter() != nil {
				return m.GetCounter().GetValue()
			}
			return m.GetGauge().GetValue()
		}
	}
	return 0
}

func TestAlertReceivedCountsByStatusAndSeverity(t *testing.T) {
	labels := map[string]string{"status": "firing", "severity": "critical"}
	before := sample(t, "iris_alerts_received_total", labels)

	AlertReceived("firing", "critical")
	AlertReceived("firing", "critical")

	if got := sample(t, "iris_alerts_received_total", labels) - before; got != 2 {
		t.Fatalf("delta = %v, want 2", got)
	}
}

func TestAlertReceivedBucketsUnknownValues(t *testing.T) {
	labels := map[string]string{"status": "other", "severity": "other"}
	before := sample(t, "iris_alerts_received_total", labels)

	AlertReceived("whatever-9f3a", "sev-17")

	if got := sample(t, "iris_alerts_received_total", labels) - before; got != 1 {
		t.Fatalf("delta = %v, want 1 under status=other,severity=other", got)
	}
}

func TestAlertReceivedNormalizesSeverityCase(t *testing.T) {
	labels := map[string]string{"status": "resolved", "severity": "warning"}
	before := sample(t, "iris_alerts_received_total", labels)

	AlertReceived("resolved", " Warning ")

	if got := sample(t, "iris_alerts_received_total", labels) - before; got != 1 {
		t.Fatalf("delta = %v, want 1", got)
	}
}

func TestNotificationCountsByProviderChannelAndResult(t *testing.T) {
	sent := map[string]string{"provider": "kavenegar", "channel": "sms", "result": "sent"}
	failed := map[string]string{"provider": "kavenegar", "channel": "sms", "result": "failed"}
	sentBefore := sample(t, "iris_notifications_total", sent)
	failedBefore := sample(t, "iris_notifications_total", failed)

	Notification("kavenegar", "sms", true)
	Notification("kavenegar", "sms", true)
	Notification("kavenegar", "sms", false)

	if got := sample(t, "iris_notifications_total", sent) - sentBefore; got != 2 {
		t.Fatalf("sent delta = %v, want 2", got)
	}
	if got := sample(t, "iris_notifications_total", failed) - failedBefore; got != 1 {
		t.Fatalf("failed delta = %v, want 1", got)
	}
}

func TestQueueGauges(t *testing.T) {
	SetQueueDepth(7)
	SetUnsentAlerts(12)

	if got := sample(t, "iris_alert_queue_depth", nil); got != 7 {
		t.Fatalf("queue depth = %v, want 7", got)
	}
	if got := sample(t, "iris_alerts_unsent", nil); got != 12 {
		t.Fatalf("unsent = %v, want 12", got)
	}
}
