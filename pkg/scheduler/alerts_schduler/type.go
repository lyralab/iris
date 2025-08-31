package alerts_schduler

import (
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
	"time"
)

type Scheduler struct {
	smsProvider notifications.NotificationInterface
	repo        alerts.AlertRepository
	ticker      *time.Ticker
	done        chan bool
	logger      *zap.SugaredLogger
}

func NewScheduler(repo alerts.AlertRepository, smsProvider notifications.NotificationInterface, logger *zap.SugaredLogger) *Scheduler {
	return &Scheduler{
		repo:        repo,
		done:        make(chan bool),
		smsProvider: smsProvider,
		logger:      logger,
	}
}

func (s *Scheduler) Start() error {
	s.logger.Info("Alert Scheduler started...")
	s.ticker = time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-s.ticker.C:
				s.processAlerts()
			}
		}
	}()
	return nil
}

func (s *Scheduler) Stop() error {
	s.logger.Info("Scheduler stopping...")
	s.ticker.Stop()
	s.done <- true
	return nil
}

func (s *Scheduler) processAlerts() {
	s.logger.Info("Processing alerts...")
	unSentAlerts, err := s.repo.GetUnsentAlerts()
	if err != nil {
		s.logger.Errorw("Error getting unsent alerts", "error", err)
	}

	// Send notifications
	for _, al := range unSentAlerts {
		alertID, err := s.repo.GetUnsentAlertID(al)
		if err != nil {
			s.logger.Errorw("Error getting unsent alert ID", "alertID", al, "error", err)
			break
		}
		msg := notifications.Message{
			Subject:   al.Name,
			Message:   al.Description,
			Receptors: []string{al.Receptor}, // It could be a list of groups
		}
		_, err = s.smsProvider.Send(msg)
		if err != nil {
			s.logger.Errorw("Error sending alert to sms", "error", err)
			continue
		}
		// After sending, update DB
		err = s.repo.MarkAlertAsSent(alertID)
		if err != nil {
			s.logger.Errorw("Error marking alert as sent", "alertID", al, "error", err)
		}
	}
}
