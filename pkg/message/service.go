package message

import (
	"go.uber.org/zap"
)

func NewService(repo Repository, logger *zap.SugaredLogger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Add(msg *Message) error {
	return s.repo.SaveMessage(msg)
}

func (s *Service) List() ([]*Message, error) {
	return s.repo.ListMessages()
}

func (s *Service) UpdateMessageStatus(msg *Message, status StatusType, response string) error {
	if StatusMap[status] != msg.Status {
		msg.Attempt = 0
	} else {
		msg.Attempt++
	}
	msg.Status = StatusMap[status]
	msg.Response = response

	return s.repo.UpdateMessage(msg)
}

func (s *Service) ListNotFinishedMessages() ([]Message, error) {
	return s.repo.ListNotFinishedMessages()
}
