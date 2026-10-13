package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	iris_error "github.com/root-ali/iris/pkg/errors"
	"github.com/root-ali/iris/pkg/groups"
	"github.com/root-ali/iris/pkg/user"
	"go.uber.org/zap"
)

type GroupsResponse struct {
	GroupID     string `json:"group_id"`
	GroupName   string `json:"group_name"`
	Description string `json:"description"`
}

type CreateGroupRequestBody struct {
	Name        string `json:"name" validate:"required,min=3,max=30"`
	Description string `json:"description,omitempty" validate:"omitempty,min=3,max=100"`
}

type AddUserToGroupRequestBody struct {
	UserID string `json:"user_id" validate:"required"`
}

type toGrouper interface {
	toGroup() *groups.Group
}

func CreateGroupHandler(gr groups.GroupServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Errorw("Failed to read request body", "error", err)
			c.JSON(400, gin.H{"error": "Failed to read request body"})
			return
		}
		var requestBody CreateGroupRequestBody
		err = json.Unmarshal(bodyBytes, &requestBody)
		if err != nil {
			logger.Errorw("Failed to unmarshal request body", "error", err)
			c.JSON(400, gin.H{"error": "Failed to unmarshal request body"})
			return
		}
		g := requestBody.toGroup()
		logger.Infow("Create group", "body", g)
		err = gr.CreateGroup(g)
		if err != nil {
			logger.Errorw("Failed to create group", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	}
}

func DeleteGroupHandler(gr groups.GroupServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Errorw("Failed to read request body", "error", err)
			c.JSON(400, gin.H{"error": "Failed to read request body"})
			return
		}
		var g *groups.Group
		err = json.Unmarshal(bodyBytes, g)
		if err != nil {
			logger.Errorw("Cannot Parse request body", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot Parse request body"})
			return
		}
		err = gr.DeleteGroup(g)
		if err != nil {
			logger.Errorw("Failed to delete group", "error", err)
		}
	}
}

func GetGroupHandler(gr groups.GroupServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupId := c.Param("group_id")
		if groupId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
			return
		}
		g, err := gr.GetGroupById(groupId)
		if errors.Is(err, iris_error.ErrGroupNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		if err != nil {
			logger.Errorw("Failed to get group", "error", err)
			c.JSON(400, gin.H{"error": "Failed to get group"})
			return
		}
		c.JSON(200, gin.H{"data": g, "status": "success"})
	}
}

func GetAllGroupHandler(gr groups.GroupServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupIDs, err := gr.GetAllGroups()
		if err != nil {
			logger.Errorw("Failed to get all groups", "error", err)
			c.AbortWithStatusJSON(400, gin.H{"error": "Failed to get all groups"})
			return
		}
		var response []GroupsResponse
		for _, g := range groupIDs {
			response = append(response, GroupsResponse{GroupID: g.ID, GroupName: g.Name, Description: g.Description})
		}
		c.JSON(200, gin.H{"data": response, "status": "success"})
	}
}

func GetUserGroupsHandler(gr groups.GroupServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		logger.Infow("GetUserGroupsHandler", "user_id", userId)
		groupIDs, err := gr.ListGroupByUser(userId)
		if err != nil {
			logger.Errorw("Failed to get user groups", "error", err)
			c.AbortWithStatusJSON(400, gin.H{"error": "Failed to get user groups"})
			return
		}
		var response []GroupsResponse
		for _, g := range groupIDs {
			response = append(response, GroupsResponse{GroupID: g.ID, GroupName: g.Name, Description: g.Description})
		}
		c.JSON(200, gin.H{"groups": response, "status": "success"})
		return
	}
}

func AddUserToGroupHandler(gr groups.GroupServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupId := c.Param("group_id")
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Errorw("Failed to read request body", "error", err)
			c.AbortWithStatusJSON(400, gin.H{"error": "Failed to read request body"})
			return
		}
		var requestBody AddUserToGroupRequestBody
		err = json.Unmarshal(bodyBytes, &requestBody)
		if err != nil {
			logger.Errorw("Failed to unmarshal request body", "error", err)
			c.AbortWithStatusJSON(400, gin.H{"error": "Failed to unmarshal request body"})
			return
		}
		userId := requestBody.UserID
		err = gr.AddUser(&groups.Group{ID: groupId}, userId)
		if err != nil {
			logger.Errorw("Failed to add user to group", "error", err)
			c.AbortWithStatusJSON(400, gin.H{"error": "Failed to add user to group"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func RemoveUserFromGroupHandler(gr groups.GroupServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupId := c.Param("group_id")
		userId := c.Param("user_id")
		if userId == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}
		err := gr.RemoveUser(&groups.Group{ID: groupId}, userId)
		if err != nil {
			logger.Errorw("Failed to remove user from group", "error", err)
			c.AbortWithStatusJSON(400, gin.H{"error": "Failed to remove user from group"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func GetUsersInGroupHandler(gr groups.GroupServiceInterface, us user.UserInterfaceService, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupId := c.Param("group_id")
		logger.Infow("GetUsersInGroupHandler", "group_id", groupId)
		gp := &groups.Group{ID: groupId}
		userIDs, err := gr.ListUsers(gp)
		if err != nil {
			logger.Errorw("Failed to get users in group", "error", err)
			c.AbortWithStatusJSON(400, gin.H{"error": "Failed to get users in group"})
			return
		}

		// Fetch full user details for each user ID
		var users []map[string]interface{}
		for _, userID := range userIDs {
			u, err := us.GetByUserId(userID)
			if err != nil {
				logger.Warnw("Failed to get user details", "user_id", userID, "error", err)
				continue // Skip this user if we can't fetch their details
			}
			users = append(users, map[string]interface{}{
				"id":       u.ID,
				"username": u.UserName,
				"email":    u.Email,
				"mobile":   u.Mobile,
			})
		}

		c.JSON(200, gin.H{"users": users, "status": "success"})
		return
	}
}

func ToGroup[T toGrouper](t T) *groups.Group {
	return t.toGroup()
}

func (g *CreateGroupRequestBody) toGroup() *groups.Group {
	group := &groups.Group{
		Name:        g.Name,
		Description: g.Description,
	}
	return group
}
