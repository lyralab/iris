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
		if !s.track(al.Id) {
			continue
		}
		select {
		case s.queue <- al:
		case <-s.ctx.Done():
			s.untrack(al.Id)
			return
		}
	}
}

func (s *Scheduler) track(id string) bool {
	s.inflightMu.Lock()
	defer s.inflightMu.Unlock()
	if _, ok := s.inflight[id]; ok {
		return false
	}
	s.inflight[id] = struct{}{}
	return true
}

func (s *Scheduler) untrack(id string) {
	s.inflightMu.Lock()
	delete(s.inflight, id)
	s.inflightMu.Unlock()
}

func (s *Scheduler) worker(id int) {
	defer s.wgWorkers.Done()
	for al := range s.queue {
		err := func() error {
			defer s.untrack(al.Id)
			return s.handleAlert(al)
		}()
		if err != nil {
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
		return err
	}

	var (
		sent    bool // at least one Send succeeded
		sendErr error
		missed  []string // groups with no cached receptors
	)

	for _, p := range provider {
		receptorIds := make(map[string]string)
		s.logger.Infow("Using provider for alert",
			"alertID", al.Id,
			"provider", p.GetName())

		for _, r := range al.Receptor {
			s.logger.Debugw("Getting receptors from cache", "method", al.Method, "receptor", r)
			cacheReceptors, ok := s.receptorRepo.Get(p.GetFlag(), r)
			s.logger.Debugw("Fetched receptors from cache",
				"method", al.Method,
				"receptor", r,
				"cached Receptors", cacheReceptors,
				"found", ok)
			if !ok {
				s.logger.Errorw("Failed to get receptors from group", "group", r)
				missed = append(missed, r)
				continue
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

		// Per-recipient results, "nil" = delivered. Set only when err.Error() is
		// shaped "r1;r2;...;" with one non-empty token per recipient (Telegram does this).
		var results []string
		if err != nil {
			s.logger.Errorw("Failed to send notification", "provider", p.GetName(), "error", err)
			if sendErr == nil {
				sendErr = err
			}
			results = strings.Split(strings.TrimSuffix(err.Error(), ";"), ";")
			if len(results) != len(receptors) || slices.Contains(results, "") {
				results = nil
			}
			// Partial delivery still counts as sent; retrying would resend to recipients who already got it.
			if slices.Contains(results, "nil") {
				sent = true
			}
		} else {
			sent = true
		}

		userIds := make(map[string]string, len(receptorIds))
		for u, c := range receptorIds {
			userIds[c] = u
		}
		for i, v := range receptors {
			senderId, response, status := "", "Sent", message.TypeMessageStatusSent
			switch {
			case err == nil || results != nil && results[i] == "nil":
				if i < len(msgIds) {
					senderId = msgIds[i]
				}
				if u, ok := userIds[v]; ok {
					userMessage[u] = true
				}
				// Telegram has no delivery receipts: an accepted API call is final, skip status polling.
				if p.GetFlag() == "telegram" {
					response, status = "Delivered", message.TypeMessageStatusDelivered
				}
			case results != nil:
				response, status = results[i], message.TypeMessageStatusFailed
			default:
				response, status = err.Error(), message.TypeMessageStatusFailed
			}
			m := message.NewMessage(senderId, saveTextMsg, v, p.GetName(), userIds[v], "",
				response, []string{p.GetName()}, status)
			if err := s.messageRepo.Add(m); err != nil {
				s.logger.Errorw("Failed to save message", "receptor", v, "error", err)
			}
		}
	}

	if !sent {
		var missErr error
		if len(missed) > 0 {
			missErr = errors.New("failed to get receptors from group: " + strings.Join(missed, ", "))
		}
		if missErr == nil && sendErr != nil {
			return sendErr
		}
		if err := errors.Join(sendErr, missErr); err != nil {
			return err
		}
	}
	// Once any channel delivered, mark sent so a later poll does not deliver it again.
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
