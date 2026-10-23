package postgresql

import (
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/roles"
	"gorm.io/gorm"
)

func (s *Storage) AddRole(r roles.Role) error {
	result := s.db.Create(&r)
	if result.Error != nil {
		s.logger.Infow("cannot insert record", "error", result.Error)
		return result.Error
	}
	s.logger.Infow("Successfully insert role", "role-name", r.Name)
	return nil
}

func (s *Storage) GetRoleByName(r *roles.Role) error {

	result := s.db.First(r, "name = ?", r.Name)
	if result.Error != nil {
		if apperr.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Infow("cannot find a role", "name", r.Name)
			return apperr.ErrRoleNotFound
		} else {
			s.logger.Infow("Cannot get role from database",
				"name", r.Name,
				"error", result.Error)
			return result.Error
		}
	}
	return nil
}

func (s *Storage) GetRoleName(role *roles.Role) error {
	result := s.db.Find(&role, "id = ?", role.ID)
	if result.Error != nil {
		s.logger.Errorw("cannot get role from database",
			"error", result.Error)
		return result.Error
	}
	s.logger.Infow("successfully get rolename from database",
		"rolename", role.Name)
	return nil
}
