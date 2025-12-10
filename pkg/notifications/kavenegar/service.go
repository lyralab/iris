package kavenegar

import (
	"strconv"

	kn "github.com/kavenegar/kavenegar-go"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

func NewKavenegarService(apiToken string, p int, sender string, logger *zap.SugaredLogger) *KavenegarService {
	api := kn.New(apiToken)
	return &KavenegarService{
		API:      api,
		Sender:   sender,
		Priority: p,
		Logger:   logger,
	}
}

func (k *KavenegarService) Send(message notifications.Message) ([]string, error) {
	return k.kavenegarSend(k.Sender, message)
}

func (k *KavenegarService) Status(messageID string) (notifications.MessageStatusType, error) {
	messageIDs := []string{messageID}
	var messageStatus notifications.MessageStatusType

	status, err := k.API.Message.Status(messageIDs)
	if err != nil {
		return messageStatus, nil
	}
	k.Logger.Infow("Kavenegar message status response", "status", status, "messageID", messageID)
	for _, s := range status {
		if s.Status == 10 {
			messageStatus = notifications.TypeMessageStatusDelivered
		} else if s.Status == 11 {
			messageStatus = notifications.TypeMessageStatusUndelivered
		} else if s.Status == 6 {
			messageStatus = notifications.TypeMessageStatusFailed
		} else if s.Status == 4 || s.Status == 5 {
			messageStatus = notifications.TypeMessageStatusSent
		}
	}
	k.Logger.Infow("Kavenegar mapped message status",
		"mappedStatus", messageStatus,
		"messageID", messageID)
	return messageStatus, nil
}

func (k *KavenegarService) kavenegarSend(_ string, messages notifications.Message) ([]string, error) {
	text := ""
	if messages.State == "firing" {
		text = "🚨" + messages.Subject + "\n" + messages.Message + "\nTime: " + messages.Time
	} else if messages.State == "resolved" {
		text = "✅" + messages.Subject + "\n" + messages.Message + "\nTime: " + messages.Time
	}
	resp, err := k.API.Message.Send("", messages.Receptors, text, nil)
	if err != nil {
		return nil, err
	}
	var messageIDs []string
	for _, r := range resp {
		messageIDs = append(messageIDs, strconv.Itoa(int(r.MessageID)))
	}
	return messageIDs, nil
}

func (k *KavenegarService) Verify() (string, error) {
	accInfo, err := k.API.Account.Info()
	if err != nil {
		k.Logger.Errorw("Failed to get account info", "error", err)
		return "", err
	}
	k.Logger.Infow("Kavenegar account info", "RemainAccount", accInfo.Remaincredit)
	return strconv.Itoa(accInfo.Remaincredit), nil
}

func (k *KavenegarService) GetName() string {
	return "Kavenegar"
}

func (k *KavenegarService) GetFlag() string {
	return "sms"
}

func (k *KavenegarService) GetPriority() int {
	return k.Priority
}
