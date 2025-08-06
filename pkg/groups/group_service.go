package groups

import (
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
