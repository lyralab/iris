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

	// ponytail: in-memory set. A second replica can still deliver the same alert. Upgrade: a claim column.
	inflight   map[string]struct{}
	inflightMu sync.Mutex

	// providers that already failed to send an alert; each provider is tried once per alert.
	// ponytail: in-memory, a restart forgets it and tries the providers once more.
	failed failedProviders
}

// failedProviders remembers, per alert id, the providers whose send already failed.
type failedProviders struct {
	mu sync.Mutex
	by map[string]map[string]struct{}
}

func (f *failedProviders) add(alertID, provider string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.by == nil {
		f.by = make(map[string]map[string]struct{})
	}
	if f.by[alertID] == nil {
		f.by[alertID] = make(map[string]struct{})
	}
	f.by[alertID][provider] = struct{}{}
}

func (f *failedProviders) has(alertID, provider string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.by[alertID][provider]
	return ok
}

func (f *failedProviders) forget(alertID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.by, alertID)
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
		inflight:     make(map[string]struct{}),
	}
}
