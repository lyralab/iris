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
	s.setOnCache()
}

// Refresh drops every cached receptor map and loads the current groups from the repository.
// Group create, delete, and membership changes call this so lookups do not keep the previous snapshot.
func (s *CacheReceptor) Refresh() {
	s.setOnCache()
}

func (s *CacheReceptor) setOnCache() {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()

	s.Logger.Info("Starting Cache Receptors Job at %v", time.Now())

	results, err := s.Repository.GetPerGroupIds()
	if err != nil {
		s.Logger.Errorw("Failed to get group numbers", "error", err)
		return
	}
	s.Logger.Infow("length group numbers", "length", len(results))

	mobileCached := make(map[string]map[string]string)
	emailCached := make(map[string]map[string]string)
	telegramCached := make(map[string]map[string]string)
	mattermostCached := make(map[string]map[string]string)
	baleCached := make(map[string]map[string]string)
	for _, group := range results {
		if mobileCached[group.GroupName] == nil {
			mobileCached[group.GroupName] = make(map[string]string)
		}
		if emailCached[group.GroupName] == nil {
			emailCached[group.GroupName] = make(map[string]string)
		}
		if telegramCached[group.GroupName] == nil {
			telegramCached[group.GroupName] = make(map[string]string)
		}
		if mattermostCached[group.GroupName] == nil {
			mattermostCached[group.GroupName] = make(map[string]string)
		}
		if baleCached[group.GroupName] == nil {
			baleCached[group.GroupName] = make(map[string]string)
		}
		if group.Mobile != "" {
			mobileCached[group.GroupName][group.UserId] = group.Mobile
		}
		if group.Email != "" {
			emailCached[group.GroupName][group.UserId] = group.Email
		}
		if group.TelegramID != "" {
			telegramCached[group.GroupName][group.UserId] = group.TelegramID
		}
		if group.MattermostID != "" {
			mattermostCached[group.GroupName][group.UserId] = group.MattermostID
		}
		if group.BaleID != "" {
			baleCached[group.GroupName][group.UserId] = group.BaleID
		}
	}

	// Miss first so a deleted group or a removed member cannot be served, then fill.
	for _, key := range s.keys {
		s.Cache.Delete(key)
	}
	s.keys = nil

	if err := s.store("mobiles_", mobileCached); err != nil {
		return
	}
	if err := s.store("emails_", emailCached); err != nil {
		return
	}
	if err := s.store("telegram_", telegramCached); err != nil {
		return
	}
	if err := s.store("mattermost_", mattermostCached); err != nil {
		return
	}
	if err := s.store("bale_", baleCached); err != nil {
		return
	}

	s.Logger.Info("Finished Cache Receptors Job at %v", time.Now())
}

func (s *CacheReceptor) store(prefix string, data map[string]map[string]string) error {
	for groupName, values := range data {
		key := prefix + groupName
		if err := s.Cache.Set(key, values, 0); err != nil {
			s.Logger.Errorw("failed to set receptor cache", "key", key, "error", err)
			return err
		}
		s.keys = append(s.keys, key)
	}
	return nil
}

func (s *CacheReceptor) GetNumbers(name string) (map[string]string, error) {
	mobiles, ok := s.Cache.Get("mobiles_" + name)
	if !ok {
		s.setOnCache()
		mobiles, ok = s.Cache.Get("mobiles_" + name)
		if !ok {
			return nil, fmt.Errorf("no mobiles found for group: %s", name)
		}
		return mobiles, nil
	}
	return mobiles, nil
}

func (s *CacheReceptor) Get(model string, groupName string) (map[string]string, bool) {
	query := ""
	switch model {
	case "sms":
		query = "mobiles_"
	case "mail":
		query = "emails_"
	case "telegram":
		query = "telegram_"
	case "mattermost":
		query = "mattermost_"
	case "bale":
		query = "bale_"
	default:
		return nil, false
	}
	s.Logger.Debugw("Getting receptors from cache", "model", model, "groupName", groupName, "query", query)
	question := query + groupName
	s.Logger.Debugw("Getting receptors from cache", "question", question)
	resp, ok := s.Cache.Get(question)
	s.Logger.Debugw("Getting receptors from cache", "resp", resp, "ok", ok)
	if !ok {
		s.setOnCache()
		resp, ok = s.Cache.Get(question)
		if !ok {
			return nil, false
		}
		return resp, true
	}
	return resp, true
}
