package alert

import (
	"errors"
	"strings"
	"time"

	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/message"
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
	if al.Method == "" || len(al.Receptor) == 0 {
		s.logger.Warnw("Alert has no method or no receptor defined, skipping", "alertID", al.Id)
		err := s.repo.MarkAlertAsSent(al.Id)
		if err != nil {
			return err
		}
		return nil
	}

	// Prepare receptors
	var receptors []string

	receptorIds := make(map[string]string)

	provider, err := s.getProvider(al.Method, 0)
	if err != nil {
		s.logger.Errorw("Failed to get provider", "error", err)
		err = s.repo.MarkAlertAsSent(al.Id)
		if err != nil {
			return err
		}
		return err
	}

	for _, r := range al.Receptor {
		cacheReceptors, ok := s.receptorRepo.Get(al.Method, r)
		s.logger.Debugw("Fetched receptors from cache",
			"method", al.Method,
			"receptor", r,
			"cached Receptors", cacheReceptors,
			"found", ok)
		if !ok {
			s.logger.Errorw("Failed to get receptors from group", "group", r, "error", err)
			err := s.repo.MarkAlertAsSent(al.Id)
			if err != nil {
				return errors.New("failed to mark alert as sent: " + err.Error())
			}
			return errors.New("failed to get receptors from group: " + r)
		}
		for k, v := range cacheReceptors {
			receptorIds[k] = v
			receptors = append(receptors, v)
		}
		s.logger.Debugw("Prepared receptors for alert",
			"method", al.Method,
			"receptors", receptors)
	}
	textMessage := al.Status + "\n" + al.Description + "\nTime: " + time.Now().Format(time.RFC1123)
	// Prepare message
	msg := notifications.Message{
		Subject:   al.Name,
		Message:   al.Description,
		State:     al.Status,
		Time:      time.Now().Format(time.DateTime),
		Receptors: receptors,
	}

	// Send notification
	msgIds, err := provider.Send(msg)

	if al.Method == "telegram" {
		s.logger.Infow("Telegram notification details", "receptors", receptors,
			"errors", err.Error())
		telegramErrors := strings.Split(err.Error(), ";")
		i := 0
		for userId, receptor := range receptorIds {
			if telegramErrors[i] != "nil" {
				// Save Message
				s.logger.Info("Save failed message to repository", "receptor", receptor)
				failedMsg := message.NewMessage("",
					al.Description,
					receptor, provider.GetName(),
					userId,
					"",
					telegramErrors[i],
					[]string{provider.GetName()},
					message.TypeMessageStatusSent)
				err := s.messageRepo.Add(failedMsg)
				if err != nil {
					s.logger.Errorw("Failed to save failed message", "receptor", receptor, "error", err)
					return err
				}
			} else {
				// Save Message
				s.logger.Info("Save success message to repository", "messageID", msgIds[i], "receptor", receptor)
				sentMsg := message.NewMessage(msgIds[i],
					textMessage,
					receptor,
					provider.GetName(),
					userId,
					"",
					"Delivered",
					[]string{provider.GetName()},
					message.TypeMessageStatusDelivered)
				err := s.messageRepo.Add(sentMsg)
				if err != nil {
					s.logger.Errorw("Failed to save sent message", "receptor", receptor, "error", err)
					return err
				}
			}
			i++
		}
		err = s.repo.MarkAlertAsSent(al.Id)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		s.logger.Errorw("Failed to send notification", "error", err)
		for r, v := range receptorIds {

			// Save Message
			s.logger.Info("Save failed message to repository", "receptor", v)
			failedMsg := message.NewMessage("",
				al.Description,
				v, provider.GetName(),
				r,
				r,
				err.Error(),
				[]string{provider.GetName()},
				message.TypeMessageStatusFailed)
			err := s.messageRepo.Add(failedMsg)
			if err != nil {
				s.logger.Errorw("Failed to save failed message", "receptor", v, "error", err)
				return err
			}
			err = s.repo.MarkAlertAsSent(al.Id)
			if err != nil {
				return err
			}
		}
		return err
	}
	var i = 0
	for r, v := range receptorIds {

		// Save Message
		s.logger.Info("Save success message to repository", "messageID", msgIds[i], "receptor", v)
		sentMsg := message.NewMessage(msgIds[i],
			textMessage,
			v,
			provider.GetName(),
			r,
			"",
			"Sent",
			[]string{provider.GetName()},
			message.TypeMessageStatusSent)
		err := s.messageRepo.Add(sentMsg)
		if err != nil {
			s.logger.Errorw("Failed to save sent message", "receptor", v, "error", err)
			return err
		}
		i++
	}

	// Mark alert as sent
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
