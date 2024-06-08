package health_check

import "go.uber.org/zap"

type HealthRepostiry interface {
	Health() error
}

type HealthService interface {
	Healthy() (*Health, error)
	Ready() error
}

type healthService struct {
	log *zap.SugaredLogger
	hr  HealthRepostiry
}

func NewHealthService(l *zap.SugaredLogger, hr HealthRepostiry) HealthService {
	return &healthService{log: l, hr: hr}
}

func (hs *healthService) Healthy() (*Health, error) {
	err := hs.hr.Health()
	var chks []Checks
	var h Health
	if err != nil {
		h.Status = "Down"
		h.Checks = append(chks, Checks{
			Name:   "postgresql",
			Status: "cannot get connection from database",
		})
		return &h, err
	} else {
		h.Status = "UP"
		h.Checks = append(chks, Checks{
			Name:   "postgresql",
			Status: "database is UP",
		})
		return &h, nil
	}
}

func (hs *healthService) Ready() error {
	return nil
}
