package alerts

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Alert struct {
	Id          string         `json:"id" gorm:"column:id"`
	Name        string         `json:"name" gorm:"column:name"`
	Severity    string         `json:"severity" gorm:"column:severity"`
	Description string         `json:"description" gorm:"column:description"`
	StartsAt    time.Time      `json:"starts_at" gorm:"column:starts_at"`
	EndsAt      time.Time      `json:"ends_at" gorm:"column:ends_at"`
	Status      string         `json:"status" gorm:"column:status"`
	Method      string         `json:"method" gorm:"column:method"`
	Receptor    pq.StringArray `json:"receptor" gorm:"column:receptor;type:text[]"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at"`
	gorm.DeletedAt
}

type AlertsBySeverity struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}
