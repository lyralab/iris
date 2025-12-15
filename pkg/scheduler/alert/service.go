package alert

import (
	"errors"
	"slices"
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
	s.ticker = time.NewTicker(5 * time.Second)
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
	s.logger.Infow("Processing alert",
		"alertID", al.Id, "name", al.Name, "methods", al.Method, "receptors", al.Receptor)

	// Check if alert has method and receptor to set candidate fot sending alert
	if len(al.Method) == 0 || len(al.Receptor) == 0 {
		s.logger.Warnw("Alert has no method or no receptor defined, skipping", "alertID", al.Id)
		err := s.repo.MarkAlertAsSent(al.Id)
		if err != nil {
			return err
		}
		return nil
	}

	// Prepare receptors list and mapping for receptor users
	var receptors []string

	userMessage := make(map[string]bool)

	// Prepare Message
	msg := notifications.Message{
		Subject: al.Name,
		Message: al.Description,
		State:   al.Status,
		Time:    time.Now().Format(time.DateTime),
	}

	saveTextMsg := msg.State + ":" + msg.Subject + ":" + msg.Message

	provider, err := s.getProvider(al.Method, 0)
	if err != nil {
		s.logger.Errorw("Failed to get provider", "error", err)
		err = s.repo.MarkAlertAsSent(al.Id)
		if err != nil {
			return err
		}
		return err
	}

	for _, p := range provider {
		receptorIds := make(map[string]string)
		s.logger.Infow("Using provider for alert",
			"alertID", al.Id,
			"provider", p.GetName())

		for _, r := range al.Receptor {
			cacheReceptors, ok := s.receptorRepo.Get(p.GetFlag(), r)
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
				if ok := userMessage[k]; ok {
					s.logger.Debugw("User already has message prepared, skipping receptor",
						"method", al.Method,
						"user", k,
						"receptor", v)
					continue
				}
				if slices.Contains(receptors, v) {
					s.logger.Debugw("Receptor already added, skipping",
						"method", al.Method,
						"receptor", v)
					continue
				}
				receptorIds[k] = v
				userMessage[k] = false
				receptors = append(receptors, v)
			}
			s.logger.Debugw("Prepared receptors for alert",
				"method", al.Method,
				"receptors", receptors)
		}
		if len(receptors) == 0 {
			s.logger.Warnw("No receptors found for alert and method, skipping",
				"alertID", al.Id,
				"method", al.Method)
			continue
		}
		msg.Receptors = receptors

		msgIds, err := p.Send(msg)
		if p.GetName() == "Telegram" {
			s.logger.Infow("Telegram notification details", "receptors", receptors,
				"errors", err.Error())
			telegramErrors := strings.Split(err.Error(), ";")
			i := 0
			for userId, receptor := range receptorIds {
				if telegramErrors[i] != "nil" {
					// Save Message
					failedMsg := message.NewMessage("",
						saveTextMsg,
						receptor, p.GetName(),
						userId,
						"",
						telegramErrors[i],
						[]string{p.GetName()},
						message.TypeMessageStatusSent)
					err := s.messageRepo.Add(failedMsg)
					if err != nil {
						s.logger.Errorw("Failed to save failed message", "receptor", receptor, "error", err)
						continue
					}
				} else {
					// Save Message
					sentMsg := message.NewMessage(msgIds[i],
						saveTextMsg,
						receptor,
						p.GetName(),
						userId,
						"",
						"Delivered",
						[]string{p.GetName()},
						message.TypeMessageStatusDelivered)
					err := s.messageRepo.Add(sentMsg)
					if err != nil {
						s.logger.Errorw("Failed to save sent message", "receptor", receptor, "error", err)
						continue
					}
				}
				i++
			}

			continue
		} else if err != nil {
			s.logger.Errorw("Failed to send notification", "error", err)
			for r, v := range receptorIds {

				// Save Message
				failedMsg := message.NewMessage("",
					saveTextMsg,
					v, p.GetName(),
					r,
					r,
					err.Error(),
					[]string{p.GetName()},
					message.TypeMessageStatusFailed)
				err := s.messageRepo.Add(failedMsg)
				if err != nil {
					s.logger.Errorw("Failed to save failed message", "receptor", v, "error", err)
					return err
				}

			}
			continue
		}
		var i = 0
		for r, v := range receptorIds {

			// Save Message
			userMessage[r] = true
			sentMsg := message.NewMessage(msgIds[i],
				saveTextMsg,
				v,
				p.GetName(),
				r,
				"",
				"Sent",
				[]string{p.GetName()},
				message.TypeMessageStatusSent)
			err := s.messageRepo.Add(sentMsg)
			if err != nil {
				s.logger.Errorw("Failed to save sent message", "receptor", v, "error", err)
				continue
			}
			i++
		}

		continue
	}

	// Mark alert as sent
	return s.repo.MarkAlertAsSent(al.Id)
}

func (s *Scheduler) getProvider(flags []string, _ int) ([]notifications.NotificationInterface, error) {
	providers, err := s.provider.GetProvidersPriority()
	if err != nil {
		s.logger.Errorw("Failed to get providers", "error", err)
		return nil, err
	}
	ni := make([]notifications.NotificationInterface, 0)
	returnProviders := make([]string, 0)
	for _, flag := range flags {
		for _, p := range providers {
			if slices.Contains(returnProviders, p.Flag) {
				continue
			}
			if p.Status == true && p.Flag == flag {
				s.logger.Debugw("Found active provider", "name", p.Name, "flag", flag)
				ni = append(ni, p.Provider)
				returnProviders = append(returnProviders, p.Flag)
			}
		}
	}

	// Sort providers by priority
	slices.SortFunc(ni, func(a, b notifications.NotificationInterface) int {
		return a.GetPriority() - b.GetPriority()
	})

	if len(ni) > 0 {
		return ni, nil
	}
	s.logger.Warnw("No active provider found for flag", "flag", flags)
	return nil, errors.New("no active provider found")
}
