package schedulers

import (
	"time"

	"github.com/root-ali/iris/pkg/notifications"
	"github.com/root-ali/iris/pkg/scheduler/alert"
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"go.uber.org/zap"
)

func StartAlertScheduler(
	logger *zap.SugaredLogger,
	repos *postgresql.Storage,
	sms notifications.NotificationInterface,
	provider notifications.ProviderStatusInterface,
	interval time.Duration,
	workers, queue int,
) error {
	cfg := alert.SchedulerConfig{
		Interval:  interval,
		Workers:   workers,
		QueueSize: queue,
	}
	a := alert.NewScheduler(repos, sms, provider, logger, cfg)
	return a.Start()
}
