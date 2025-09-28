package alert

import (
	"errors"
	"time"

	"github.com/avast/retry-go"
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/notifications"
)

func (s *Scheduler) Start() error {
	s.logger.Infow("Alert Scheduler started",
		"workers", s.cfg.Workers,
		"interval", s.cfg.Interval,
		"queueSize", s.cfg.QueueSize,
	)

	// start workers
	for i := 0; i < s.cfg.Workers; i++ {
		s.wgWorkers.Add(1)
		go s.worker(i)
	}

	// start producer loop
	s.ticker = time.NewTicker(s.cfg.Interval)
	s.wgLoop.Add(1)
	go s.loop()

	return nil
}

func (s *Scheduler) Stop() error {
	s.logger.Info("Scheduler stopping...")
	s.cancel()
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.wgLoop.Wait()
	close(s.queue)
	s.wgWorkers.Wait()
	s.logger.Info("Scheduler stopped.")
	return nil
}

func (s *Scheduler) loop() {
	defer s.wgLoop.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.ticker.C:
			s.fetchAndEnqueue()
		}
	}
}

func (s *Scheduler) fetchAndEnqueue() {
	s.logger.Debug("Fetching unsent alerts...")
	unsent, err := s.repo.GetUnsentAlerts()
	if err != nil {
		s.logger.Errorw("Error getting unsent alerts", "error", err)
		return
	}
	for _, al := range unsent {
		select {
		case s.queue <- al:
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) worker(id int) {
	defer s.wgWorkers.Done()
	for al := range s.queue {
		if err := s.handleAlert(al); err != nil {
			s.logger.Errorw("Failed processing alert", "worker", id, "alert", al, "error", err)
		}
	}
}

func (s *Scheduler) handleAlert(al alerts.Alert) error {
	msg := notifications.Message{
		Subject:   al.Name,
		Message:   al.Description,
		Receptors: []string{al.Receptor},
	}
	provider, err := s.getProvider(al.Method, 0)
	if err != nil {
		s.logger.Errorw("Failed to get provider", "error", err)
		return err
	}
	// Retry attempts
	retry.DefaultAttempts = 3
	// Retry Delay
	retry.DefaultDelay = 2 * time.Second
	err = retry.Do(
		func() error {
			mesgIDs, err := provider.Send(msg)
			for _, mesgID := range mesgIDs {
				status, err := provider.Status(mesgID)

				if err != nil {
					s.logger.Errorw("Failed to get message status", "error", err)
					return err
				}

				if status == notifications.TypeMessageStatusFailed {
					s.logger.Errorw("Message status is failed", "status", status)
					return errors.New("message status is failed")
				} else if status == notifications.TypeMessageStatusUndelivered {
					s.logger.Errorw("Message status is undelivered", "status", status)
					return errors.New("message status is undelivered")
				} else if status == notifications.TypeMessageStatusDelivered {
					s.logger.Infow("Message status is delivered", "status", status)
					return nil
				} else if status == notifications.TypeMessageStatusSent {
					s.logger.Infow("Message status is sent", "status", status)
					return nil
				}
			}

			if err != nil {
				s.logger.Errorw("Failed to send alert via provider",
					"provider", provider.GetName(), "error", err)
				return err
			}
			return nil
		},
	)
	if err != nil {
		s.logger.Errorw("Retry got limited with provider: ", "provider",
			provider.GetName(), "error", err)
		return err
	}
	return s.repo.MarkAlertAsSent(al.Id)
}

func (s *Scheduler) getProvider(flag string, _ int) (notifications.NotificationInterface, error) {
	s.logger.Debugw("Fetching provider for flag", "flag", flag)
	providers, err := s.provider.GetProvidersPriority()
	if err != nil {
		s.logger.Errorw("Failed to get providers", "error", err)
		return nil, err
	}
	for _, p := range providers {
		if p.Status == true && p.Flag == flag {
			s.logger.Debugw("Found active provider", "name", p.Name, "flag", flag)
			return p.Provider, nil
		}
	}
	s.logger.Warnw("No active provider found for flag", "flag", flag)
	return notifications.NotificationInterface(nil), errors.New("no active provider found")
}
