package schedulers

import (
	"time"

	"github.com/root-ali/iris/pkg/cache"
	"github.com/root-ali/iris/pkg/scheduler/cache_receptors"
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"go.uber.org/zap"
)

func StartCacheReceptors(
	logger *zap.SugaredLogger,
	repos *postgresql.Storage,
	startAtSeconds time.Duration,
	interval time.Duration,
	workers, queueSize, cacheCapacity int,
) (*cache_receptors.CacheReceptor, error) {
	logger.Debug("Starting cache receptor service...")
	c := cache.New[string, map[string]string](logger, cache.WithCapacity(cacheCapacity))
	cfg := cache_receptors.Config{
		StartAt:   time.Now().Add(startAtSeconds),
		Interval:  interval,
		Workers:   workers,
		QueueSize: queueSize,
	}
	cr, err := cache_receptors.NewCacheReceptorsScheduler(repos, c, logger, cfg)
	if err != nil {
		return nil, err
	}
	if err := cr.Start(); err != nil {
		return nil, err
	}
	logger.Info("Cache receptor service started")
	return cr, nil
}
