package notifications

import (
	"errors"
	"testing"

	"github.com/root-ali/iris/pkg/cache"
	"go.uber.org/zap"
)

type fakeRepo struct {
	providers map[string]*Providers
	reads     int
	err       error
}

func (f *fakeRepo) AddProvider(p *Providers) error { f.providers[p.Name] = p; return f.err }
func (f *fakeRepo) ModifyProvider(p *Providers) error {
	if f.err != nil {
		return f.err
	}
	f.providers[p.Name].Status = p.Status
	return nil
}
func (f *fakeRepo) GetProvider(p *Providers) error { *p = *f.providers[p.Name]; return f.err }
func (f *fakeRepo) SetStatusFalse(p *Providers) error {
	if f.err != nil {
		return f.err
	}
	f.providers[p.Name].Status = false
	return nil
}
func (f *fakeRepo) GetProviders() ([]Providers, error) {
	f.reads++
	out := make([]Providers, 0, len(f.providers))
	for _, p := range f.providers {
		out = append(out, *p)
	}
	return out, f.err
}

func statusOf(t *testing.T, ps []Providers, name string) bool {
	t.Helper()
	for _, p := range ps {
		if p.Name == name {
			return p.Status
		}
	}
	t.Fatalf("provider %q not returned", name)
	return false
}

func TestDisableProviderInvalidatesCache(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := &fakeRepo{providers: map[string]*Providers{
		"sms":    {Name: "sms", Priority: 1, Status: true},
		"backup": {Name: "backup", Priority: 2, Status: false},
	}}
	c := cache.New[string, *[]Providers](logger, cache.WithJanitor(false))
	svc := NewProvidersService(repo, nil, c, logger)

	if _, err := svc.GetActiveProviders(); err != nil {
		t.Fatal(err)
	}
	ps, err := svc.GetProvidersPriority()
	if err != nil {
		t.Fatal(err)
	}
	if !statusOf(t, ps, "sms") {
		t.Fatal("sms should start active")
	}
	if _, ok := c.Get("active_providers"); !ok {
		t.Fatal("active_providers not cached")
	}

	if err := svc.DisableProvider("sms"); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Get("active_providers"); ok {
		t.Fatal("active_providers still cached after disable")
	}

	readsBefore := repo.reads
	ps, err = svc.GetProvidersPriority()
	if err != nil {
		t.Fatal(err)
	}
	if repo.reads != readsBefore+1 {
		t.Fatal("repository not re-read after disable")
	}
	if statusOf(t, ps, "sms") {
		t.Fatal("disabled sms still returned as active")
	}

	// Failed write must keep the cache: flip repo state behind the service's back,
	// then a failing DisableProvider must not drop the stale-but-cached list.
	repo.providers["sms"].Status = true
	repo.err = errors.New("db down")
	if err := svc.DisableProvider("backup"); !errors.Is(err, repo.err) {
		t.Fatalf("want repo error, got %v", err)
	}
	repo.err = nil

	readsBefore = repo.reads
	ps, err = svc.GetProvidersPriority()
	if err != nil {
		t.Fatal(err)
	}
	if repo.reads != readsBefore {
		t.Fatal("cache dropped after failed DisableProvider")
	}
	if statusOf(t, ps, "sms") {
		t.Fatal("expected cached list (sms disabled)")
	}
}
