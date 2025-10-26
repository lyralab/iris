package postgresql

import (
	"context"
	"errors"
	"time"

	"github.com/root-ali/iris/pkg/alerts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (s *Storage) GetUnsentAlerts() ([]alerts.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	als := make([]alerts.Alert, 0)

	result := s.db.Table("alerts").
		Where("send_notif = ?", false).
		WithContext(ctx).
		Find(&als)

	if result.Error != nil {
		s.logger.Error("Error fetching unsent alerts from database:", result.Error)
		return nil, result.Error
	}

	s.logger.Infof("Fetched %d unsent alerts", len(als))
	results := make([]alerts.Alert, len(als))
	copy(results, als)

	return results, nil
}

func (s *Storage) MarkAlertAsSent(alertID string) error {
	// Start a new transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Lock the row for update using FOR UPDATE
	if err := tx.Model(&alerts.Alert{}).
		Where("id = ?", alertID).
		Clauses(clause.Locking{Strength: "UPDATE",
			Options: "SKIP LOCKED"}). // Lock the row for update
		Update("send_notif", true).Error; err != nil {
		// Log the error and rollback transaction in case of failure
		s.logger.Errorf("failed to update send_notif for alert %s: %v", alertID, err)
		tx.Rollback()
		return err
	}

	// Commit the transaction to release the lock
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetUnsentAlertID(alert alerts.Alert) (string, error) {
	// Start a new transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return "", tx.Error
	}

	// Lock the row using `FOR UPDATE`
	if err := tx.Where("send_notif = ?", false).Clauses(clause.Locking{Strength: "UPDATE",
		Options: "SKIP LOCKED"}).
		First(&alert).Error; err != nil {
		// If no record is found, return nil
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		// Log and return the error if any occurs during the fetching process
		s.logger.Errorf("failed to fetch unsent alert ID: %v", err)
		// Rollback transaction in case of an error
		tx.Rollback()
		return "", err
	}

	// Commit the transaction to release the lock
	if err := tx.Commit().Error; err != nil {
		return "", err
	}

	// Return the ID of the alert
	return alert.Id, nil
}
