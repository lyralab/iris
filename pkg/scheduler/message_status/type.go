package message_status

import (
	"context"
	"sync"
	"time"

	"github.com/root-ali/iris/pkg/message"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

type MessageListRepository interface {
	Add(msg *message.Message) error
	UpdateMessageStatus(msg *message.Message, status message.StatusType, response string) error
	ListNotFinishedMessages() ([]message.Message, error)
}

type ProviderRepository interface {
	GetActiveProviders() ([]notifications.Providers, error)
	GetProvidersPriority() ([]notifications.Providers, error)
}

type Service struct {
	// scheduler dependencies
	messageRepo  MessageListRepository
	providerRepo ProviderRepository

	//Config
	config Config

	// runtime
	ctx    context.Context
	cancel context.CancelFunc

	mu      sync.Mutex
	started bool

	wg sync.WaitGroup

	taskCh chan message.Message
	ticker *time.Ticker

	logger *zap.SugaredLogger
}

type Config struct {
	StartAt   time.Time
	Interval  time.Duration
	Workers   int
	QueueSize int
}

var (
	MaxAttempt = 3
)
