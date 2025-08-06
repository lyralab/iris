package groups

import (
	"errors"
	iris_error "github.com/root-ali/iris/pkg/errors"
	"github.com/root-ali/iris/pkg/util"
	"go.uber.org/zap"
	"time"
)

func NewGroupService(logger *zap.SugaredLogger, gr GroupRepositoryInterface) GroupServiceInterface {
	return &GroupService{log: logger, gr: gr}
}

func (gr *GroupService) CreateGroup(g *Group) error {
	checkGroup, err := gr.gr.GetGroupByName(g.Name)
	if err != nil {
		if errors.Is(err, iris_error.ErrGroupNotFound) {
			gr.log.Infow("Group does not exist, creating new group", "group", g.Name)
			groupId, _ := util.NewUUIDv7()
			g.ID = groupId
			g.CreatedAt = time.Now()
			g.ModifiedAt = time.Now()
			gr.log.Infow("we are going to save group", "group", g)
			return gr.gr.AddGroup(g)
		}
		gr.log.Errorw("Error getting group", "error", err)
		return err
	}
	if checkGroup != nil {
		gr.log.Infow("Group already exists", "group", checkGroup.Name)
		return iris_error.ErrGroupAlreadyexisted
	}
	groupId, _ := util.NewUUIDv7()
	g.ID = groupId
	g.CreatedAt = time.Now()
	g.ModifiedAt = time.Now()
	gr.log.Infow("we are going to save group", "group", g)
	return gr.gr.AddGroup(g)
}

func (gr *GroupService) GetGroup(name string) (*Group, error) {
	return gr.gr.GetGroupByName(name)
}

func (gr *GroupService) DeleteGroup(g *Group) error {
	return gr.gr.DeleteGroup(g)
}

func (gr *GroupService) GetAllGroups() ([]*Group, error) {
	return gr.gr.GetAllGroups()
}

func (gr *GroupService) AddUser(g *Group, userId string) error {
	return gr.gr.AddUserToGroup(userId, g.ID)
}

func (gr *GroupService) RemoveUser(g *Group, userId string) error {
	return gr.gr.RemoveUserFromGroup(userId, g.ID)
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
