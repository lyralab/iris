package alerts

import (
	"testing"
	"time"

	"github.com/root-ali/iris/internal/apperr"
	"go.uber.org/zap"
)

type silenceRepo struct {
	AlertRepository
	alert *Alert
	err   error

	calls    int
	silenced bool
	until    *time.Time
	by       string
}

func (r *silenceRepo) GetAlertById(string) (*Alert, error) { return r.alert, r.err }

func (r *silenceRepo) SetAlertSilence(_ string, silenced bool, until *time.Time, by string) error {
	r.calls++
	r.silenced, r.until, r.by = silenced, until, by
	return nil
}

func silenceService(r *silenceRepo) Service { return NewAlertService(zap.NewNop().Sugar(), r) }

func TestSilenceAlertTimed(t *testing.T) {
	r := &silenceRepo{alert: &Alert{Id: "a", Status: "firing"}}
	got, err := silenceService(r).SilenceAlert("a", "admin", time.Hour)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !r.silenced || r.by != "admin" || r.until == nil {
		t.Fatalf("repo got silenced=%v by=%q until=%v", r.silenced, r.by, r.until)
	}
	if d := time.Until(*r.until); d < 59*time.Minute || d > time.Hour {
		t.Fatalf("until in %v, want about 1h", d)
	}
	if !got.IsSilenced(time.Now()) || got.IsSilenced(time.Now().Add(2*time.Hour)) {
		t.Fatal("returned alert should be silenced now and not in 2h")
	}
}

func TestSilenceAlertUntilUnsilenced(t *testing.T) {
	r := &silenceRepo{alert: &Alert{Id: "a", Status: "firing"}}
	got, err := silenceService(r).SilenceAlert("a", "admin", 0)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if r.until != nil || !got.IsSilenced(time.Now().Add(365*24*time.Hour)) {
		t.Fatalf("until = %v, want nil (no expiry)", r.until)
	}
}

func TestSilenceAlertRejects(t *testing.T) {
	cases := []struct {
		name string
		repo *silenceRepo
		d    time.Duration
		want error
	}{
		{"resolved", &silenceRepo{alert: &Alert{Status: "resolved"}}, time.Hour, apperr.ErrAlertNotFiring},
		{"missing", &silenceRepo{err: apperr.ErrAlertNotFound}, time.Hour, apperr.ErrAlertNotFound},
		{"too long", &silenceRepo{alert: &Alert{Status: "firing"}}, MaxSilenceDuration + time.Hour, apperr.ErrInvalidSilenceDuration},
		{"negative", &silenceRepo{alert: &Alert{Status: "firing"}}, -time.Hour, apperr.ErrInvalidSilenceDuration},
	}
	for _, c := range cases {
		_, err := silenceService(c.repo).SilenceAlert("a", "admin", c.d)
		if !apperr.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
		if c.repo.calls != 0 {
			t.Errorf("%s: SetAlertSilence called, want no write", c.name)
		}
	}
}

func TestUnsilenceAlertClears(t *testing.T) {
	r := &silenceRepo{alert: &Alert{Id: "a", Status: "firing", Silenced: true, SilencedBy: "admin"}}
	got, err := silenceService(r).UnsilenceAlert("a")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if r.silenced || r.until != nil || r.by != "" || got.IsSilenced(time.Now()) {
		t.Fatalf("silence not cleared: repo silenced=%v by=%q, alert %+v", r.silenced, r.by, got)
	}
}

func TestNewAlertKeepsSilenceOfFiringRow(t *testing.T) {
	until := time.Now().Add(time.Hour)
	existing := &Alert{Id: "a", Status: "firing", Silenced: true, SilencedUntil: &until, SilencedBy: "admin"}
	got := newAlert(t, &fakeAlertRepo{alert: existing}, "resolved")
	if !got.Silenced || got.SilencedUntil != &until || got.SilencedBy != "admin" {
		t.Fatalf("silence lost on update: %+v", got)
	}
}
