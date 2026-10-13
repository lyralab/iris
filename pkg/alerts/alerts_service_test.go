package alerts

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
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
		[]string{"mattermost"}, now, now, []string{"ops"})
	if err != nil {
		t.Fatalf("NewAlert err = %v", err)
	}
	return got
}

func TestNewAlertNoExistingFiringRow(t *testing.T) {
	got := newAlert(t, &fakeAlertRepo{err: gorm.ErrRecordNotFound}, "firing")

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
