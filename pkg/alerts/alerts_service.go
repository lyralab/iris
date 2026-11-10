package alerts

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/metrics"
	"go.uber.org/zap"
)

type AlertRepository interface {
	AddAlert(*Alert) (int64, error)
	GetAlertById(id string) (*Alert, error)
	AlertsBySeverity() ([]*AlertsBySeverity, error)
	// GetAlerts lists alerts; silencedOnly keeps those whose silence is in effect.
	GetAlerts(status, severity string, silencedOnly bool, limit, page int) ([]*Alert, error)
	GetUnsentAlerts() ([]Alert, error)
	MarkAlertAsSent(alertID string) error
	GetUnsentAlertID(alert Alert) (string, error)
	GetAlertByFingerPrintAndStatus(fingerPrint, status string) (*Alert, error)
	SetAlertSilence(id string, silenced bool, until *time.Time, by string) error
}

// MaxSilenceDuration caps a timed silence.
const MaxSilenceDuration = 30 * 24 * time.Hour

type Service interface {
	AddAlertManagerAlerts([]Alert) (int64, error)
	NewAlert(id, name, severity, description, status string, method []string,
		startsAt, endsAt time.Time,
		receptor []string, details Details) (Alert, error)
	GetFiringAlertsBySeverity() ([]*AlertsBySeverity, error)
	// GetAlerts lists alerts; silencedOnly keeps those whose silence is in effect.
	GetAlerts(status, severity string, silencedOnly bool, limit, page int) ([]*Alert, error)
	// SilenceAlert suppresses notifications of a firing alert for every user.
	// A zero duration silences it until UnsilenceAlert.
	SilenceAlert(id, by string, duration time.Duration) (*Alert, error)
	UnsilenceAlert(id string) (*Alert, error)
}

type alertsService struct {
	log *zap.SugaredLogger
	ar  AlertRepository
}

func NewAlertService(l *zap.SugaredLogger, ai AlertRepository) Service {
	return &alertsService{l, ai}
}

func (as *alertsService) NewAlert(fingerprint, name, severity, description, status string,
	method []string,
	startsAt, endsAt time.Time,
	receptor []string,
	details Details,
) (Alert, error) {
	var als Alert
	checkAlert, err := as.ar.GetAlertByFingerPrintAndStatus(fingerprint, "firing")
	if apperr.Is(err, apperr.ErrAlertNotFound) {
		als.Id = uuid.New().String()
		als.FingerPrint = fingerprint
		als.Name = name
		als.Severity = strings.ToLower(strings.TrimSpace(severity))
		als.Description = description
		als.Status = status
		als.Method = method
		als.StartsAt = startsAt
		als.EndsAt = endsAt
		als.Receptor = receptor
		als.Labels = details.Labels
		als.Annotations = details.Annotations
		als.GeneratorURL = details.GeneratorURL
		als.SendNotif = false
		als.Silenced = false
		als.CreatedAt = time.Now()
		als.UpdatedAt = time.Now()
		return als, nil
	} else if err != nil {
		as.log.Errorw("error checking alert in database", "error", err)
		return als, err
	}
	als.Id = checkAlert.Id
	als.FingerPrint = checkAlert.FingerPrint
	als.Name = checkAlert.Name
	als.Severity = checkAlert.Severity
	als.Description = checkAlert.Description
	als.Status = status
	als.Method = checkAlert.Method
	als.StartsAt = checkAlert.StartsAt
	als.EndsAt = endsAt
	als.Receptor = checkAlert.Receptor
	// Annotations can change between notifications; keep the latest.
	als.Labels = details.Labels
	als.Annotations = details.Annotations
	als.GeneratorURL = details.GeneratorURL
	als.Silenced = checkAlert.Silenced
	als.SilencedUntil = checkAlert.SilencedUntil
	als.SilencedBy = checkAlert.SilencedBy
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
			if r > 0 {
				metrics.AlertReceived(alert.Status, alert.Severity)
			}
			num += r
		}
	}
	return num, nil
}

func (as *alertsService) GetAlerts(status, severity string, silencedOnly bool, limit, page int) ([]*Alert, error) {
	var als []*Alert
	als, err := as.ar.GetAlerts(status, severity, silencedOnly, limit, page)
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

func (as *alertsService) SilenceAlert(id, by string, duration time.Duration) (*Alert, error) {
	if duration < 0 || duration > MaxSilenceDuration {
		return nil, apperr.ErrInvalidSilenceDuration
	}
	al, err := as.ar.GetAlertById(id)
	if err != nil {
		return nil, err
	}
	if al.Status != "firing" {
		return nil, apperr.ErrAlertNotFiring
	}
	var until *time.Time
	if duration > 0 {
		t := time.Now().Add(duration)
		until = &t
	}
	if err := as.ar.SetAlertSilence(id, true, until, by); err != nil {
		return nil, err
	}
	al.Silenced, al.SilencedUntil, al.SilencedBy = true, until, by
	as.log.Infow("Alert silenced", "alertID", id, "name", al.Name, "by", by, "until", until)
	return al, nil
}

func (as *alertsService) UnsilenceAlert(id string) (*Alert, error) {
	al, err := as.ar.GetAlertById(id)
	if err != nil {
		return nil, err
	}
	if err := as.ar.SetAlertSilence(id, false, nil, ""); err != nil {
		return nil, err
	}
	as.log.Infow("Alert unsilenced", "alertID", id, "name", al.Name, "previouslyBy", al.SilencedBy)
	al.Silenced, al.SilencedUntil, al.SilencedBy = false, nil, ""
	return al, nil
}
