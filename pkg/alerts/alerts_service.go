package alerts

import (
	"github.com/root-ali/iris/pkg/messages/alertmanager"
	"go.uber.org/zap"
)

type AlertRepository interface {
	AddAlert(*Alert) (int64, error)
	//GetNumberofFiringAlerts() (int64, error)
	//GetLatestResolvedAlerts(int) ([]*Alert, error)
	//GetLatestFiringAlerts(int) ([]*Alert, error)
	AlertsBySeverity() ([]*AlertsBySeverity, error)
	GetAlerts(string, string, int, int) ([]*Alert, error)
	GetUnsentAlerts() ([]Alert, error)
	MarkAlertAsSent(alertID string) error
	GetUnsentAlertID(alert Alert) (string, error)
}

type AlertsService interface {
	AddAlertManagerAlerts([]alertmanager.Alert) (int64, error)
	//GetLatestFiringAlerts() ([]*Alert, error)
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

func (as *alertsService) AddAlertManagerAlerts(alerts []alertmanager.Alert) (int64, error) {
	var num int64 = 0
	as.log.Infow("we are going to save alerts", "alerts", alerts)
	for _, alert := range alerts {
		var al Alert
		as.log.Info("we are going to save alerts")
		al.convertAlertMangerAlerts(alert)
		r, err := as.ar.AddAlert(&al)
		if err != nil {
			as.log.Error("Error in Saving to the Database ", err)
			return num, err
		} else {
			as.log.Info(al.Name, "is saved to the db")
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
