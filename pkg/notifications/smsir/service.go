package smsir

import (
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

func NewSmsirService(apikey string, logger *zap.SugaredLogger) notifications.NotificationInterface {
	return &smsirService{ApiKey: apikey, Logger: logger}
}

func (s *smsirService) Send(message notifications.Message) ([]string, error) {
	return "", nil
}

func (s *smsirService) Status(string) (string, error) {
	return "", nil
}
