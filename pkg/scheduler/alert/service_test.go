package alert

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/message"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

func TestFetchAndEnqueueSkipsAlertAlreadyQueued(t *testing.T) {
	const id = "a1"
	repo := &fakeAlertRepo{alert: alerts.Alert{Id: id}}
	s := newTestScheduler(repo, 1)

	s.fetchAndEnqueue()
	mustReturn(t, s.fetchAndEnqueue)

	if len(s.queue) != 1 {
		t.Fatalf("queue len = %d, want 1", len(s.queue))
	}
	got := <-s.queue
	if got.Id != id {
		t.Fatalf("queued id = %q, want %q", got.Id, id)
	}
}

func TestFetchAndEnqueueRetriesAfterWorkerFinishes(t *testing.T) {
	const id = "a1"
	// MarkAlertAsSent blocks inside handleAlert and leaves the alert unsent.
	release := make(chan struct{})
	entered := make(chan struct{}, 1)
	repo := &fakeAlertRepo{
		alert:   alerts.Alert{Id: id},
		entered: entered,
		release: release,
	}
	s := newTestScheduler(repo, 1)
	s.wgWorkers.Add(1)
	go s.worker(0)
	defer func() {
		close(release)
		close(s.queue)
		s.wgWorkers.Wait()
	}()

	s.fetchAndEnqueue()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("worker did not enter handleAlert")
	}
	if len(s.queue) != 0 {
		t.Fatalf("queue len = %d before second poll, want 0", len(s.queue))
	}

	mustReturn(t, s.fetchAndEnqueue)
	if len(s.queue) != 0 {
		t.Fatal("alert queued while worker is still handling it")
	}

	release <- struct{}{}
	deadline := time.Now().Add(time.Second)
	for tracking(s, id) {
		if time.Now().After(deadline) {
			t.Fatal("id still tracked after handleAlert returned")
		}
		time.Sleep(time.Millisecond)
	}

	mustReturn(t, s.fetchAndEnqueue)
	if !tracking(s, id) {
		t.Fatal("unsent alert was not queued again after the worker finished")
	}
}

func handleScheduler(repo *fakeAlertRepo, groups map[string]map[string]string, ps stubProviders) *Scheduler {
	return NewScheduler(nil, mapReceptor(groups), repo, ps, fakeMessage{}, zap.NewNop().Sugar(),
		SchedulerConfig{QueueSize: 1, Workers: 1, Interval: time.Hour})
}

func providers(ns ...*fakeNotifier) stubProviders {
	var ps stubProviders
	for _, n := range ns {
		ps.list = append(ps.list, notifications.Providers{Name: n.flag, Flag: n.flag, Status: true, Provider: n})
	}
	return ps
}

func TestHandleAlertGetProviderErrorStaysUnsent(t *testing.T) {
	want := errors.New("providers down")
	repo := &fakeAlertRepo{}
	ps := stubProviders{err: want}
	s := handleScheduler(repo, nil, ps)

	err := s.handleAlert(alerts.Alert{Id: "a", Method: []string{"sms"}, Receptor: []string{"g1"}})

	if err != want {
		t.Fatalf("err = %v, want %v", err, want)
	}
	if repo.marked != 0 {
		t.Fatalf("marked = %d, want 0", repo.marked)
	}
}

func TestHandleAlertCacheMissStaysUnsent(t *testing.T) {
	// g1 misses, g2 is cached but its Send fails: nothing delivered, so stay unsent.
	n := &fakeNotifier{flag: "sms", err: errors.New("boom")}
	repo := &fakeAlertRepo{}
	s := handleScheduler(repo, map[string]map[string]string{"g2": {"u2": "0912"}}, providers(n))

	err := s.handleAlert(alerts.Alert{Id: "a", Method: []string{"sms"}, Receptor: []string{"g1", "g2"}})

	if err == nil || !strings.Contains(err.Error(), "g1") {
		t.Fatalf("err = %v, want it to name g1", err)
	}
	if n.sends != 1 {
		t.Fatalf("sends = %d, want 1 (g2 must still be tried after the g1 miss)", n.sends)
	}
	if repo.marked != 0 {
		t.Fatalf("marked = %d, want 0", repo.marked)
	}
}

func TestHandleAlertSendErrorStaysUnsent(t *testing.T) {
	want := errors.New("send failed")
	n := &fakeNotifier{flag: "sms", err: want}
	repo := &fakeAlertRepo{}
	s := handleScheduler(repo, map[string]map[string]string{"g1": {"u1": "0912"}}, providers(n))

	err := s.handleAlert(alerts.Alert{Id: "a", Method: []string{"sms"}, Receptor: []string{"g1"}})

	if err != want {
		t.Fatalf("err = %v, want %v", err, want)
	}
	if repo.marked != 0 {
		t.Fatalf("marked = %d, want 0", repo.marked)
	}
}

func TestHandleAlertEmptyMethodMarksSent(t *testing.T) {
	repo := &fakeAlertRepo{}
	s := handleScheduler(repo, nil, stubProviders{})

	if err := s.handleAlert(alerts.Alert{Id: "a", Receptor: []string{"g1"}}); err != nil {
		t.Fatal(err)
	}
	if repo.marked != 1 {
		t.Fatalf("marked = %d, want 1", repo.marked)
	}
}

func TestHandleAlertPartialSuccessMarksSent(t *testing.T) {
	ok := &fakeNotifier{flag: "sms", prio: 1}
	bad := &fakeNotifier{flag: "email", prio: 2, err: errors.New("smtp down")}
	repo := &fakeAlertRepo{}
	groups := map[string]map[string]string{"g1": {"u1": "0912"}}
	s := handleScheduler(repo, groups, providers(ok, bad))

	err := s.handleAlert(alerts.Alert{Id: "a", Method: []string{"sms", "email"}, Receptor: []string{"g1"}})

	if err != nil {
		t.Fatal(err)
	}
	if ok.sends != 1 || bad.sends != 1 {
		t.Fatalf("sends ok=%d bad=%d, want 1/1", ok.sends, bad.sends)
	}
	if repo.marked != 1 {
		t.Fatalf("marked = %d, want 1", repo.marked)
	}
}

func TestHandleAlertTelegramNilErrorStoresDelivered(t *testing.T) {
	n := &fakeNotifier{flag: "telegram", name: "Telegram", ids: []string{"42"}}
	repo := &fakeAlertRepo{}
	s := handleScheduler(repo, map[string]map[string]string{"g1": {"u1": "100"}}, providers(n))
	log := &msgLog{}
	s.messageRepo = log

	if err := s.handleAlert(alerts.Alert{Id: "a", Method: []string{"telegram"}, Receptor: []string{"g1"}}); err != nil {
		t.Fatal(err)
	}
	if len(log.rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(log.rows))
	}
	r := log.rows[0]
	if r.Status != "Delivered" || r.SenderId != "42" || r.UserId != "u1" || r.Receptor != "100" {
		t.Fatalf("row = %+v, want Delivered/42/u1/100", r)
	}
}

func TestHandleAlertPerRecipientErrorsInSendOrder(t *testing.T) {
	n := &fakeNotifier{flag: "telegram", name: "Telegram", ids: []string{"1", "2"}, err: errors.New("nil;boom;")}
	repo := &fakeAlertRepo{}
	users := map[string]string{"u1": "100", "u2": "200"}
	s := handleScheduler(repo, map[string]map[string]string{"g1": users}, providers(n))
	log := &msgLog{}
	s.messageRepo = log

	_ = s.handleAlert(alerts.Alert{Id: "a", Method: []string{"telegram"}, Receptor: []string{"g1"}})

	if len(log.rows) != 2 || len(n.got) != 2 {
		t.Fatalf("rows = %d, sent receptors = %v, want 2 each", len(log.rows), n.got)
	}
	for i, r := range log.rows {
		if r.Receptor != n.got[i] || users[r.UserId] != r.Receptor {
			t.Fatalf("row %d = %+v, want receptor %s with its user", i, r, n.got[i])
		}
	}
	if r := log.rows[0]; r.Status != "Delivered" || r.SenderId != "1" {
		t.Fatalf("row 0 = %+v, want Delivered with sender id 1", r)
	}
	if r := log.rows[1]; r.Status != "Failed" || r.Response != "boom" {
		t.Fatalf("row 1 = %+v, want Failed with response boom", r)
	}
	if repo.marked != 1 {
		t.Fatalf("marked = %d, want 1: partial delivery must not be resent", repo.marked)
	}
}

func TestHandleAlertPlainErrorStoresFailed(t *testing.T) {
	n := &fakeNotifier{flag: "sms", err: errors.New("gateway down")}
	repo := &fakeAlertRepo{}
	s := handleScheduler(repo, map[string]map[string]string{"g1": {"u1": "0912"}}, providers(n))
	log := &msgLog{}
	s.messageRepo = log

	_ = s.handleAlert(alerts.Alert{Id: "a", Method: []string{"sms"}, Receptor: []string{"g1"}})

	if len(log.rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(log.rows))
	}
	if r := log.rows[0]; r.Status != "Failed" || r.Response != "gateway down" || r.UserId != "u1" {
		t.Fatalf("row = %+v, want Failed/gateway down/u1", r)
	}
}

func mustReturn(t *testing.T, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		fn()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("call blocked")
	}
}

func tracking(s *Scheduler, id string) bool {
	s.inflightMu.Lock()
	defer s.inflightMu.Unlock()
	_, ok := s.inflight[id]
	return ok
}

func newTestScheduler(repo alerts.AlertRepository, queueSize int) *Scheduler {
	return NewScheduler(
		nil,
		fakeReceptor{},
		repo,
		fakeProviderStatus{},
		fakeMessage{},
		zap.NewNop().Sugar(),
		SchedulerConfig{QueueSize: queueSize, Workers: 1, Interval: time.Hour},
	)
}

type fakeAlertRepo struct {
	alert   alerts.Alert
	entered chan struct{}
	release chan struct{}
	marked  int
}

func (f *fakeAlertRepo) GetUnsentAlerts() ([]alerts.Alert, error) {
	return []alerts.Alert{f.alert}, nil
}

func (f *fakeAlertRepo) MarkAlertAsSent(string) error {
	f.marked++
	if f.entered != nil {
		select {
		case f.entered <- struct{}{}:
		default:
		}
	}
	if f.release != nil {
		<-f.release
	}
	return nil
}

func (f *fakeAlertRepo) AddAlert(*alerts.Alert) (int64, error) { return 0, nil }
func (f *fakeAlertRepo) GetAlertById(string) (*alerts.Alert, error) {
	return nil, nil
}
func (f *fakeAlertRepo) AlertsBySeverity() ([]*alerts.AlertsBySeverity, error) {
	return nil, nil
}
func (f *fakeAlertRepo) GetAlerts(string, string, int, int) ([]*alerts.Alert, error) {
	return nil, nil
}
func (f *fakeAlertRepo) GetUnsentAlertID(alerts.Alert) (string, error) { return "", nil }
func (f *fakeAlertRepo) GetAlertByFingerPrintAndStatus(string, string) (*alerts.Alert, error) {
	return nil, nil
}

type fakeReceptor struct{}

func (fakeReceptor) GetNumbers(string) (map[string]string, error) { return nil, nil }
func (fakeReceptor) Get(string, string) (map[string]string, bool) { return nil, false }

type fakeMessage struct{}

func (fakeMessage) Add(*message.Message) error { return nil }

type msgLog struct{ rows []*message.Message }

func (l *msgLog) Add(m *message.Message) error {
	l.rows = append(l.rows, m)
	return nil
}

type fakeProviderStatus struct{}

func (fakeProviderStatus) GetActiveProviders() ([]notifications.Providers, error) {
	return nil, nil
}
func (fakeProviderStatus) GetProvidersPriority() ([]notifications.Providers, error) {
	return nil, nil
}

// mapReceptor serves cached receptors per group; unknown groups miss.
type mapReceptor map[string]map[string]string

func (mapReceptor) GetNumbers(string) (map[string]string, error) { return nil, nil }
func (m mapReceptor) Get(_ string, g string) (map[string]string, bool) {
	r, ok := m[g]
	return r, ok
}

type stubProviders struct {
	list []notifications.Providers
	err  error
}

func (stubProviders) GetActiveProviders() ([]notifications.Providers, error) { return nil, nil }
func (p stubProviders) GetProvidersPriority() ([]notifications.Providers, error) {
	return p.list, p.err
}

type fakeNotifier struct {
	flag  string
	name  string // defaults to flag
	prio  int
	ids   []string // defaults to ["id"]
	err   error
	sends int
	got   []string // receptors of the last Send
}

func (n *fakeNotifier) Send(m notifications.Message) ([]string, error) {
	n.sends++
	n.got = m.Receptors
	if n.ids != nil {
		return n.ids, n.err
	}
	return []string{"id"}, n.err
}
func (n *fakeNotifier) Status(string) (notifications.MessageStatusType, error) {
	var st notifications.MessageStatusType
	return st, nil
}
func (n *fakeNotifier) Verify() (string, error) { return "", nil }
func (n *fakeNotifier) GetName() string {
	if n.name != "" {
		return n.name
	}
	return n.flag
}
func (n *fakeNotifier) GetFlag() string  { return n.flag }
func (n *fakeNotifier) GetPriority() int { return n.prio }
