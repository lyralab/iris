package alert

import (
	"testing"
	"time"
)

// Admin silences (issue: silence alerts for all users). Seam: Scheduler.handleAlert.

func TestHandleAlertSilencedSkipsProvidersAndMarksSent(t *testing.T) {
	for _, status := range []string{"firing", "resolved"} {
		n := sms("Kavenegar", 1, nil)
		repo := &fakeAlertRepo{}
		s := fallbackScheduler(repo, smsProviders(n), &msgLog{})
		al := smsAlert()
		al.Status, al.Silenced = status, true

		if err := s.handleAlert(al); err != nil {
			t.Fatalf("%s: err = %v, want nil", status, err)
		}
		if n.sends != 0 {
			t.Fatalf("%s: sends = %d, want 0 for a silenced alert", status, n.sends)
		}
		if repo.marked != 1 {
			t.Fatalf("%s: marked = %d, want 1 so the state change is not retried", status, repo.marked)
		}
	}
}

func TestHandleAlertExpiredSilenceNotifies(t *testing.T) {
	n := sms("Kavenegar", 1, nil)
	repo := &fakeAlertRepo{}
	s := fallbackScheduler(repo, smsProviders(n), &msgLog{})
	past := time.Now().Add(-time.Minute)
	al := smsAlert()
	al.Status, al.Silenced, al.SilencedUntil = "firing", true, &past

	if err := s.handleAlert(al); err != nil {
		t.Fatalf("err = %v", err)
	}
	if n.sends != 1 {
		t.Fatalf("sends = %d, want 1: the silence expired", n.sends)
	}
}
