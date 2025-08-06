package roles

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type Role struct {
	ID             string    `gorm:"column:id;not null"`
	Name           string    `gorm:"column:name;not null"`
	Access         string    `gorm:"column:access;not null"`
	Created_at     time.Time `gorm:"column:created_at;not null"`
	Modified_at    time.Time `gorm:"column:modified_at;not null"`
	gorm.DeletedAt `gorm:"column:deleted_at"`
}

type roleServiceImpl struct {
	rr     RolesInterfaceRepository
	logger *zap.SugaredLogger
}
