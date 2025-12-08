package telegram

import (
	"github.com/go-telegram/bot"
	"go.uber.org/zap"
)

type service struct {
	bot    *bot.Bot
	logger *zap.SugaredLogger
}
