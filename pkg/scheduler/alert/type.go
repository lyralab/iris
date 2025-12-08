package alert

import (
	"context"
	"sync"
	"time"

	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/cache"
	"github.com/root-ali/iris/pkg/message"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

type SchedulerConfig struct {
	Interval  time.Duration
	Workers   int
	QueueSize int
}

type ReceptorInterface interface {
	GetNumbers(group string) (map[string]string, error)
	Get(model string, groupName string) (map[string]string, bool)
}

type MessageInterface interface {
	Add(msg *message.Message) error
}

type Scheduler struct {
	// dependencies
	cache        cache.Interface[string, []string]
	receptorRepo ReceptorInterface
	messageRepo  MessageInterface
	provider     notifications.ProviderStatusInterface
	repo         alerts.AlertRepository
	logger       *zap.SugaredLogger

	// config
	cfg SchedulerConfig

	// runtime
	ctx       context.Context
	cancel    context.CancelFunc
	queue     chan alerts.Alert
	wgWorkers sync.WaitGroup
	wgLoop    sync.WaitGroup
	ticker    *time.Ticker
}

func NewScheduler(
	c cache.Interface[string, []string],
	receptorRepo ReceptorInterface,
	repo alerts.AlertRepository,
	provider notifications.ProviderStatusInterface,
	messageRepo MessageInterface,
	logger *zap.SugaredLogger,
	cfg SchedulerConfig,
) *Scheduler {
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 10 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cache:        c,
		receptorRepo: receptorRepo,
		repo:         repo,
		provider:     provider,
		messageRepo:  messageRepo,
		logger:       logger,
		cfg:          cfg,
		ctx:          ctx,
		cancel:       cancel,
		queue:        make(chan alerts.Alert, cfg.QueueSize),
	}
}
