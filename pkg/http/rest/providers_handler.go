package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/message"
	"github.com/root-ali/iris/pkg/notifications"
)

type modifyProviderRequest struct {
	Name     *string `json:"name" binding:"required"`
	Priority *int    `json:"priority"`
	Status   *bool   `json:"status"`
}

type ProviderResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

func GetProvidersHandler(ps notifications.ProviderServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {

		id := c.Query("id")
		name := c.Query("name")
		var pr ProviderResponse
		if id != "" && name != "" {
			abortWithMessage(c, http.StatusBadRequest, "provide either id or name, not both")
			return
		}
		if name != "" {
			provider, err := ps.GetProviderByName(name)
			if err != nil {
				abortWithError(c, err)
				return
			}
			if provider == nil {
				abortWithError(c, apperr.ErrProviderNotFound)
				return
			}

			pr.ID = provider.ID
			pr.Name = provider.Name
			pr.Description = provider.Description
			pr.Priority = provider.Priority
			pr.Enabled = provider.Status
			c.AbortWithStatusJSON(200, gin.H{"provider": pr, "status": "success"})
			return
		}
		if id != "" {
			provider, err := ps.GetProviderByID(id)
			if err != nil {
				abortWithError(c, err)
				return
			}
			if provider == nil {
				abortWithError(c, apperr.ErrProviderNotFound)
				return
			}

			pr.ID = provider.ID
			pr.Name = provider.Name
			pr.Description = provider.Description
			pr.Priority = provider.Priority
			pr.Enabled = provider.Status
			c.AbortWithStatusJSON(200, gin.H{"provider": pr, "status": "success"})
		}

		if id == "" && name == "" {

			providers, err := ps.GetAllProviders()
			providerResponses := make([]ProviderResponse, 0, len(providers))
			for p := range providers {
				providerResponses = append(providerResponses, ProviderResponse{
					ID:          providers[p].ID,
					Name:        providers[p].Name,
					Description: providers[p].Description,
					Priority:    providers[p].Priority,
					Enabled:     providers[p].Status,
				})
			}
			if err != nil {
				abortWithError(c, err)
				return
			}
			c.JSON(200, gin.H{"providers": providerResponses, "status": "success"})

		}

	}
}

const (
	defaultMessagesLimit = 10
	maxMessagesLimit     = 100
)

type messageLister interface {
	ListBySender(sender string, limit int) ([]message.Sent, error)
}

// parseMessagesLimit falls back to the default for missing or invalid values and caps at the max.
func parseMessagesLimit(raw string) int {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return defaultMessagesLimit
	}
	return min(n, maxMessagesLimit)
}

func GetProviderMessagesHandler(ml messageLister) gin.HandlerFunc {
	return func(c *gin.Context) {
		msgs, err := ml.ListBySender(c.Param("name"), parseMessagesLimit(c.Query("limit")))
		if err != nil {
			abortWithError(c, err)
			return
		}
		c.JSON(200, gin.H{"messages": msgs, "status": "success"})
	}
}

func ModifyProviderHandler(ps notifications.ProviderServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req modifyProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			abortWithMessage(c, http.StatusBadRequest, "invalid request body")
			return
		}
		if *req.Name == "" {
			abortWithMessage(c, http.StatusBadRequest, "name is required")
			return
		}
		if req.Status == nil && req.Priority == nil {
			abortWithMessage(c, http.StatusBadRequest, "either priority or status must be provided")
			return
		}

		if req.Status == nil {

			if *req.Priority < 1 || *req.Priority > 5 {
				abortWithMessage(c, http.StatusBadRequest, "priority must be between 1 and 5")
				return
			} else {
				err := ps.ModifyProviderPriority(*req.Name, *req.Priority)
				if err != nil {
					abortWithError(c, err)
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"status": "success"})
				return
			}
		}
		if req.Priority == nil {
			if *req.Status {
				err := ps.EnableProvider(*req.Name)
				if err != nil {
					abortWithError(c, err)
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"status": "success"})
				return
			} else if !*req.Status {
				err := ps.DisableProvider(*req.Name)
				if err != nil {
					abortWithError(c, err)
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"status": "success"})
				return
			}
		}
		if req.Status != nil && req.Priority != nil {
			if *req.Priority < 1 || *req.Priority > 5 {
				abortWithMessage(c, http.StatusBadRequest, "priority must be between 1 and 5")
				return
			}
			if *req.Status {
				err := ps.EnableProvider(*req.Name)
				if err != nil {
					abortWithError(c, err)
					return
				}
				err = ps.ModifyProviderPriority(*req.Name, *req.Priority)
				if err != nil {
					abortWithError(c, err)
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"status": "success"})
				return
			} else if !*req.Status {
				err := ps.DisableProvider(*req.Name)
				if err != nil {
					abortWithError(c, err)
					return
				}
				err = ps.ModifyProviderPriority(*req.Name, *req.Priority)
				if err != nil {
					abortWithError(c, err)
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"status": "success"})
				return
			}
		}
	}
}
