package cache_receptors

import (
	"testing"
	"time"

	"github.com/root-ali/iris/pkg/cache"
	"go.uber.org/zap"
)

type mockRepo struct{}

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
	},
	{
		GroupID:   "group1",
		GroupName: "testing",
		UserId:    "user3",
		Mobile:    "1111111111",
	},
}

func (m *mockRepo) GetPerGroupIds(...string) ([]GroupWithMobiles, error) {
	return groupMobiles, nil
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

func newMockService() (*CacheReceptor, error) {
	repo := &mockRepo{}
	logger := mockLogger()
	c := cache.New[string, map[string]string](logger, cache.WithCapacity(3))
	config := Config{
		Interval:  1 * time.Minute,
		Workers:   2,
		QueueSize: 5,
	}
	return NewCacheReceptorsScheduler(repo, c, logger, config)
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

}
