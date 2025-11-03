package alerts

import (
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AlertRepository interface {
	AddAlert(*Alert) (int64, error)
	GetAlertById(id string) (*Alert, error)
	AlertsBySeverity() ([]*AlertsBySeverity, error)
	GetAlerts(string, string, int, int) ([]*Alert, error)
	GetUnsentAlerts() ([]Alert, error)
	MarkAlertAsSent(alertID string) error
	GetUnsentAlertID(alert Alert) (string, error)
}

type AlertsService interface {
	AddAlertManagerAlerts([]Alert) (int64, error)
	NewAlert(id, name, severity, description, status, method string,
		startsAt, endsAt time.Time,
		receptor []string) (Alert, error)
	GetFiringAlertsBySeverity() ([]*AlertsBySeverity, error)
	GetAlerts(string, string, int, int) ([]*Alert, error)
}

type alertsService struct {
	log *zap.SugaredLogger
	ar  AlertRepository
}

func NewAlertService(l *zap.SugaredLogger, ai AlertRepository) AlertsService {
	return &alertsService{l, ai}
}

func (as *alertsService) NewAlert(id, name, severity, description, status, method string,
	startsAt, endsAt time.Time,
	receptor []string,
) (Alert, error) {
	var als Alert
	checkAlert, err := as.ar.GetAlertById(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		als.Id = id
		als.Name = name
		als.Severity = severity
		als.Description = description
		als.Status = status
		als.Method = method
		als.StartsAt = startsAt
		als.EndsAt = endsAt
		als.Receptor = receptor
		als.SendNotif = false
		als.Silenced = false
		als.CreatedAt = time.Now()
		als.UpdatedAt = time.Now()
		return als, nil
	} else if err != nil {
		as.log.Error("error checking alert in database", err)
		return als, err
	}
	als.Id = checkAlert.Id
	als.Name = checkAlert.Name
	als.Severity = checkAlert.Severity
	als.Description = checkAlert.Description
	als.Status = status
	als.Method = checkAlert.Method
	als.StartsAt = checkAlert.StartsAt
	als.EndsAt = endsAt
	als.Receptor = checkAlert.Receptor
	als.Silenced = checkAlert.Silenced
	als.CreatedAt = checkAlert.CreatedAt
	als.UpdatedAt = time.Now()
	if status != checkAlert.Status {
		als.SendNotif = false
	} else {
		als.SendNotif = checkAlert.SendNotif
	}
	return als, nil
}

func (as *alertsService) AddAlertManagerAlerts(alerts []Alert) (int64, error) {
	var num int64 = 0
	as.log.Infow("we are going to save alerts", "alerts", alerts)
	for _, alert := range alerts {

		as.log.Info("we are going to save alerts")

		r, err := as.ar.AddAlert(&alert)
		if err != nil {
			as.log.Error("Error in Saving to the Database ", err)
			return num, err
		} else {
			as.log.Info(alert.Name, "is saved to the db")
			num += r
		}
	}
	return num, nil
}

func (as *alertsService) GetAlerts(status string, severity string, limit int, page int) ([]*Alert, error) {
	var als []*Alert
	als, err := as.ar.GetAlerts(status, severity, limit, page)
	if err != nil {
		as.log.Error("Error getting alerts from db ")
		return nil, err
	}
	return als, nil
}

func (as *alertsService) GetFiringAlertsBySeverity() ([]*AlertsBySeverity, error) {
	var als []*AlertsBySeverity
	als, err := as.ar.AlertsBySeverity()
	if err != nil {
		as.log.Error("an error getting from database")
		return als, err
	}
	return als, nil
}
