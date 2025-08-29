package notifications

import (
	"go.uber.org/zap"
)

func NewProvidersService(repo ProviderRepositryInterface, logger *zap.SugaredLogger) ProviderServiceInterface {
	return &providerService{
		repo:   repo,
		Logger: logger,
	}
}

func (p *providerService) AddProvider(providers *Providers) error {
	return p.repo.AddProvider(providers)
}

func (p *providerService) EnableProvider(name string) error {
	providers := &Providers{Name: name, Status: true}
	p.Logger.Info("Enabling provider: ", providers.Name)
	err := p.repo.ModifyProvider(providers)
	if err != nil {
		p.Logger.Error("Error enabling provider: ", err)
		return err
	}
	return nil
}

func (p *providerService) DisableProvider(name string) error {
	provider := &Providers{Name: name, Status: false}
	p.Logger.Info("Disabling provider: ", provider.Name)

	err := p.repo.SetStatusFalse(provider)
	if err != nil {
		p.Logger.Error("Error disabling provider: ", err)
		return err
	}
	return nil
}

func (p *providerService) ModifyProviderPriority(name string, priority int) error {
	provider := &Providers{Name: name, Priority: priority}
	p.Logger.Infow("Change provider priority: ", "name", provider.Name,
		"priority", provider.Priority)

	if err := p.repo.ModifyProvider(provider); err != nil {
		p.Logger.Error("Error changing provider priority: ", err)
		return err
	}
	return nil
}

func (p *providerService) GetProviderByName(name string) (*Providers, error) {
	provider := &Providers{Name: name}
	if err := p.repo.GetProvider(provider); err != nil {
		p.Logger.Errorw("Error getting provider by name: ", "error", err)
		return nil, err
	}
	return provider, nil
}

func (p *providerService) GetProviderByID(id string) (*Providers, error) {
	provider := &Providers{ID: id}
	if err := p.repo.GetProvider(provider); err != nil {
		p.Logger.Errorw("Error getting provider by ID: ", "error", err)
		return nil, err
	}
	return provider, nil
}

func (p *providerService) GetAllProviders() ([]Providers, error) {
	providers, err := p.repo.GetProviders()
	if err != nil {
		p.Logger.Errorw("Error getting all providers: ", "error", err)
		return nil, err
	}
	return providers, nil
}
