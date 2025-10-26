package message_status

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/root-ali/iris/pkg/message"
	"github.com/root-ali/iris/pkg/notifications"
	"github.com/root-ali/iris/pkg/scheduler"
	"go.uber.org/zap"
)

func NewMessageStatusMessageService(
	messageRepo MessageListRepository,
	providerRepo ProviderRepository,
	config Config,
	logger *zap.SugaredLogger,
) (scheduler.ServiceInterface, error) {
	s := &Service{}
	s.providerRepo = providerRepo
	s.messageRepo = messageRepo

	s.wg = sync.WaitGroup{}
	s.ctx = context.Background()
	s.config = config
	s.mu = sync.Mutex{}
	s.taskCh = make(chan message.Message, config.QueueSize)
	if config.Interval <= 0 {
		return nil, errors.New("interval must be > 0")
	}
	if config.Workers < 1 {
		return nil, errors.New("workers must be >= 1")
	}
	if config.QueueSize <= 0 {
		config.QueueSize = config.Workers
	}

	s.logger = logger
	return s, nil
}

func (s *Service) Start() error {
	s.logger.Infow("Starting message retry and check status scheduler",
		"workers", s.config.Workers,
		"interval", s.config.Interval,
		"queueSize", s.config.QueueSize,
		"startAt", s.config.StartAt)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return errors.New("service already started")
	}
	s.started = true
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.taskCh = make(chan message.Message, s.config.QueueSize)

	for i := 0; i < s.config.Workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	go s.run()

	return nil

}

func (s *Service) Stop() error {
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

func (s *Service) run() {
	if s.config.StartAt.IsZero() {
		s.enqueueCache()
	} else {
		if d := time.Until(s.config.StartAt); d > 0 {
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

	ticker := time.NewTicker(s.config.Interval)
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

func (s *Service) enqueueCache() {
	messages, err := s.messageRepo.ListNotFinishedMessages()
	if err != nil {
		s.logger.Errorf("failed to get not finished messages: %v", err)
		return
	}
	if len(messages) == 0 {
		s.logger.Info("no not finished messages found")
	}
	for _, msg := range messages {
		select {
		case s.taskCh <- msg:
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Service) worker(id int) {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg, ok := <-s.taskCh:
			if !ok {
				return
			}
			s.logger.Info("message status scheduler worker #" + strconv.Itoa(id) +
				" started for checking message " + msg.Id)
			s.safeRunJob(msg)
		}
	}
}

func (s *Service) safeRunJob(msg message.Message) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("panic in job: %v", r)
		}
	}()
	s.checkMessageStatus(msg)
}

func (s *Service) checkMessageStatus(msg message.Message) {
	if msg.Attempt >= MaxAttempt {
		s.sendAlternativeNotification(msg)
		return
	}

	s.logger.Infow("checking message with provider",
		"message", msg.Id,
		"senderId", msg.SenderId)
	provider, err := s.getNotificationProvider(msg.Sender)
	if err != nil {
		s.logger.Errorw("failed to get notification provider for message",
			"message", msg.Id, "error", err)
		return
	}
	s.logger.Infow("checking message status provider successfully get for message",
		"message", msg.Id,
		"provider", provider.GetName())
	messageStatus, err := provider.Status(msg.SenderId)
	if err != nil {
		s.logger.Errorw("failed to get message status from provider",
			"message", msg.Id, "error", err)
		err = s.messageRepo.UpdateMessageStatus(&msg,
			0,
			"Failed to get status from provider: "+err.Error())
		if err != nil {
			s.logger.Errorw("failed to update message status to Failed for message",
				"message", msg.Id, "error", err)
		}
		return
	}
	s.logger.Infow("got message status from provider",
		"message", msg.Id,
		"status", messageStatus)
	if messageStatus == 10 {
		err := s.messageRepo.UpdateMessageStatus(&msg, 10, "Delivered")
		if err != nil {
			s.logger.Errorw("failed to update message status to Delivered for message",
				"message", msg.Id, "error", err.Error())
		}
		return
	} else if messageStatus == 6 {
		err := s.messageRepo.UpdateMessageStatus(&msg, 6, "Failed")
		if err != nil {
			s.logger.Errorw("failed to update message status to Failed for message",
				"message", msg.Id, "error", err.Error())
		}
		return
	} else {
		err := s.messageRepo.UpdateMessageStatus(&msg, 0, "Sent")
		if err != nil {
			s.logger.Errorw("failed to update message status to Sent for message",
				"message", msg.Id, "error", err.Error())
		}
		return
	}
}

func (s *Service) sendAlternativeNotification(msg message.Message) {
	provider, err := s.getNotificationProvider(msg.Sender)
	if err != nil {
		s.logger.Errorw("failed to get notification provider for message",
			"message", msg.Id, "error", err)
		return
	}

	method := provider.GetFlag()
	alternativeProvider, err := s.getAlternativeProvider(method, msg.LastProviders)
	if err != nil {

		s.logger.Errorw("failed to get alternative provider for message",
			"message", msg.Id, "error", err)
		err := s.messageRepo.UpdateMessageStatus(&msg, 6, "failed to get alternative provider")
		if err != nil {
			return
		}

		return
	}

	s.logger.Infow("resending message with alternative provider",
		"message", msg.Id,
		"provider", alternativeProvider.GetName())
	msgId, err := s.sendNotification(msg, alternativeProvider)
	if err != nil {
		s.logger.Errorw("failed to resend message with alternative provider",
			"message", msg.Id, "error", err)
		return
	}
	err = s.messageRepo.UpdateMessageStatus(&msg,
		6,
		"Max attempts reached, trying alternative provider")
	if err != nil {
		s.logger.Errorw("failed to update message status to Failed for message before resending",
			"message", msg.Id, "error", err)
		return
	}
	newMessage := message.NewMessage(msgId,
		msg.Message,
		msg.Receptor,
		alternativeProvider.GetName(),
		msg.UserId,
		msg.GroupName,
		"Resent with alternative provider: "+alternativeProvider.GetName(),
		[]string{provider.GetName(), alternativeProvider.GetName()})
	err = s.messageRepo.Add(newMessage)
	if err != nil {
		s.logger.Errorw("failed to save resent message to repository",
			"message", msg.Id, "error", err)
	}
	return
}

func (s *Service) getNotificationProvider(providerName string) (notifications.NotificationInterface, error) {
	var provider notifications.NotificationInterface

	providers, err := s.providerRepo.GetActiveProviders()
	if err != nil {
		return provider, fmt.Errorf("failed to get active providers: %w", err)
	}
	if len(providers) == 0 {
		return provider, errors.New("no active providers found")
	}

	for _, p := range providers {
		if p.Provider.GetName() == providerName {
			provider = p.Provider
		}
	}

	s.logger.Infow("notification provider fetched for message status scheduler",
		"providerName", provider.GetName())

	return provider, nil
}

func (s *Service) checkStatus(msg message.Message,
	provider notifications.NotificationInterface) (notifications.MessageStatusType, error) {
	status, err := provider.Status(msg.SenderId)
	if err != nil {
		return 0, fmt.Errorf("failed to get message status from provider: %w", err)
	}
	return status, nil
}

func (s *Service) sendNotification(msg message.Message, provider notifications.NotificationInterface) (string, error) {
	notificationMessage := notifications.Message{
		Subject:   "",
		Message:   msg.Message,
		Receptors: []string{msg.Receptor},
	}

	msgIds, err := provider.Send(notificationMessage)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return msgIds[0], nil
}

func (s *Service) getAlternativeProvider(flag string, providerName []string) (notifications.NotificationInterface, error) {
	var provider notifications.NotificationInterface
	var providerMap = make(map[string]string)

	providers, err := s.providerRepo.GetActiveProviders()
	if err != nil {
		return provider, fmt.Errorf("failed to get active providers: %w", err)
	}
	if len(providers) == 0 {
		return provider, errors.New("no active providers found")
	}

	for _, p := range providers {
		providerMap[p.Provider.GetName()] = p.Flag
	}

	for _, name := range providerName {
		if _, ok := providerMap[name]; ok {
			delete(providerMap, name)
			continue
		}

	}

	if len(providerMap) > 0 {
		for name := range providerMap {
			if providerMap[name] == flag {
				for _, p := range providers {
					if p.Provider.GetName() == name {
						return p.Provider, nil
					}
				}
			} else {
				for _, p := range providers {
					if p.Provider.GetName() == name {
						return p.Provider, nil
					}
				}
			}
		}
	} else {
		return nil, fmt.Errorf("no alternative provider found")
	}

	return nil, fmt.Errorf("no alternative provider found")
}
