package alerts

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Alert struct {
	Id           string         `json:"id" gorm:"column:id"`
	FingerPrint  string         `json:"fingerprint" gorm:"column:fingerprint"`
	Name         string         `json:"name" gorm:"column:name"`
	Severity     string         `json:"severity" gorm:"column:severity"`
	Description  string         `json:"description" gorm:"column:description"`
	StartsAt     time.Time      `json:"starts_at" gorm:"column:starts_at"`
	EndsAt       time.Time      `json:"ends_at" gorm:"column:ends_at"`
	Status       string         `json:"status" gorm:"column:status"`
	Labels       StringMap      `json:"labels" gorm:"column:labels;type:jsonb"`
	Annotations  StringMap      `json:"annotations" gorm:"column:annotations;type:jsonb"`
	GeneratorURL string         `json:"generator_url" gorm:"column:generator_url"`
	Method       pq.StringArray `json:"-" gorm:"column:method;type:text[]"`
	Receptor     pq.StringArray `json:"-" gorm:"column:receptor;type:text[]"`
	SendNotif    bool           `json:"-" gorm:"column:send_notif;default:false"`
	Silenced     bool           `json:"-" gorm:"column:silenced;default:false"`
	CreatedAt    time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"column:updated_at"`
	gorm.DeletedAt
}

type AlertsBySeverity struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}

// Details carries the Alertmanager fields kept as-is for display.
type Details struct {
	Labels       StringMap
	Annotations  StringMap
	GeneratorURL string
}

// StringMap is a string map stored as a JSONB column.
type StringMap map[string]string

func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func (m *StringMap) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*m = nil
		return nil
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return fmt.Errorf("StringMap: cannot scan %T", src)
	}
}
