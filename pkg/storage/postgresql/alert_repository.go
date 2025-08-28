package postgresql

import (
	"errors"
	"github.com/root-ali/iris/pkg/alerts"
	"gorm.io/gorm"
)

func (s *Storage) AddAlert(alert *alerts.Alert) (int64, error) {
	result := s.db.Save(alert)
	if result.Error != nil {
		s.logger.Error("Database Error is: ", result.Error)
		return 0, result.Error
	}
	s.logger.Info(alert.Id, alert.Name, " saved to the db")
	return result.RowsAffected, nil
}

func (s *Storage) FiringAlertsBySeverity() (int64, error) {
	var count int64
	err := s.db.Model(&alerts.Alert{}).Where("status = ?", "firing").Group("severity").Count(&count).Error
	if err != nil {
		s.logger.Error("error in getting result from database")
		return count, err
	}
	return count, nil
}

func (s *Storage) GetLatestFiringAlerts(l int) ([]*alerts.Alert, error) {
	var as []*alerts.Alert
	err := s.db.Where("status = ?", "firing").Where("severity = ?", "critical").Limit(l).Find(as).Error
	if err != nil {
		s.logger.Error("error getting as from database", err)
		return nil, err
	}
	return as, nil
}

func (s *Storage) GetLatestResolvedAlerts(l int) ([]*alerts.Alert, error) {
	var as []*alerts.Alert
	err := s.db.Where("status = ?", "resolved").Limit(l).Find(as).Error
	if err != nil {
		s.logger.Error("error getting as from database")
		return nil, err
	}
	return as, err
}

func (s *Storage) GetAlerts(status string, severity string, l int, p int) ([]*alerts.Alert, error) {
	s.logger.Info("status", status, "limit", l, "page ", p)
	var statusQuery string
	var severityQuery string
	var as []*alerts.Alert
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
	s.logger.Info("query", statusQuery, "limit", l, "offset", offset)
	err := s.db.Where("status LIKE ?", statusQuery).Where("severity LIKE ?", severityQuery).Limit(l).Offset(offset).Find(&as).Error
	if err != nil {
		s.logger.Error("Error getting as from database")
		return as, err
	}
	return as, nil
}

func (s *Storage) AlertsBySeverity() ([]*alerts.AlertsBySeverity, error) {
	var als []*alerts.AlertsBySeverity
	result := s.db.Table("alerts").
		Select("severity,count(severity) as count").
		Group("severity").
		Scan(&als)
	if result.Error != nil {
		s.logger.Error("Error getting query from database ,", result.Error)
		return nil, result.Error
	}
	return als, nil
}

func (s *Storage) Health() error {
	err := s.db.Raw("select 1").Error
	if err != nil {
		s.logger.Error("Error in getting connection from database", err)
		return err
	}
	return nil
}

func (s *Storage) GetUnsentAlerts() ([]alerts.Alert, error) {
	var results []alerts.Alert
	if err := s.db.
		Where("send_notif = ?", false).
		Find(&results).Error; err != nil {
		s.logger.Errorf("failed to fetch unsent alerts: %v", err)
		return nil, err
	}
	return results, nil
}

func (s *Storage) MarkAlertAsSent(alertID string) error {
	if err := s.db.Model(&alerts.Alert{}).
		Where("id = ?", alertID).
		Update("send_notif", true).Error; err != nil {
		s.logger.Errorf("failed to update send_notif for alert %s: %v", alertID, err)
		return err
	}
	return nil
}

func (s *Storage) GetUnsentAlertID(alert alerts.Alert) (string, error) {
	if err := s.db.
		Where("send_notif = ?", false).
		First(&alert).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		s.logger.Errorf("failed to fetch unsent alert ID: %v", err)
		return "", err
	}
	return alert.Id, nil
}
