package telegram

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

func NewTelegramService(token string, proxy string, logger *zap.SugaredLogger) (notifications.NotificationInterface, error) {
	var bopts []bot.Option
	if proxy != "" {
		logger.Warn("Running telegram bot with proxy")
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			logger.Errorw("Cannot parse proxy url", "error", err)
			return nil, err
		}
		bopts = append(bopts, bot.WithHTTPClient(10*time.Second, &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			},
		}))
	}
	bopts = append(bopts, bot.WithDefaultHandler(nil))
	b, err := bot.New(token, bopts...)
	if err != nil {
		return nil, err
	}
	return &service{bot: b, logger: logger}, nil
}

func (s *service) GetName() string {
	return "Telegram"
}

func (s *service) GetFlag() string {
	return "telegram"
}

func (s *service) GetPriority() int {
	return 3
}

func (s *service) Verify() (string, error) {
	return "success", nil
}

func (s *service) Send(message notifications.Message) ([]string, error) {
	var errStack errorStack
	ctx := context.Background()
	responses := make([]string, 0)
	text := ""
	if message.State == "firing" {
		text = `🚨<b> Firing </b>🚨` + "\n\n<b>" + message.Subject + "</b>\n\n" + message.Message + "\n\n" + message.Time
	} else if message.State == "resolved" {
		text = `✅<b> Resolved </b>✅` + "\n\n<b>" + message.Subject + "</b>\n\n" + message.Message + "\n\n" + message.Time
	}

	for _, receptor := range message.Receptors {
		chatID, err := strconv.ParseInt(receptor, 10, 64)
		if err != nil {
			s.logger.Errorw("Cannot parse chat id", "receptor", receptor, "error", err)
			errStack.Append(err)
			continue
		}

		resp, err := s.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		})

		if err != nil {
			s.logger.Errorw("Error sending telegram message", "chatID", chatID, "error", err)
			errStack.Append(err)
			responses = append(responses, receptor)
			continue
		}
		s.logger.Infow("Telegram message sent",
			"chatID", chatID, "messageID", resp.ID)
		errStack = append(errStack, nil)
		responses = append(responses, strconv.Itoa(resp.ID))
	}
	if len(errStack) > 0 {
		return responses, errStack
	}
	return responses, nil
}

func (s *service) Status(_ string) (notifications.MessageStatusType, error) {
	return notifications.TypeMessageStatusFailed, nil
}
