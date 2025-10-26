package cache_receptors

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/root-ali/iris/pkg/cache"
)

type Repository interface {
	GetGroupsNumbers(...string) ([]GroupWithMobiles, error)
	GetGroupEmails() (string, []string, error)
	GetUserEmail() (string, []string, error)
	GetUserNumber() (string, []string, error)
}

type CacheService interface {
	GetNumbers(group string) (map[string]string, error)
}

type CacheReceptor struct {
	Repository Repository
	Cache      cache.Interface[string, map[string]string]

	conf   Config
	ctx    context.Context
	cancel context.CancelFunc

	mu      sync.Mutex
	started bool

	taskCh chan struct{}
	wg     sync.WaitGroup

	Logger *zap.SugaredLogger
}

type GroupWithMobiles struct {
	GroupID   string `gorm:"column:group_id"`
	GroupName string `gorm:"column:group_name"`
	UserId    string `gorm:"column:user_id"`
	Mobile    string `gorm:"column:mobiles,type:varchar"`
}

type Config struct {
	StartAt   time.Time
	Interval  time.Duration
	Workers   int
	QueueSize int
}
