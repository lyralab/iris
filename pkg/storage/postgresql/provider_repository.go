package postgresql

import (
	"github.com/root-ali/iris/pkg/errors"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

func (s *Storage) AddProvider(p *notifications.Providers) error {
	if err := s.db.Create(p).Error; err != nil {
		zap.S().Errorf("AddProvider error: %v", err)
		return err
	}
	return nil
}

func (s *Storage) ModifyProvider(p *notifications.Providers) error {
	s.logger.Infow("Modifying provider", "name", p.Name, "priority", p.Priority, "status", p.Status)
	result := s.db.Model(&notifications.Providers{}).Where("name = ?", p.Name).Updates(p)
	if result.Error != nil {
		s.logger.Errorw("ModifyProvider error", "error", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.ErrProviderNotFound
	}
	return nil
}

func (s *Storage) SetStatusFalse(p *notifications.Providers) error {
	result := s.db.Model(&notifications.Providers{}).Where("name = ?", p.Name).Update("is_active", false)
	if result.Error != nil {
		s.logger.Errorf("error while disabling provider: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.ErrProviderNotFound
	}
	return nil
}

func (s *Storage) GetProvider(p *notifications.Providers) error {
	result := s.db.Where("id = ? OR name = ?", &p.ID, &p.Name).First(p)
	if result.Error != nil {
		if err := result.Error; err.Error() == "record not found" {
			return errors.ErrProviderNotFound
		}

		s.logger.Errorf("GetProvider error: %v", result.Error)
		return result.Error
	}
	return nil
}

func (s *Storage) GetProviders() ([]notifications.Providers, error) {
	var providers []notifications.Providers
	if err := s.db.Order("priority asc").Find(&providers).Error; err != nil {
		zap.S().Errorf("GetProviders error: %v", err)
		return nil, err
	}
	return providers, nil
}
