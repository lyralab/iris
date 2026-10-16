package groups

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type GroupRepositoryInterface interface {
	AddGroup(*Group) error
	GetGroupById(string) (*Group, error)
	GetGroupByName(string) (*Group, error)
	GetAllGroups() ([]*Group, error)
	DeleteGroup(*Group) error
	AddUserToGroup(string, string) error
	RemoveUserFromGroup(string, string) error
	FindUsersById(string) ([]string, error)
	FindGroupById(string) ([]string, error)
}

type GroupServiceInterface interface {
	CreateGroup(*Group) error
	GetGroup(string) (*Group, error)
	GetGroupById(string) (*Group, error)
	GetAllGroups() ([]*Group, error)
	DeleteGroup(*Group) error
	RemoveUser(*Group, string) error
	AddUser(*Group, string) error
	ListUsers(*Group) ([]string, error)
	ListGroupByUser(string) ([]*Group, error)
}

// ReceptorCache misses and refills notification receptors after a group change.
type ReceptorCache interface {
	Refresh()
}

type GroupService struct {
	log   *zap.SugaredLogger
	gr    GroupRepositoryInterface
	cache ReceptorCache
}
type Group struct {
	ID          string    `json:"id,omitempty" gorm:"column:id;primary_key" binding:"-"`
	Name        string    `json:"name" gorm:"column:name;unique" binding:"required"`
	Description string    `json:"description" gorm:"column:description" binding:"-"`
	CreatedAt   time.Time `json:"-" gorm:"created_at" binding:"-"`
	ModifiedAt  time.Time `json:"-" gorm:"modified_at" binding:"-"`
	// Named, not embedded: an embedded DeletedAt promotes its MarshalJSON and the whole Group encodes as null.
	DeletedAt gorm.DeletedAt `json:"-"`
}

type UserGroup struct {
	UId        string    `json:"user_id" gorm:"column:user_id;primary_key" binding:"required"`
	GId        string    `json:"group_id" gorm:"column:group_id;primary_key" binding:"required"`
	CreatedAt  time.Time `json:"-" gorm:"created_at" binding:"-"`
	ModifiedAt time.Time `json:"-" gorm:"modified_at" binding:"-"`
	gorm.DeletedAt
}
