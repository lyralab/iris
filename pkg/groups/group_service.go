package groups

import (
	"errors"
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/util"
	"go.uber.org/zap"
	"time"
)

func NewGroupService(logger *zap.SugaredLogger, gr GroupRepositoryInterface, cache ReceptorCache) GroupServiceInterface {
	return &GroupService{log: logger, gr: gr, cache: cache}
}

func (gr *GroupService) CreateGroup(g *Group) error {
	checkGroup, err := gr.gr.GetGroupByName(g.Name)
	if err != nil {
		if errors.Is(err, apperr.ErrGroupNotFound) {
			gr.log.Infow("Group does not exist, creating new group", "group", g.Name)
			return gr.insertGroup(g)
		}
		gr.log.Errorw("Error getting group", "error", err)
		return err
	}
	if checkGroup != nil {
		gr.log.Infow("Group already exists", "group", checkGroup.Name)
		return apperr.ErrGroupAlreadyexisted
	}
	return gr.insertGroup(g)
}

func (gr *GroupService) insertGroup(g *Group) error {
	groupID, err := util.NewUUIDv7()
	if err != nil {
		return err
	}
	g.ID = groupID
	now := time.Now()
	g.CreatedAt = now
	g.ModifiedAt = now
	gr.log.Infow("we are going to save group", "group", g)
	if err := gr.gr.AddGroup(g); err != nil {
		return err
	}
	gr.refreshReceptors()
	return nil
}

func (gr *GroupService) GetGroup(name string) (*Group, error) {
	return gr.gr.GetGroupByName(name)
}

func (gr *GroupService) GetGroupById(id string) (*Group, error) {
	return gr.gr.GetGroupById(id)
}

func (gr *GroupService) DeleteGroup(g *Group) error {
	if err := gr.gr.DeleteGroup(g); err != nil {
		return err
	}
	gr.refreshReceptors()
	return nil
}

func (gr *GroupService) GetAllGroups() ([]*Group, error) {
	return gr.gr.GetAllGroups()
}

func (gr *GroupService) AddUser(g *Group, userId string) error {
	if err := gr.gr.AddUserToGroup(userId, g.ID); err != nil {
		return err
	}
	gr.refreshReceptors()
	return nil
}

func (gr *GroupService) RemoveUser(g *Group, userId string) error {
	if err := gr.gr.RemoveUserFromGroup(userId, g.ID); err != nil {
		return err
	}
	gr.refreshReceptors()
	return nil
}

func (gr *GroupService) refreshReceptors() {
	if gr.cache == nil {
		return
	}
	gr.cache.Refresh()
}

func (gr *GroupService) ListUsers(g *Group) ([]string, error) {
	return gr.gr.FindUsersById(g.ID)
}

func (gr *GroupService) ListGroupByUser(userId string) ([]*Group, error) {
	var g []*Group
	groupIDs, err := gr.gr.FindGroupById(userId)
	if err != nil {
		return nil, err
	}
	for _, groupId := range groupIDs {
		group, err := gr.gr.GetGroupById(groupId)
		if err != nil {
			return nil, err
		}
		g = append(g, group)
	}
	return g, nil
}
