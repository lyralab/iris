package alert

import (
	"context"
	"sync"
	"time"

	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

type SchedulerConfig struct {
	Interval  time.Duration
	Workers   int
	QueueSize int
}

type Scheduler struct {
	// deps
	smsProvider notifications.NotificationInterface
	repo        alerts.AlertRepository
	logger      *zap.SugaredLogger

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
	repo alerts.AlertRepository,
	smsProvider notifications.NotificationInterface,
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
		repo:        repo,
		smsProvider: smsProvider,
		logger:      logger,
		cfg:         cfg,
		ctx:         ctx,
		cancel:      cancel,
		queue:       make(chan alerts.Alert, cfg.QueueSize),
	}
}
