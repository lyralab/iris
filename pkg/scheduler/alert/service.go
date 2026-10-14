package alert

import (
	"cmp"
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
	s.logger.Debug("Polling for unsent alerts")
	unsent, err := s.repo.GetUnsentAlerts()
	if err != nil {
		s.logger.Errorw("Cannot load unsent alerts from the database", "error", err)
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
			s.logger.Errorw("Alert was not delivered",
				"worker", id, "alertID", al.Id, "name", al.Name, "methods", al.Method, "error", err)
		}
	}
}

func (s *Scheduler) handleAlert(al alerts.Alert) error {
	s.logger.Infow("Processing alert",
		"alertID", al.Id, "name", al.Name, "methods", al.Method, "receptors", al.Receptor)

	// Check if alert has method and receptor to set candidate fot sending alert
	if len(al.Method) == 0 || len(al.Receptor) == 0 {
		s.logger.Warnw("Alert has no method or no receptor; marking it as sent without notifying",
			"alertID", al.Id, "methods", al.Method, "receptors", al.Receptor)
		err := s.repo.MarkAlertAsSent(al.Id)
		if err != nil {
			return err
		}
		return nil
	}

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
		s.logger.Errorw("Cannot send alert: no usable notification provider for its methods",
			"alertID", al.Id, "methods", al.Method, "error", err)
		return err
	}
	s.logger.Infow("Sending alert to providers", "alertID", al.Id, "methods", al.Method, "candidateProviders", len(provider))
	var (
		sent    bool // at least one Send succeeded
		sendErr error
		missed  []string                // groups with no cached receptors
		skipped []string                // providers that already failed for this alert
		done    = make(map[string]bool) // flags already delivered; lower priority providers of that flag are fallbacks
	)

	for _, p := range provider {
		if done[p.GetFlag()] {
			s.logger.Debugw("Skipping fallback provider: alert already delivered on this channel",
				"alertID", al.Id, "provider", p.GetName(), "channel", p.GetFlag())
			continue
		}
		if s.failed.has(al.Id, p.GetName()) {
			s.logger.Warnw("Skipping provider: it already failed for this alert and is not retried",
				"alertID", al.Id, "provider", p.GetName(), "channel", p.GetFlag())
			skipped = append(skipped, p.GetName())
			continue
		}
		var receptors []string
		receptorIds := make(map[string]string)
		s.logger.Infow("Trying provider to send alert",
			"alertID", al.Id,
			"provider", p.GetName(),
			"channel", p.GetFlag(),
			"priority", p.GetPriority())

		for _, r := range al.Receptor {
			cacheReceptors, ok := s.receptorRepo.Get(p.GetFlag(), r)
			s.logger.Debugw("Looked up group receptors in cache",
				"alertID", al.Id,
				"channel", p.GetFlag(),
				"group", r,
				"receptors", cacheReceptors,
				"found", ok)
			if !ok {
				s.logger.Errorw("Group has no cached receptors for this channel; alert will be retried",
					"alertID", al.Id, "channel", p.GetFlag(), "group", r)
				missed = append(missed, r)
				continue
			}
			for k, v := range cacheReceptors {
				if slices.Contains(receptors, v) {
					s.logger.Debugw("Skipping duplicate receptor",
						"alertID", al.Id, "channel", p.GetFlag(), "receptor", v)
					continue
				}
				receptorIds[k] = v
				receptors = append(receptors, v)
			}
		}
		if len(receptors) == 0 {
			s.logger.Warnw("Provider skipped: no receptors resolved for this channel",
				"alertID", al.Id, "provider", p.GetName(), "channel", p.GetFlag(), "groups", al.Receptor)
			continue
		}
		msg.Receptors = receptors

		msgIds, err := p.Send(msg)

		// Per-recipient results, "nil" = delivered. Set only when err.Error() is
		// shaped "r1;r2;...;" with one non-empty token per recipient (Telegram does this).
		var results []string
		if err != nil {
			s.logger.Errorw("Provider failed to send alert",
				"alertID", al.Id, "provider", p.GetName(), "channel", p.GetFlag(),
				"receptors", len(receptors), "error", err)
			if sendErr == nil {
				sendErr = err
			}
			results = strings.Split(strings.TrimSuffix(err.Error(), ";"), ";")
			if len(results) != len(receptors) || slices.Contains(results, "") {
				results = nil
			}
			// Partial delivery still counts as sent; retrying would resend to recipients who already got it.
			if slices.Contains(results, "nil") {
				sent, done[p.GetFlag()] = true, true
			}
			if !done[p.GetFlag()] {
				s.failed.add(al.Id, p.GetName())
				s.logger.Warnw("Provider marked as failed for this alert; trying the next provider of the channel if any",
					"alertID", al.Id, "provider", p.GetName(), "channel", p.GetFlag())
			}
		} else {
			sent, done[p.GetFlag()] = true, true
			s.logger.Infow("Provider accepted alert",
				"alertID", al.Id, "provider", p.GetName(), "channel", p.GetFlag(),
				"receptors", len(receptors), "messageIDs", msgIds)
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
				s.logger.Errorw("Cannot save message record",
					"alertID", al.Id, "provider", p.GetName(), "receptor", v, "error", err)
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
	// Reaching here without delivery means every provider already failed once: give up instead of retrying forever.
	if sent {
		s.logger.Infow("Alert delivered, marking it as sent", "alertID", al.Id, "channels", done)
	} else {
		s.logger.Errorw("Alert not delivered by any provider; giving up (failed attempts are stored in message)",
			"alertID", al.Id, "methods", al.Method, "alreadyFailedProviders", skipped)
	}
	s.failed.forget(al.Id)
	return s.repo.MarkAlertAsSent(al.Id)
}

// getProvider returns every active, constructed provider for the given flags ordered by priority
// then name, so the caller can fall back to the next provider of a flag when one fails.
func (s *Scheduler) getProvider(flags []string, _ int) ([]notifications.NotificationInterface, error) {
	providers, err := s.provider.GetProvidersPriority()
	if err != nil {
		s.logger.Errorw("Cannot load providers from the provider service", "error", err)
		return nil, err
	}
	var active []notifications.Providers
	for _, p := range providers {
		// Provider is nil when the row is active but the service is disabled in config.
		if p.Status && p.Provider != nil && slices.Contains(flags, p.Flag) {
			s.logger.Debugw("Candidate provider", "name", p.Name, "channel", p.Flag, "priority", p.Priority)
			active = append(active, p)
		}
	}
	slices.SortFunc(active, func(a, b notifications.Providers) int {
		return cmp.Or(a.Priority-b.Priority, strings.Compare(a.Name, b.Name))
	})

	ni := make([]notifications.NotificationInterface, 0, len(active))
	for _, p := range active {
		ni = append(ni, p.Provider)
	}
	if len(ni) > 0 {
		return ni, nil
	}
	s.logger.Warnw("No active provider with a working implementation for these methods", "methods", flags)
	return nil, errors.New("no active provider found")
}
