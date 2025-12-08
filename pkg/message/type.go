package message

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type Repository interface {
	SaveMessage(msg *Message) error
	GetMessageByID(id int) (*Message, error)
	UpdateMessage(msg *Message) error
	ListMessages() ([]*Message, error)
	ListNotFinishedMessages() ([]Message, error)
}

type Service struct {
	repo   Repository
	logger *zap.SugaredLogger
}

type Message struct {
	Id string

	UserId    string
	GroupName string
	Message   string
	Receptor  string

	SenderId string
	Sender   string
	Status   string

	Attempt       int
	LastAttempt   time.Time
	LastProviders pq.StringArray `gorm:"type:text[]"`
	Response      string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type StatusType int

const (
	TypeMessageStatusSent      StatusType = 1
	TypeMessageStatusFailed    StatusType = 6
	TypeMessageStatusDelivered StatusType = 10
)

var StatusMap = map[StatusType]string{
	TypeMessageStatusSent:      "Sent",
	TypeMessageStatusFailed:    "Failed",
	TypeMessageStatusDelivered: "Delivered",
}

func NewMessage(senderId, message, receptor, sender, userId, groupName, response string,
	providerChain []string,
	messageStatus StatusType) *Message {

	lastProvider := pq.StringArray(providerChain)
	return &Message{
		Id:            uuid.New().String(),
		UserId:        userId,
		GroupName:     groupName,
		SenderId:      senderId,
		Message:       message,
		Receptor:      receptor,
		Sender:        sender,
		Status:        StatusMap[messageStatus],
		Attempt:       0,
		LastAttempt:   time.Now(),
		LastProviders: lastProvider,
		Response:      response,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
