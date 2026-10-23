package alert

import (
	"reflect"
	"testing"
	"time"

	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

// SMS provider fallback (issue #2). Seam: Scheduler.handleAlert, observed through
// the providers it calls, the message rows it stores and whether the alert is marked sent.

// callLog records the order in which providers are called.
type callLog struct{ names []string }

type recNotifier struct {
	*fakeNotifier
	log *callLog
}

func (n recNotifier) Send(m notifications.Message) ([]string, error) {
	n.log.names = append(n.log.names, n.GetName())
	return n.fakeNotifier.Send(m)
}

func sms(name string, prio int, err error) *fakeNotifier {
	return &fakeNotifier{flag: "sms", name: name, prio: prio, err: err}
}

// smsProviders lists providers the way the provider service does: one entry per provider,
// each with its own name and flag (the shared providers() helper names them by flag).
func smsProviders(ns ...*fakeNotifier) stubProviders {
	var ps stubProviders
	for _, n := range ns {
		ps.list = append(ps.list, notifications.Providers{
			Name: n.GetName(), Flag: n.flag, Priority: n.prio, Status: true, Provider: n,
		})
	}
	return ps
}

// flagReceptor serves different receptors per method, like the real cache (sms -> mobile, telegram -> chat id).
type flagReceptor map[string]map[string]map[string]string

func (f flagReceptor) GetNumbers(string) (map[string]string, error) { return nil, nil }
func (f flagReceptor) Get(flag, group string) (map[string]string, bool) {
	r, ok := f[flag][group]
	return r, ok
}

func fallbackScheduler(repo *fakeAlertRepo, ps stubProviders, log *msgLog) *Scheduler {
	s := handleScheduler(repo, map[string]map[string]string{"g1": {"u1": "09120000000"}}, ps)
	s.messageRepo = log
	return s
}

func smsAlert() alerts.Alert {
	return alerts.Alert{Id: "a", Method: []string{"sms"}, Receptor: []string{"g1"}}
}

func TestSmsFallbackFirstFailsSecondDelivers(t *testing.T) {
	first := sms("Asiatech", 1, apperr.New("cannot get token from asiatech"))
	second := sms("Kavenegar", 2, nil)
	repo, log := &fakeAlertRepo{}, &msgLog{}
	s := fallbackScheduler(repo, smsProviders(first, second), log)

	if err := s.handleAlert(smsAlert()); err != nil {
		t.Fatalf("err = %v, want nil: second sms provider delivered", err)
	}
	if first.sends != 1 || second.sends != 1 {
		t.Fatalf("sends first=%d second=%d, want 1/1", first.sends, second.sends)
	}
	if len(log.rows) != 2 {
		t.Fatalf("rows = %d, want 2 (one per attempt)", len(log.rows))
	}
	if r := log.rows[0]; r.Sender != "Asiatech" || r.Status != "Failed" {
		t.Fatalf("row 0 = %+v, want Asiatech/Failed", r)
	}
	if r := log.rows[1]; r.Sender != "Kavenegar" || r.Status != "Sent" {
		t.Fatalf("row 1 = %+v, want Kavenegar/Sent", r)
	}
	if repo.marked != 1 {
		t.Fatalf("marked = %d, want 1", repo.marked)
	}
}

func TestSmsFallbackStopsAtFirstSuccess(t *testing.T) {
	first := sms("Asiatech", 1, nil)
	second := sms("Kavenegar", 2, nil)
	repo, log := &fakeAlertRepo{}, &msgLog{}
	s := fallbackScheduler(repo, smsProviders(first, second), log)

	if err := s.handleAlert(smsAlert()); err != nil {
		t.Fatal(err)
	}
	if first.sends != 1 || second.sends != 0 {
		t.Fatalf("sends first=%d second=%d, want 1/0: no duplicate sms once delivered", first.sends, second.sends)
	}
	if len(log.rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(log.rows))
	}
}

func TestSmsFallbackAllProvidersFailStaysUnsent(t *testing.T) {
	first := sms("Asiatech", 1, apperr.New("asiatech down"))
	second := sms("Kavenegar", 2, apperr.New("kavenegar 416"))
	repo, log := &fakeAlertRepo{}, &msgLog{}
	s := fallbackScheduler(repo, smsProviders(first, second), log)

	if err := s.handleAlert(smsAlert()); err == nil {
		t.Fatal("err = nil, want the send error: nothing delivered")
	}
	if first.sends != 1 || second.sends != 1 {
		t.Fatalf("sends first=%d second=%d, want both tried once", first.sends, second.sends)
	}
	if len(log.rows) != 2 || log.rows[0].Status != "Failed" || log.rows[1].Status != "Failed" {
		t.Fatalf("rows = %+v, want two Failed rows", log.rows)
	}
	if repo.marked != 0 {
		t.Fatalf("marked = %d, want 0: alert must be retried", repo.marked)
	}
}

func TestSmsFallbackTriesLowerPriorityNumberFirst(t *testing.T) {
	log := &callLog{}
	low := recNotifier{sms("Kavenegar", 1, nil), log}
	high := recNotifier{sms("Asiatech", 3, nil), log}
	s := handleScheduler(&fakeAlertRepo{}, map[string]map[string]string{"g1": {"u1": "0912"}}, stubProviders{})
	// stub returns Asiatech first although Kavenegar has the lower priority number.
	s.provider = stubProviders{list: []notifications.Providers{
		{Name: "Asiatech", Flag: "sms", Priority: 3, Status: true, Provider: high},
		{Name: "Kavenegar", Flag: "sms", Priority: 1, Status: true, Provider: low},
	}}

	if err := s.handleAlert(smsAlert()); err != nil {
		t.Fatal(err)
	}
	if want := []string{"Kavenegar"}; !reflect.DeepEqual(log.names, want) {
		t.Fatalf("called %v, want %v", log.names, want)
	}
}

func TestSmsFallbackEqualPriorityOrderIsDeterministic(t *testing.T) {
	// Asiatech and Kavenegar both sit at priority 1 in the providers table. Which one goes first
	// must not depend on the order the repository happens to return rows in.
	for _, order := range [][]string{{"Asiatech", "Kavenegar"}, {"Kavenegar", "Asiatech"}} {
		log := &callLog{}
		byName := map[string]recNotifier{
			"Asiatech":  {sms("Asiatech", 1, nil), log},
			"Kavenegar": {sms("Kavenegar", 1, nil), log},
		}
		var list []notifications.Providers
		for _, n := range order {
			list = append(list, notifications.Providers{Name: n, Flag: "sms", Priority: 1, Status: true, Provider: byName[n]})
		}
		s := handleScheduler(&fakeAlertRepo{}, map[string]map[string]string{"g1": {"u1": "0912"}}, stubProviders{list: list})

		if err := s.handleAlert(smsAlert()); err != nil {
			t.Fatal(err)
		}
		if want := []string{"Asiatech"}; !reflect.DeepEqual(log.names, want) {
			t.Fatalf("repository order %v: called %v, want %v first", order, log.names, want)
		}
	}
}

func TestSmsFallbackDoesNotLeakReceptorsAcrossProviders(t *testing.T) {
	sms1 := sms("Asiatech", 1, apperr.New("asiatech down"))
	sms2 := sms("Kavenegar", 2, nil)
	tg := &fakeNotifier{flag: "telegram", name: "Telegram", prio: 3}
	recept := flagReceptor{
		"sms":      {"g1": {"u1": "09120000000"}},
		"telegram": {"g1": {"u1": "-1003444791590"}},
	}
	repo := &fakeAlertRepo{}
	s := NewScheduler(nil, recept, repo, smsProviders(sms1, sms2, tg), fakeMessage{}, zap.NewNop().Sugar(),
		SchedulerConfig{QueueSize: 1, Workers: 1, Interval: time.Hour})

	err := s.handleAlert(alerts.Alert{Id: "a", Method: []string{"sms", "telegram"}, Receptor: []string{"g1"}})

	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"09120000000"}; !reflect.DeepEqual(sms1.got, want) || !reflect.DeepEqual(sms2.got, want) {
		t.Fatalf("sms receptors = %v / %v, want only %v", sms1.got, sms2.got, want)
	}
	if want := []string{"-1003444791590"}; !reflect.DeepEqual(tg.got, want) {
		t.Fatalf("telegram receptors = %v, want only %v (no mobile numbers)", tg.got, want)
	}
	if repo.marked != 1 {
		t.Fatalf("marked = %d, want 1", repo.marked)
	}
}

func TestSmsFallbackProviderWithoutImplementationDoesNotPanic(t *testing.T) {
	// providers row is active but the service was never constructed (e.g. enabled: false in config.yml).
	ps := stubProviders{list: []notifications.Providers{
		{Name: "Kavenegar", Flag: "sms", Priority: 1, Status: true, Provider: nil},
	}}
	repo := &fakeAlertRepo{}
	s := handleScheduler(repo, map[string]map[string]string{"g1": {"u1": "0912"}}, ps)

	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("handleAlert panicked: %v", r)
			}
		}()
		err = s.handleAlert(smsAlert())
	}()

	if err == nil {
		t.Fatal("err = nil, want an error: no usable sms provider")
	}
	if repo.marked != 0 {
		t.Fatalf("marked = %d, want 0", repo.marked)
	}
}

func TestSmsFallbackSkipsProviderWithoutImplementation(t *testing.T) {
	broken := notifications.Providers{Name: "Kavenegar", Flag: "sms", Priority: 1, Status: true, Provider: nil}
	good := sms("Asiatech", 2, nil)
	ps := smsProviders(good)
	ps.list = append([]notifications.Providers{broken}, ps.list...)
	repo := &fakeAlertRepo{}
	s := handleScheduler(repo, map[string]map[string]string{"g1": {"u1": "0912"}}, ps)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("handleAlert panicked: %v", r)
		}
	}()
	if err := s.handleAlert(smsAlert()); err != nil {
		t.Fatal(err)
	}
	if good.sends != 1 || repo.marked != 1 {
		t.Fatalf("sends = %d marked = %d, want 1/1: usable provider must still deliver", good.sends, repo.marked)
	}
}

func TestSmsFallbackFailedProviderIsNotRetriedForSameAlert(t *testing.T) {
	first := sms("Asiatech", 1, apperr.New("cannot get token from asiatech"))
	second := sms("Kavenegar", 2, apperr.New("kavenegar 416"))
	repo, log := &fakeAlertRepo{}, &msgLog{}
	s := fallbackScheduler(repo, smsProviders(first, second), log)

	if err := s.handleAlert(smsAlert()); err == nil {
		t.Fatal("first pass: err = nil, want the send error")
	}
	// the scheduler polls unsent alerts again
	_ = s.handleAlert(smsAlert())

	if first.sends != 1 || second.sends != 1 {
		t.Fatalf("sends first=%d second=%d, want 1/1: each provider is tried once per alert", first.sends, second.sends)
	}
	if len(log.rows) != 2 {
		t.Fatalf("rows = %d, want 2 (no new Failed rows on the second poll)", len(log.rows))
	}
	if repo.marked != 1 {
		t.Fatalf("marked = %d, want 1: every provider failed, alert must leave the unsent queue", repo.marked)
	}
}

func TestSmsFallbackFailedProviderIsStillSkippedWhenAnotherWasNotReachedYet(t *testing.T) {
	// Asiatech fails on the first poll while the group has no cached receptors for Kavenegar's pass;
	// afterwards only Kavenegar may be called.
	first := sms("Asiatech", 1, apperr.New("cannot get token from asiatech"))
	second := sms("Kavenegar", 2, nil)
	repo := &fakeAlertRepo{}
	s := fallbackScheduler(repo, smsProviders(first, second), &msgLog{})
	s.failed.add("a", "Asiatech")

	if err := s.handleAlert(smsAlert()); err != nil {
		t.Fatal(err)
	}
	if first.sends != 0 || second.sends != 1 {
		t.Fatalf("sends first=%d second=%d, want 0/1", first.sends, second.sends)
	}
}
