package kavenegar

import (
	kn "github.com/kavenegar/kavenegar-go"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
	"strconv"
)

func NewKavenegarService(apiToken string, sender string, logger *zap.SugaredLogger) notifications.NotificationInterface {
	api := kn.New(apiToken)
	return &kavenegarService{
		API:    api,
		Sender: sender,
		Logger: logger,
	}
}

func (k *kavenegarService) Send(message notifications.Message) ([]string, error) {
	return k.kavenegarSend(k.Sender, message)
}

func (k *kavenegarService) Status(messageID string) (notifications.MessageStatusType, error) {
	messageIDs := []string{messageID}
	var messageStatus notifications.MessageStatusType
	status, err := k.API.Message.Status(messageIDs)
	if err != nil {
		return messageStatus, nil
	}
	for _, s := range status {
		if s.Status == 10 {
			messageStatus = notifications.TypeMessageStatusDelivered
		}

	}
	return messageStatus, nil
}

func (k *kavenegarService) kavenegarSend(sender string, messages notifications.Message) ([]string, error) {
	resp, err := k.API.Message.Send("", messages.Receptors, messages.Message, nil)
	if err != nil {
		return nil, err
	}
	var messageIDs []string
	for _, r := range resp {
		messageIDs = append(messageIDs, strconv.Itoa(int(r.MessageID)))
	}
	return messageIDs, nil
}
