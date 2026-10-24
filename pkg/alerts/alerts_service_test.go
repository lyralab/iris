package alerts

import (
	"testing"
	"time"

	"github.com/root-ali/iris/internal/apperr"
	"go.uber.org/zap"
)

// fakeAlertRepo implements only GetAlertByFingerPrintAndStatus; other methods panic via the nil embedded interface.
type fakeAlertRepo struct {
	AlertRepository
	alert *Alert
	err   error
}

func (f *fakeAlertRepo) GetAlertByFingerPrintAndStatus(fingerPrint, status string) (*Alert, error) {
	return f.alert, f.err
}

func newAlert(t *testing.T, repo *fakeAlertRepo, status string) Alert {
	t.Helper()
	s := NewAlertService(zap.NewNop().Sugar(), repo)
	now := time.Now()
	got, err := s.NewAlert("fp1", "HighCPU", "critical", "cpu high", status,
		[]string{"mattermost"}, now, now, []string{"ops"}, Details{})
	if err != nil {
		t.Fatalf("NewAlert err = %v", err)
	}
	return got
}

func TestNewAlertNoExistingFiringRow(t *testing.T) {
	got := newAlert(t, &fakeAlertRepo{err: apperr.ErrAlertNotFound}, "firing")

	if got.Id == "" {
		t.Fatal("Id is empty, want new uuid")
	}
	if got.Status != "firing" {
		t.Fatalf("Status = %q, want firing", got.Status)
	}
	if got.SendNotif || got.Silenced {
		t.Fatalf("SendNotif = %v, Silenced = %v, want both false", got.SendNotif, got.Silenced)
	}
}

func TestNewAlertExistingSameStatusKeepsSendNotif(t *testing.T) {
	existing := &Alert{Id: "existing-id", FingerPrint: "fp1", Status: "firing", SendNotif: true}
	got := newAlert(t, &fakeAlertRepo{alert: existing}, "firing")

	if got.Id != "existing-id" || got.FingerPrint != "fp1" {
		t.Fatalf("Id, FingerPrint = %q, %q, want existing-id, fp1", got.Id, got.FingerPrint)
	}
	if !got.SendNotif {
		t.Fatal("SendNotif = false, want true")
	}
}

func TestNewAlertExistingFiringNowResolvedResetsSendNotif(t *testing.T) {
	existing := &Alert{Id: "existing-id", FingerPrint: "fp1", Status: "firing", SendNotif: true}
	got := newAlert(t, &fakeAlertRepo{alert: existing}, "resolved")

	if got.Id != "existing-id" {
		t.Fatalf("Id = %q, want existing-id", got.Id)
	}
	if got.Status != "resolved" {
		t.Fatalf("Status = %q, want resolved", got.Status)
	}
	if got.SendNotif {
		t.Fatal("SendNotif = true, want false")
	}
}

func TestStringMapRoundTrip(t *testing.T) {
	in := StringMap{"namespace": "batch", "job_name": "nightly-report"}
	v, err := in.Value()
	if err != nil {
		t.Fatalf("Value err = %v", err)
	}
	var out StringMap
	if err := out.Scan([]byte(v.(string))); err != nil {
		t.Fatalf("Scan err = %v", err)
	}
	if out["namespace"] != "batch" || out["job_name"] != "nightly-report" {
		t.Fatalf("round trip = %v", out)
	}

	var empty StringMap
	if v, _ := empty.Value(); v != nil {
		t.Fatalf("nil map Value = %v, want nil", v)
	}
	if err := out.Scan(nil); err != nil || out != nil {
		t.Fatalf("Scan(nil) = %v, %v, want nil map", out, err)
	}
}

func TestNewAlertExistingRowTakesLatestDetails(t *testing.T) {
	existing := &Alert{Id: "existing-id", FingerPrint: "fp1", Status: "firing"}
	s := NewAlertService(zap.NewNop().Sugar(), &fakeAlertRepo{alert: existing})
	now := time.Now()
	got, err := s.NewAlert("fp1", "HighCPU", "critical", "cpu high", "firing",
		nil, now, now, nil, Details{Labels: StringMap{"pod": "api-1"}, GeneratorURL: "http://prom"})
	if err != nil {
		t.Fatalf("NewAlert err = %v", err)
	}
	if got.Labels["pod"] != "api-1" || got.GeneratorURL != "http://prom" {
		t.Fatalf("Labels, GeneratorURL = %v, %q", got.Labels, got.GeneratorURL)
	}
}
