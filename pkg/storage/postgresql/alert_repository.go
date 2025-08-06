package postgresql

import (
	"fmt"
	"github.com/root-ali/iris/pkg/alerts"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

func New(l *zap.SugaredLogger, p *Postgres) *Storage {

	storage := Storage{}
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", p.Host, p.User, p.Password, p.DBname, p.Port)

	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		l.Panic("Failed to connect to the Database")
	}
	l.Info("? Connected Successfully to the Database")
	storage.db = DB
	storage.log = l
	return &storage

}

func (s *Storage) AddAlert(alert *alerts.Alert) (int64, error) {
	result := s.db.Save(alert)
	if result.Error != nil {
		s.log.Error("Database Error is: ", result.Error)
		return 0, result.Error
	}
	s.log.Info(alert.Id, alert.Name, " saved to the db")
	return result.RowsAffected, nil
}

func (s *Storage) FiringAlertsBySeverity() (int64, error) {
	var count int64
	err := s.db.Model(&alerts.Alert{}).Where("status = ?", "firing").Group("severity").Count(&count).Error
	if err != nil {
		s.log.Error("error in getting result from database")
		return count, err
	}
	return count, nil
}

func (s *Storage) GetLatestFiringAlerts(l int) ([]*alerts.Alert, error) {
	var alerts []*alerts.Alert
	err := s.db.Where("status = ?", "firing").Where("severity = ?", "critical").Limit(l).Find(alerts).Error
	if err != nil {
		s.log.Error("error getting alerts from database", err)
		return nil, err
	}
	return alerts, nil
}

func (s *Storage) GetLatestResolvedAlerts(l int) ([]*alerts.Alert, error) {
	var alerts []*alerts.Alert
	err := s.db.Where("status = ?", "resolved").Limit(l).Find(alerts).Error
	if err != nil {
		s.log.Error("error getting alerts from database")
		return nil, err
	}
	return alerts, err
}

func (s *Storage) GetAlerts(status string, severity string, l int, p int) ([]*alerts.Alert, error) {
	s.log.Info("status", status, "limit", l, "page ", p)
	var statusQuery string
	var severityQuery string
	var alerts []*alerts.Alert
	if status == "" {
		statusQuery = "%%"
	} else {
		statusQuery = "%" + status + "%"
	}
	if severity == "" {
		severityQuery = "%%"
	} else {
		severityQuery = "%" + severity + "%"
	}
	var offset int
	if l > 0 && p > 1 {
		offset = p * l
	} else {
		offset = -1
	}
	if l <= 0 {
		l = -1
	}
	s.log.Info("query", statusQuery, "limit", l, "offset", offset)
	err := s.db.Where("status LIKE ?", statusQuery).Where("severity LIKE ?", severityQuery).Limit(l).Offset(offset).Find(&alerts).Error
	if err != nil {
		s.log.Error("Error getting alerts from database")
		return alerts, err
	}
	return alerts, nil
}

func (s *Storage) AlertsBySeverity() ([]*alerts.AlertsBySeverity, error) {
	var als []*alerts.AlertsBySeverity
	result := s.db.Table("alerts").
		Select("severity,count(severity) as count").
		Group("severity").
		Scan(&als)
	if result.Error != nil {
		s.log.Error("Error getting query from database ,", result.Error)
		return nil, result.Error
	}
	return als, nil
}

func (s *Storage) Health() error {
	err := s.db.Raw("select 1").Error
	if err != nil {
		s.log.Error("Error in getting connection from database", err)
		return err
	}
	return nil
}
