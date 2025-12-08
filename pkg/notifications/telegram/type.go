package telegram

import (
	"github.com/go-telegram/bot"
	"go.uber.org/zap"
)

type service struct {
	bot    *bot.Bot
	logger *zap.SugaredLogger
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
