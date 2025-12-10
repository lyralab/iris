package notifications

import (
	"time"

	"github.com/root-ali/iris/pkg/cache"
	"go.uber.org/zap"
)

func NewProvidersService(repo RepositoryInterface,
	np []NotificationInterface,
	c cache.Interface[string,
		*[]Providers],
	logger *zap.SugaredLogger) *ProviderService {
	return &ProviderService{
		repo:   repo,
		cache:  c,
		ns:     np,
		Logger: logger,
	}
}

func (p *ProviderService) AddProvider(providers *Providers) error {
	return p.repo.AddProvider(providers)
}

func (p *ProviderService) EnableProvider(name string) error {
	providers := &Providers{Name: name, Status: true}
	p.Logger.Info("Enabling provider: ", providers.Name)
	err := p.repo.ModifyProvider(providers)
	if err != nil {
		p.Logger.Error("Error enabling provider: ", err)
		return err
	}
	return nil
}

func (p *ProviderService) DisableProvider(name string) error {
	provider := &Providers{Name: name, Status: false}
	p.Logger.Info("Disabling provider: ", provider.Name)

	err := p.repo.SetStatusFalse(provider)
	if err != nil {
		p.Logger.Error("Error disabling provider: ", err)
		return err
	}
	return nil
}

func (p *ProviderService) ModifyProviderPriority(name string, priority int) error {
	provider := &Providers{Name: name, Priority: priority}
	p.Logger.Infow("Change provider priority: ", "name", provider.Name,
		"priority", provider.Priority)

	if err := p.repo.ModifyProvider(provider); err != nil {
		p.Logger.Error("Error changing provider priority: ", err)
		return err
	}
	return nil
}

func (p *ProviderService) GetProviderByName(name string) (*Providers, error) {
	provider := &Providers{Name: name}
	if err := p.repo.GetProvider(provider); err != nil {
		p.Logger.Errorw("Error getting provider by name: ", "error", err)
		return nil, err
	}
	return provider, nil
}

func (p *ProviderService) GetProviderByID(id string) (*Providers, error) {
	provider := &Providers{ID: id}
	if err := p.repo.GetProvider(provider); err != nil {
		p.Logger.Errorw("Error getting provider by ID: ", "error", err)
		return nil, err
	}
	return provider, nil
}

func (p *ProviderService) GetAllProviders() ([]Providers, error) {
	providers, err := p.repo.GetProviders()
	if err != nil {
		p.Logger.Errorw("Error getting all providers: ", "error", err)
		return nil, err
	}
	return providers, nil
}

func (p *ProviderService) GetActiveProviders() ([]Providers, error) {
	if cachedProviders, ok := p.cache.Get("active_providers"); ok {
		p.Logger.Info("Active providers fetched from cache")
		return *cachedProviders, nil
	}

	providers, err := p.repo.GetProviders()
	if err != nil {
		p.Logger.Errorw("Error getting active providers: ", "error", err)
		return nil, err
	}

	activeProviders := make([]Providers, 0)
	for _, provider := range providers {
		for _, np := range p.ns {
			if provider.Name == np.GetName() {
				provider.Provider = np
				break
			}
		}
		if provider.Status {
			activeProviders = append(activeProviders, provider)
		}
	}

	if err := p.cache.Set("active_providers", &activeProviders, 24*time.Hour); err != nil {
		p.Logger.Errorw("Error caching active providers: ", "error", err)
	}

	return activeProviders, nil
}

func (p *ProviderService) GetProvidersPriority() ([]Providers, error) {
	if cachedProviders, ok := p.cache.Get("providers_priority"); ok {
		p.Logger.Info("Providers by priority fetched from cache")
		return *cachedProviders, nil
	}

	providers, err := p.repo.GetProviders()
	if err != nil {
		p.Logger.Errorw("Error getting providers by priority: ", "error", err)
		return nil, err
	}

	sortedProviders := make([]Providers, len(providers))
	copy(sortedProviders, providers)

	// Sort providers by priority (lower number means higher priority)
	for i := 0; i < len(sortedProviders)-1; i++ {
		for j := i + 1; j < len(sortedProviders); j++ {
			if sortedProviders[i].Priority > sortedProviders[j].Priority {
				sortedProviders[i], sortedProviders[j] = sortedProviders[j], sortedProviders[i]
			}
		}
	}

	for i, provider := range sortedProviders {
		for _, np := range p.ns {
			if provider.Name == np.GetName() {
				p.Logger.Infow("Mapping provider interface: ", "provider", provider.Name)
				p.Logger.Info("provider ", np)
				sortedProviders[i].Provider = np
				break
			}
		}
	}

	if err := p.cache.Set("providers_priority", &sortedProviders, 24*time.Hour); err != nil {
		p.Logger.Errorw("Error caching providers by priority: ", "error", err)
	}

	return sortedProviders, nil
}
