package rest

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	iriserror "github.com/root-ali/iris/pkg/errors"
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
			c.AbortWithStatusJSON(400, gin.H{"error": "provide either id or name, not both"})
			return
		}
		if name != "" {
			fmt.Println("name:", name)
			provider, err := ps.GetProviderByName(name)
			if errors.Is(err, iriserror.ErrProviderNotFound) {
				c.AbortWithStatusJSON(404, gin.H{"error": err.Error()})
				return
			}
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}
			if provider == nil {
				c.AbortWithStatusJSON(404, gin.H{"error": "provider not found"})
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
			fmt.Println("name:", name, " id:", id)
			provider, err := ps.GetProviderByID(id)
			if errors.Is(err, iriserror.ErrProviderNotFound) {
				c.AbortWithStatusJSON(404, gin.H{"error": err.Error()})
				return
			} else if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}
			if provider == nil {
				c.AbortWithStatusJSON(404, gin.H{"error": "provider not found"})
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
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"providers": providerResponses, "status": "success"})

		}

	}
}

func ModifyProviderHandler(ps notifications.ProviderServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req modifyProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
			return
		}
		if *req.Name == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "name is required"})
			return
		}
		if req.Status == nil && req.Priority == nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "either priority or status must be provided"})
			return
		}

		if req.Status == nil {

			if *req.Priority < 1 || *req.Priority > 5 {
				c.AbortWithStatusJSON(400, gin.H{"error": "priority must be between 1 and 5"})
				return
			} else {
				err := ps.ModifyProviderPriority(*req.Name, *req.Priority)
				if err != nil {
					c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
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
					c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"status": "success"})
				return
			} else if !*req.Status {
				err := ps.DisableProvider(*req.Name)
				if err != nil {
					c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"status": "success"})
				return
			}
		}
		if req.Status != nil && req.Priority != nil {
			if *req.Priority < 1 || *req.Priority > 5 {
				c.AbortWithStatusJSON(400, gin.H{"error": "priority must be between 1 and 5"})
				return
			}
			if *req.Status {
				err := ps.EnableProvider(*req.Name)
				if err != nil {
					c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
					return
				}
				err = ps.ModifyProviderPriority(*req.Name, *req.Priority)
				if err != nil {
					c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"status": "success"})
				return
			} else if !*req.Status {
				err := ps.DisableProvider(*req.Name)
				if err != nil {
					c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
					return
				}
				err = ps.ModifyProviderPriority(*req.Name, *req.Priority)
				if err != nil {
					c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"status": "success"})
				return
			}
		}
	}
}
