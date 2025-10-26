package cache_receptors

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/root-ali/iris/pkg/cache"
	"go.uber.org/zap"
)

func NewCacheReceptorsScheduler(
	repository Repository,
	cacheService cache.Interface[string, map[string]string],
	logger *zap.SugaredLogger,
	config Config,
) (*CacheReceptor, error) {
	if config.Interval <= 0 {
		return nil, errors.New("interval must be > 0")
	}
	if config.Workers < 1 {
		return nil, errors.New("workers must be >= 1")
	}
	if config.QueueSize <= 0 {
		config.QueueSize = config.Workers
	}

	return &CacheReceptor{
		Repository: repository,
		Cache:      cacheService,
		conf:       config,
		taskCh:     make(chan struct{}, config.QueueSize),
		wg:         sync.WaitGroup{},
		ctx:        context.Background(),
		Logger:     logger,
	}, nil
}

func (s *CacheReceptor) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return errors.New("scheduler already started")
	}
	s.started = true
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.taskCh = make(chan struct{}, s.conf.QueueSize)

	for i := 0; i < s.conf.Workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	go s.run()

	return nil
}

func (s *CacheReceptor) Stop() error {
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

func (s *CacheReceptor) worker(id int) {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case _, ok := <-s.taskCh:
			if !ok {
				return
			}
			s.Logger.Infow("Starting worker job", "id", id)
			s.safeRunJob()
		}
	}
}

func (s *CacheReceptor) run() {
	s.Logger.Info("Starting Cache Receptors Scheduler...")
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

func (s *CacheReceptor) enqueueCache() {
	for {
		select {
		case s.taskCh <- struct{}{}:
			return
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *CacheReceptor) safeRunJob() {
	defer func() {
		if r := recover(); r != nil {
			_ = fmt.Sprintf("panic in job: %v", r)
		}
	}()
	s.Logger.Info("Starting safe run job on cache receptor job...")
	s.setMobilesOnCache()
}

func (s *CacheReceptor) setMobilesOnCache() {
	s.Logger.Info("Starting Cache Receptors Job at %v", time.Now())

	results, err := s.Repository.GetGroupsNumbers()
	if err != nil {
		s.Logger.Errorw("Failed to get group numbers", "error", err)
		return
	}
	s.Logger.Info("length group numbers", "length", len(results))

	cached := make(map[string]map[string]string)
	for _, group := range results {
		fmt.Println("Processing group:", group.GroupName)
		s.Logger.Infow("Group with mobiles",
			"group_id", group.GroupID,
			"group_name", group.GroupName,
			"mobiles", group.Mobile)
		if cached[group.GroupName] == nil {
			cached[group.GroupName] = make(map[string]string)
		}
		cached[group.GroupName][group.UserId] = group.Mobile
	}

	for groupName, _ := range cached {
		err := s.Cache.Set("mobiles_"+groupName, cached[groupName], 0)
		if err != nil {
			return
		}
	}

}

func (s *CacheReceptor) GetNumbers(name string) (map[string]string, error) {
	mobiles, ok := s.Cache.Get("mobiles_" + name)
	if !ok {
		s.setMobilesOnCache()
		mobiles, ok = s.Cache.Get("mobiles_" + name)
		if !ok {
			return nil, fmt.Errorf("no mobiles found for group: %s", name)
		}
		return mobiles, nil
	}
	return mobiles, nil
}
