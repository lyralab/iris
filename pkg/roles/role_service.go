package roles

import (
	"errors"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"math/rand"
	"time"
)

type RolesInterfaceRepository interface {
	AddRole(Role) error
	GetRoleByName(*Role) error
	GetRoleName(*Role) error
}

type RoleInterfaceService interface {
	AddRoles([]Role) error
	InitiateDefaultRoles() error
	GetRoleByID(string) (string, string, error)
	GetRoleIDByName(name string) (string, error)
}

func NewRolesService(logger *zap.SugaredLogger, rr RolesInterfaceRepository) RoleInterfaceService {

	return &roleServiceImpl{
		rr:     rr,
		logger: logger,
	}
}

func (r *roleServiceImpl) AddRoles([]Role) error {
	return nil
}

func (r *roleServiceImpl) InitiateDefaultRoles() error {
	adminRole := Role{Name: "admin"}
	err := r.rr.GetRoleByName(&adminRole)
	if errors.Is(err, logger.ErrRecordNotFound) {
		t := time.Now()
		entropy := rand.New(rand.NewSource(t.UnixNano()))
		adminRole.ID = ulid.MustNew(ulid.Timestamp(t), entropy).String()
		adminRole.Name = "admin"
		adminRole.Access = "*"
		adminRole.Created_at = t
		adminRole.Modified_at = t
		err = r.rr.AddRole(adminRole)
		if err != nil {
			r.logger.Infow("Cannot create default admin role", "error", err)
		}
	} else if err == nil {
		r.logger.Info("Admin role already exist nothing to do... ")
	} else {
		r.logger.Error("Cannot create default admin role ", "error", err)
		return err
	}
	viewerRole := Role{Name: "viewer"}
	err = r.rr.GetRoleByName(&viewerRole)
	if errors.Is(err, logger.ErrRecordNotFound) {
		t := time.Now()
		entropy := rand.New(rand.NewSource(t.UnixNano()))
		viewerRole.ID = ulid.MustNew(ulid.Timestamp(t), entropy).String()
		viewerRole.Name = "viewer"
		viewerRole.Access = ""
		viewerRole.Created_at = t
		viewerRole.Modified_at = t
		err = r.rr.AddRole(viewerRole)
		if err != nil {
			r.logger.Infow("Cannot create default viewer role", "error", err)
		}
	} else if err == nil {
		r.logger.Info("Viewer role already exist nothing to do... ")
	} else {
		r.logger.Error("Cannot create default viewer role ", "error", err)
		return err
	}
	return nil
}

func (r *roleServiceImpl) GetRoleByID(id string) (string, string, error) {
	role := Role{ID: id}
	err := r.rr.GetRoleName(&role)
	if err != nil {
		return "", "", err
	}
	return role.Name, role.Access, nil
}

func (r *roleServiceImpl) GetRoleIDByName(name string) (string, error) {
	role := Role{Name: name}
	err := r.rr.GetRoleByName(&role)
	if err != nil {
		return "", nil
	}
	return role.ID, nil
}
