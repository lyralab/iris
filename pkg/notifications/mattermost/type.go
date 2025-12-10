package mattermost

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/root-ali/iris/pkg/notifications"

	"go.uber.org/zap"
)

type service struct {
	client *model.Client4

	priority int

	logger *zap.SugaredLogger
}

type Config struct {
	Url      string
	BotToken string
	Priority int
}

type errorStack []error

func (e *errorStack) Append(err error) {
	*e = append(*e, err)
}

func (e errorStack) Error() string {
	errMsg := ""
	for _, err := range e {
		if err == nil {
			errMsg += "nil;"
		} else {
			errMsg += err.Error() + ";"
		}
	}
	return errMsg
}

func NewService(cfg Config, logger *zap.SugaredLogger) notifications.NotificationInterface {
	client := model.NewAPIv4Client(cfg.Url)
	client.SetToken(cfg.BotToken)

	return &service{
		client:   client,
		priority: cfg.Priority,
		logger:   logger,
	}
}
