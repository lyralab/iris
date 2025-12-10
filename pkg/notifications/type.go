package notifications

import (
	"time"

	"github.com/root-ali/iris/pkg/cache"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotificationInterface interface {
	Send(message Message) ([]string, error)
	Status(messageID string) (MessageStatusType, error)
	Verify() (string, error)
	GetName() string
	GetFlag() string
	GetPriority() int
}

type RepositoryInterface interface {
	AddProvider(providers *Providers) error
	ModifyProvider(providers *Providers) error
	GetProvider(providers *Providers) error
	SetStatusFalse(providers *Providers) error
	GetProviders() ([]Providers, error)
}

// ProviderServiceInterface is an interface for managing notification providers.
type ProviderServiceInterface interface {
	AddProvider(providers *Providers) error
	EnableProvider(name string) error
	DisableProvider(name string) error
	ModifyProviderPriority(name string, priority int) error
	GetProviderByName(name string) (*Providers, error)
	GetProviderByID(id string) (*Providers, error)
	GetAllProviders() ([]Providers, error)
}

type ProviderStatusInterface interface {
	GetActiveProviders() ([]Providers, error)
	GetProvidersPriority() ([]Providers, error)
}

type Providers struct {
	ID             string                `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name           string                `gorm:"uniqueIndex;not null"`
	Description    string                `gorm:"not null"`
	Flag           string                `gorm:"not null"`
	Priority       int                   `gorm:"not null;"`
	Provider       NotificationInterface `gorm:"-"`
	Status         bool                  `gorm:"column:is_active;not null;default:true"`
	CreatedAt      time.Time             `gorm:"autoCreateTime"`
	ModifiedAt     time.Time             `gorm:"autoUpdateTime"`
	gorm.DeletedAt `gorm:"index;default:null" json:"-"`
}
type Message struct {
	Subject   string
	Message   string
	State     string
	Time      string
	Receptors []string
}

type MessageStatusType int

const (
	TypeMessageStatusSent        MessageStatusType = 1
	TypeMessageStatusFailed      MessageStatusType = 6
	TypeMessageStatusDelivered   MessageStatusType = 10
	TypeMessageStatusUndelivered MessageStatusType = 11
)

type ProviderService struct {
	repo   RepositoryInterface
	cache  cache.Interface[string, *[]Providers]
	ns     []NotificationInterface
	Logger *zap.SugaredLogger
}
