package postgresql

import (
	"errors"
	iris_error "github.com/root-ali/iris/pkg/errors"
	"github.com/root-ali/iris/pkg/roles"
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
		if errors.Is(result.Error, errors.New("record not found")) {
			s.logger.Info("cannot find a role with name ", r.Name)
			return iris_error.ErrRoleNotFound
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
		s.logger.Error("cannot get role from database",
			"error", result.Error)
		return result.Error
	}
	s.logger.Infow("successfully get rolename from database",
		"rolename", role.Name)
	return nil
}
