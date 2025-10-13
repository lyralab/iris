package schedulers

import (
	"time"

	"github.com/root-ali/iris/pkg/cache"
	"github.com/root-ali/iris/pkg/notifications"
	"github.com/root-ali/iris/pkg/scheduler/alert"
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"go.uber.org/zap"
)

func StartAlertScheduler(
	logger *zap.SugaredLogger,
	repos *postgresql.Storage,
	receptor alert.ReceptorInterface,
	cache cache.Interface[string, []string],
	provider notifications.ProviderStatusInterface,
	interval time.Duration,
	workers, queue int,
) error {
	cfg := alert.SchedulerConfig{
		Interval:  interval,
		Workers:   workers,
		QueueSize: queue,
	}
	a := alert.NewScheduler(cache, receptor, repos, provider, logger, cfg)
	return a.Start()
}
