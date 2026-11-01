package alert

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/alerts"
)

// metric returns the value of the series with exactly the given labels, 0 if absent.
func metric(t *testing.T, name string, labels map[string]string) float64 {
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
			got := map[string]string{}
			for _, l := range m.GetLabel() {
				got[l.GetName()] = l.GetValue()
			}
			if len(got) != len(labels) {
				continue
			}
			match := true
			for k, v := range labels {
				if got[k] != v {
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

func TestHandleAlertCountsSentNotificationPerRecipient(t *testing.T) {
	n := &fakeNotifier{flag: "sms", name: "metrics-sent-gw", ids: []string{"1", "2"}}
	s := handleScheduler(&fakeAlertRepo{}, map[string]map[string]string{"g1": {"u1": "0912", "u2": "0913"}}, providers(n))
	labels := map[string]string{"provider": "metrics-sent-gw", "channel": "sms", "result": "sent"}
	before := metric(t, "iris_notifications_total", labels)

	if err := s.handleAlert(alerts.Alert{Id: "a", Method: []string{"sms"}, Receptor: []string{"g1"}}); err != nil {
		t.Fatalf("err = %v", err)
	}

	if got := metric(t, "iris_notifications_total", labels) - before; got != 2 {
		t.Fatalf("sent delta = %v, want 2", got)
	}
}

func TestHandleAlertCountsFailedNotification(t *testing.T) {
	n := &fakeNotifier{flag: "sms", name: "metrics-fail-gw", err: apperr.New("gateway down")}
	s := handleScheduler(&fakeAlertRepo{}, map[string]map[string]string{"g1": {"u1": "0912"}}, providers(n))
	failed := map[string]string{"provider": "metrics-fail-gw", "channel": "sms", "result": "failed"}
	sent := map[string]string{"provider": "metrics-fail-gw", "channel": "sms", "result": "sent"}
	failedBefore := metric(t, "iris_notifications_total", failed)
	sentBefore := metric(t, "iris_notifications_total", sent)

	_ = s.handleAlert(alerts.Alert{Id: "a", Method: []string{"sms"}, Receptor: []string{"g1"}})

	if got := metric(t, "iris_notifications_total", failed) - failedBefore; got != 1 {
		t.Fatalf("failed delta = %v, want 1", got)
	}
	if got := metric(t, "iris_notifications_total", sent) - sentBefore; got != 0 {
		t.Fatalf("sent delta = %v, want 0", got)
	}
}

func TestHandleAlertCountsFallbackProviderSeparately(t *testing.T) {
	primary := &fakeNotifier{flag: "sms", name: "metrics-primary", prio: 1, err: apperr.New("down")}
	backup := &fakeNotifier{flag: "sms", name: "metrics-backup", prio: 2}
	s := handleScheduler(&fakeAlertRepo{}, map[string]map[string]string{"g1": {"u1": "0912"}}, providers(primary, backup))
	primaryFailed := map[string]string{"provider": "metrics-primary", "channel": "sms", "result": "failed"}
	backupSent := map[string]string{"provider": "metrics-backup", "channel": "sms", "result": "sent"}
	pBefore := metric(t, "iris_notifications_total", primaryFailed)
	bBefore := metric(t, "iris_notifications_total", backupSent)

	_ = s.handleAlert(alerts.Alert{Id: "a", Method: []string{"sms"}, Receptor: []string{"g1"}})

	if got := metric(t, "iris_notifications_total", primaryFailed) - pBefore; got != 1 {
		t.Fatalf("primary failed delta = %v, want 1", got)
	}
	if got := metric(t, "iris_notifications_total", backupSent) - bBefore; got != 1 {
		t.Fatalf("backup sent delta = %v, want 1", got)
	}
}

func TestFetchAndEnqueueSetsQueueGauges(t *testing.T) {
	s := newTestScheduler(&fakeAlertRepo{alert: alerts.Alert{Id: "a1"}}, 5)

	s.fetchAndEnqueue()

	if got := metric(t, "iris_alert_queue_depth", nil); got != 1 {
		t.Fatalf("queue depth = %v, want 1", got)
	}
	if got := metric(t, "iris_alerts_unsent", nil); got != 1 {
		t.Fatalf("unsent = %v, want 1", got)
	}
}
