package postgresql

import (
	"errors"
	iris_error "github.com/root-ali/iris/pkg/errors"
	"github.com/root-ali/iris/pkg/groups"
	"gorm.io/gorm"
)

func (s *Storage) AddGroup(g *groups.Group) error {
	result := s.db.Create(g)
	if result.Error != nil {
		return result.Error
	}
	s.logger.Infow("group is saved", "group", g.Name, "number of row affected", result.RowsAffected)
	return nil
}

func (s *Storage) GetGroupById(id string) (*groups.Group, error) {
	g := &groups.Group{ID: id}
	result := s.db.First(g)
	if result.Error != nil {
		return nil, result.Error
	}
	return g, nil
}

func (s *Storage) GetGroupByName(name string) (*groups.Group, error) {
	s.logger.Infow("Get group by name", "name", name)
	var g *groups.Group
	result := s.db.First(&g, "name = ?", name)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, iris_error.ErrGroupNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return g, nil
}

func (s *Storage) GetAllGroups() ([]*groups.Group, error) {
	var gs []*groups.Group
	result := s.db.Find(&gs)
	if result.Error != nil {
		return nil, result.Error
	}
	return gs, nil
}

func (s *Storage) DeleteGroup(g *groups.Group) error {
	result := s.db.Delete(g)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
