package cache_receptors

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/root-ali/iris/pkg/cache"
	"go.uber.org/zap"
)

type mockRepo struct {
	mu    sync.Mutex
	rows  []GroupWithMobiles
	err   error
	calls int
}

var groupMobiles = []GroupWithMobiles{
	{
		GroupID:   "group1",
		GroupName: "testing",
		UserId:    "user1",
		Mobile:    "1234567890",
	},
	{
		GroupID:   "group2",
		GroupName: "admin",
		UserId:    "user2",
		Mobile:    "5555555555",
	},
	{
		GroupID:   "group1",
		GroupName: "testing",
		UserId:    "user2",
		Mobile:    "7777777777",
		BaleID:    "123123123",
	},
	{
		GroupID:   "group1",
		GroupName: "testing",
		UserId:    "user3",
		Mobile:    "1111111111",
	},
}

func (m *mockRepo) GetPerGroupIds(...string) ([]GroupWithMobiles, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	src := m.rows
	if src == nil {
		src = groupMobiles
	}
	out := make([]GroupWithMobiles, len(src))
	copy(out, src)
	return out, nil
}

func (m *mockRepo) setRows(rows []GroupWithMobiles) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows = rows
}

func (m *mockRepo) setErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

func (m *mockRepo) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}
func (m *mockRepo) GetGroupEmails() (string, []string, error) {
	return "", nil, nil
}
func (m *mockRepo) GetUserEmail() (string, []string, error) {
	return "", nil, nil
}
func (m *mockRepo) GetUserNumber() (string, []string, error) {
	return "", nil, nil
}

func mockLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

func newScheduler(repo Repository, interval time.Duration) (*CacheReceptor, error) {
	logger := mockLogger()
	c := cache.New[string, map[string]string](logger, cache.WithCapacity(64), cache.WithJanitor(false))
	return NewCacheReceptorsScheduler(repo, c, logger, Config{
		Interval:  interval,
		Workers:   1,
		QueueSize: 2,
	})
}

func newMockService() (*CacheReceptor, error) {
	return newScheduler(&mockRepo{}, time.Minute)
}

func cloneGroups(in []GroupWithMobiles) []GroupWithMobiles {
	out := make([]GroupWithMobiles, len(in))
	copy(out, in)
	return out
}

func waitFor(t *testing.T, timeout time.Duration, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("timed out waiting for cache scheduler")
}

func TestNewCacheReceptorsScheduler(t *testing.T) {
	t.Parallel()
	s, err := newMockService()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if s == nil {
		t.Fatal("expected service instance, got nil")
	}
	s.setOnCache()

	cached, err := s.GetNumbers("admin")
	if err != nil {
		t.Fatalf("expected no error getting cached numbers, got %v", err)
	}
	if len(cached) != 1 || cached["user2"] != "5555555555" {
		t.Fatalf("expected cached numbers to contain '5555555555', got %v", cached)
	}

	baleCached, ok := s.Get("bale", "testing")
	if !ok {
		t.Fatal("expected bale receptors to be cached")
	}
	if baleCached["user2"] != "123123123" {
		t.Fatalf("expected bale id 123123123, got %v", baleCached)
	}
}

func TestNewCacheReceptorsSchedulerRejectsConfig(t *testing.T) {
	t.Parallel()
	logger := mockLogger()
	c := cache.New[string, map[string]string](logger, cache.WithJanitor(false))
	repo := &mockRepo{}

	if _, err := NewCacheReceptorsScheduler(repo, c, logger, Config{Interval: 0, Workers: 1}); err == nil {
		t.Fatal("expected interval error")
	}
	if _, err := NewCacheReceptorsScheduler(repo, c, logger, Config{Interval: time.Second, Workers: 0}); err == nil {
		t.Fatal("expected workers error")
	}

	s, err := NewCacheReceptorsScheduler(repo, c, logger, Config{Interval: time.Second, Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	if s.conf.QueueSize != 2 {
		t.Fatalf("queue size = %d, want 2", s.conf.QueueSize)
	}
}

func TestSchedulerFillsOnStartAndStops(t *testing.T) {
	repo := &mockRepo{}
	s, err := newScheduler(repo, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err == nil {
		t.Fatal("expected second start to fail")
	}

	waitFor(t, 5*time.Second, func() bool {
		got, ok := s.Cache.Get("mobiles_admin")
		return ok && got["user2"] == "5555555555"
	})

	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}

	before := repo.callCount()
	time.Sleep(80 * time.Millisecond)
	if got := repo.callCount(); got != before {
		t.Fatalf("scheduler kept loading after stop: before %d after %d", before, got)
	}
}

func TestSchedulerRefillsOnInterval(t *testing.T) {
	repo := &mockRepo{}
	s, err := newScheduler(repo, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Stop() })

	waitFor(t, 5*time.Second, func() bool {
		got, ok := s.Cache.Get("mobiles_testing")
		return ok && got["user1"] == "1234567890"
	})

	repo.setRows([]GroupWithMobiles{
		{GroupName: "admin", UserId: "user2", Mobile: "0000000000"},
	})
	before := repo.callCount()
	waitFor(t, 5*time.Second, func() bool {
		if repo.callCount() <= before {
			return false
		}
		if _, ok := s.Cache.Get("mobiles_testing"); ok {
			return false
		}
		got, ok := s.Cache.Get("mobiles_admin")
		return ok && got["user2"] == "0000000000"
	})
}

func TestRefreshMissesAndFillsAfterGroupChange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		next  []GroupWithMobiles
		check func(t *testing.T, s *CacheReceptor)
	}{
		{
			name: "user added",
			next: append(cloneGroups(groupMobiles), GroupWithMobiles{
				GroupName: "admin",
				UserId:    "user9",
				Mobile:    "9999999999",
				Email:     "new@example.com",
			}),
			check: func(t *testing.T, s *CacheReceptor) {
				got, err := s.GetNumbers("admin")
				if err != nil {
					t.Fatal(err)
				}
				if got["user9"] != "9999999999" || got["user2"] != "5555555555" {
					t.Fatalf("admin mobiles = %v", got)
				}
				emails, ok := s.Get("mail", "admin")
				if !ok || emails["user9"] != "new@example.com" {
					t.Fatalf("admin emails = %v ok=%v", emails, ok)
				}
			},
		},
		{
			name: "user removed",
			next: []GroupWithMobiles{
				{GroupName: "admin", UserId: "user2", Mobile: "5555555555"},
				{GroupName: "testing", UserId: "user2", Mobile: "7777777777", BaleID: "123123123"},
				{GroupName: "testing", UserId: "user3", Mobile: "1111111111"},
			},
			check: func(t *testing.T, s *CacheReceptor) {
				got, err := s.GetNumbers("testing")
				if err != nil {
					t.Fatal(err)
				}
				if _, still := got["user1"]; still {
					t.Fatalf("removed user still cached: %v", got)
				}
				if got["user3"] != "1111111111" {
					t.Fatalf("testing mobiles = %v", got)
				}
			},
		},
		{
			name: "mobile replaced",
			next: []GroupWithMobiles{
				{GroupName: "admin", UserId: "user2", Mobile: "0000000000"},
				{GroupName: "testing", UserId: "user1", Mobile: "1234567890"},
			},
			check: func(t *testing.T, s *CacheReceptor) {
				got, err := s.GetNumbers("admin")
				if err != nil {
					t.Fatal(err)
				}
				if got["user2"] != "0000000000" {
					t.Fatalf("stale mobile kept: %v", got)
				}
			},
		},
		{
			name: "group added",
			next: append(cloneGroups(groupMobiles), GroupWithMobiles{
				GroupName: "ops",
				UserId:    "user8",
				Email:     "ops@example.com",
			}),
			check: func(t *testing.T, s *CacheReceptor) {
				emails, ok := s.Get("mail", "ops")
				if !ok || emails["user8"] != "ops@example.com" {
					t.Fatalf("ops emails = %v ok=%v", emails, ok)
				}
			},
		},
		{
			name: "empty group added",
			next: append(cloneGroups(groupMobiles), GroupWithMobiles{GroupName: "ops"}),
			check: func(t *testing.T, s *CacheReceptor) {
				got, ok := s.Get("sms", "ops")
				if !ok {
					t.Fatal("expected empty group to be cached")
				}
				if len(got) != 0 {
					t.Fatalf("ops mobiles = %v", got)
				}
			},
		},
		{
			name: "group deleted",
			next: []GroupWithMobiles{
				{GroupName: "admin", UserId: "user2", Mobile: "5555555555"},
			},
			check: func(t *testing.T, s *CacheReceptor) {
				if _, ok := s.Get("sms", "testing"); ok {
					t.Fatal("deleted group mobiles still cached")
				}
				if _, ok := s.Get("bale", "testing"); ok {
					t.Fatal("deleted group bale still cached")
				}
				if _, ok := s.Cache.Get("emails_testing"); ok {
					t.Fatal("deleted group emails key still present")
				}
				got, err := s.GetNumbers("admin")
				if err != nil || got["user2"] != "5555555555" {
					t.Fatalf("admin = %v err=%v", got, err)
				}
			},
		},
		{
			name: "all groups deleted",
			next: []GroupWithMobiles{},
			check: func(t *testing.T, s *CacheReceptor) {
				for _, key := range []string{
					"mobiles_admin", "emails_admin", "telegram_admin", "mattermost_admin", "bale_admin",
					"mobiles_testing", "emails_testing", "telegram_testing", "mattermost_testing", "bale_testing",
				} {
					if _, ok := s.Cache.Get(key); ok {
						t.Fatalf("key %s still cached", key)
					}
				}
				if _, err := s.GetNumbers("admin"); err == nil {
					t.Fatal("expected miss for deleted admin group")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockRepo{}
			s, err := newScheduler(repo, time.Hour)
			if err != nil {
				t.Fatal(err)
			}
			s.setOnCache()
			if _, ok := s.Cache.Get("mobiles_admin"); !ok {
				t.Fatal("expected initial cache fill")
			}

			repo.setRows(tt.next)
			s.Refresh()
			tt.check(t, s)
		})
	}
}

func TestRefreshUpdatesBeforeNextTick(t *testing.T) {
	repo := &mockRepo{}
	s, err := newScheduler(repo, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Stop() })

	waitFor(t, 5*time.Second, func() bool {
		got, ok := s.Cache.Get("mobiles_admin")
		return ok && got["user2"] == "5555555555"
	})

	repo.setRows([]GroupWithMobiles{
		{GroupName: "admin", UserId: "user4", Mobile: "4444444444"},
	})
	s.Refresh()

	if _, ok := s.Cache.Get("mobiles_testing"); ok {
		t.Fatal("testing group still cached before the next tick")
	}
	got, err := s.GetNumbers("admin")
	if err != nil {
		t.Fatal(err)
	}
	if got["user4"] != "4444444444" {
		t.Fatalf("admin mobiles = %v", got)
	}
	if _, stale := got["user2"]; stale {
		t.Fatalf("previous member still cached: %v", got)
	}
}

func TestRefreshKeepsSnapshotWhenRepositoryFails(t *testing.T) {
	t.Parallel()
	repo := &mockRepo{}
	s, err := newScheduler(repo, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	s.setOnCache()

	repo.setErr(errors.New("db down"))
	s.Refresh()

	got, ok := s.Cache.Get("mobiles_admin")
	if !ok || got["user2"] != "5555555555" {
		t.Fatalf("cache changed after a failed reload: %v ok=%v", got, ok)
	}
	if _, ok := s.Cache.Get("bale_testing"); !ok {
		t.Fatal("expected previous bale cache to stay after a failed reload")
	}
}
