package notifications

type NotificationInterface interface {
	Send(message Message) ([]string, error)
	Status(messageID string) (MessageStatusType, error)
}

type Message struct {
	Subject   string
	Message   string
	Receptors []string
}

type MessageStatusType int

const (
	TypeMessageStatusSent      MessageStatusType = 1
	TypeMessageStatusFailed    MessageStatusType = 0
	TypeMessageStatusDelivered MessageStatusType = 10
)

var MessageStatusMap = map[MessageStatusType]string{
	TypeMessageStatusSent:      "Sent",
	TypeMessageStatusFailed:    "Failed",
	TypeMessageStatusDelivered: "Delivered",
}
