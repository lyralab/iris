package cache_receptors

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/root-ali/iris/pkg/cache"
	"github.com/root-ali/iris/pkg/scheduler"
	"go.uber.org/zap"
)

func NewCacheReceptorsScheduler(
	repository Repository,
	cacheService cache.Interface[string, []string],
	logger *zap.SugaredLogger,
	config Config,
) (scheduler.ServiceInterface, error) {
	if config.Interval <= 0 {
		return nil, errors.New("interval must be > 0")
	}
	if config.Workers < 1 {
		return nil, errors.New("workers must be >= 1")
	}
	if config.QueueSize <= 0 {
		config.QueueSize = config.Workers
	}

	return &cacheReceptor{
		Repository: repository,
		Cache:      cacheService,
		conf:       config,
		taskCh:     make(chan struct{}, config.QueueSize),
		wg:         sync.WaitGroup{},
		ctx:        context.Background(),
		Logger:     logger,
	}, nil
}

func (s *cacheReceptor) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return errors.New("scheduler already started")
	}
	s.started = true
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.taskCh = make(chan struct{}, s.conf.QueueSize)

	// Start workers.
	for i := 0; i < s.conf.Workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	// Start scheduling loop.
	go s.run()

	return nil
}

func (s *cacheReceptor) Stop() error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	s.started = false
	cancel := s.cancel
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	s.wg.Wait()
	return nil
}

func (s *cacheReceptor) setMobilesOnCache() {

	s.Logger.Info("Starting Cache Receptors Service")

	gn, err := s.Repository.GetGroupNumbers()
	if err != nil {
		s.Logger.Errorw("Failed to get group numbers", "error", err)
		return
	}
	for _, group := range gn {
		s.Logger.Infow("Group with mobiles",
			"group_id", group.GroupID,
			"group_name", group.GroupName,
			"mobiles", group.Mobiles)
		mobiles := []string(group.Mobiles)
		s.Cache.Delete("mobiles_" + group.GroupName)
		if err := s.Cache.Set("mobiles_"+group.GroupName, mobiles, 0); err != nil {
			s.Logger.Errorw("Failed to set mobiles in cache",
				"group_id", group.GroupName,
				"error", err)
			return
		}
	}

}

func (s *cacheReceptor) worker(id int) {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case _, ok := <-s.taskCh:
			if !ok {
				return
			}
			s.safeRunJob()
		}
	}
}

func (s *cacheReceptor) safeRunJob() {
	defer func() {
		if r := recover(); r != nil {
			_ = fmt.Sprintf("panic in job: %v", r)
		}
	}()
	s.setMobilesOnCache()
}

func (s *cacheReceptor) run() {
	// Initial start handling
	if s.conf.StartAt.IsZero() {
		s.enqueueCache()
	} else {
		if d := time.Until(s.conf.StartAt); d > 0 {
			timer := time.NewTimer(d)
			select {
			case <-timer.C:
			case <-s.ctx.Done():
				timer.Stop()
				return
			}
		}
		s.enqueueCache()
	}

	ticker := time.NewTicker(s.conf.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.enqueueCache()
		}
	}
}

func (s *cacheReceptor) enqueueCache() {
	for {
		select {
		case s.taskCh <- struct{}{}:
			return
		case <-s.ctx.Done():
			return
		}
	}
}
