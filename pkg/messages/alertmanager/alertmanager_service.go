package alertmanager

type AlertManagerRepository interface {
	AddAlertManagerAlerts(a *[]Alert) (int64, error)
}

type AlertManagerService interface {
}

type alertManagerService struct {
	amr *AlertManagerRepository
}

func NewAlertManagerService(a AlertManagerRepository) AlertManagerService {
	return &alertManagerService{&a}
}

func (a *alertManagerService) AlertManager() error {

	return nil
}
