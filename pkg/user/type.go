package user

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type User struct {
	ID             string    `gorm:"column:id;primary_key"`
	UserName       string    `gorm:"column:user_name;unique"`
	FirstName      string    `gorm:"column:first_name;not null"`
	LastName       string    `gorm:"column:last_name;not null"`
	Password       string    `gorm:"column:password;not null"`
	Salt           string    `gorm:"column:salt;not null"`
	Email          string    `gorm:"column:email;unique"`
	Status         string    `gorm:"column:status;not null"`
	Mobile         string    `gorm:"column:mobile;not null"`
	Role           string    `gorm:"column:role_id"`
	CreatedAt      time.Time `gorm:"created_at"`
	ModifiedAt     time.Time `gorm:"modified_at"`
	gorm.DeletedAt `gorm:"deleted_at"`
}

type userServiceImpl struct {
	repo   UserInterfaceRepository
	role   UserRoleRepository
	logger *zap.SugaredLogger
}

type UserRoleRepository interface {
	GetRoleIDByName(name string) (string, error)
}

type UserInterfaceRepository interface {
	AddUser(user *User) error
	GetUserByID(id string) (*User, error)
	GetUserByUsername(user *User) error
	GetUserByEmail(email string) (*User, error)
	GetAllUsers() ([]*User, error)
	DeleteUser(id string) error
	UpdateUserPassword(id string, newPassword string) error
	GetRole(u *User) error
	VerifyUser(u *User) error
	UpdateUserData(u *User) error
}

type UserInterfaceService interface {
	AddUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	CreateDefaultAdminUser() error
	ValidateUser(*User) error
	GetUserRole(*User) error
	VerifyUser(*User) error
	UpdateUser(*User) error
	GetAllUsers() ([]*User, error)
}
