package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/alerts"
)

type AlertResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	StartsAt    string `json:"starts_at"`
	EndsAt      string `json:"ends_at"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func GetAlerts(as alerts.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var status, limit, page, severity string
		var l, p int
		var err error
		status = c.Request.URL.Query().Get("status")
		page = c.Request.URL.Query().Get("page")
		limit = c.Request.URL.Query().Get("limit")
		severity = c.Request.URL.Query().Get("severity")
		if limit != "" {
			l, err = strconv.Atoi(limit)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid page in query param",
				})
				return
			}
		}
		if page != "" {
			p, err = strconv.Atoi(page)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid page in query param",
				})
				return
			}
		}
		als, err := as.GetAlerts(status, severity, l, p)
		alertResponses := toAlertResponse(als)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status": "OK",
				"alerts": alertResponses,
				"count":  len(alertResponses),
			})
		}
	}
}

func GetFiringAlertsBySeverity(as alerts.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		als, err := as.GetFiringAlertsBySeverity()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "An error happend while getting records",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":    "OK",
				"severites": als,
			})
		}
	}
}

func toAlertResponse(alert []*alerts.Alert) []AlertResponse {
	var alertResponses []AlertResponse
	for _, a := range alert {
		alertResponse := AlertResponse{
			Id:          a.Id,
			Name:        a.Name,
			Severity:    a.Severity,
			Description: a.Description,
			StartsAt:    a.StartsAt.String(),
			EndsAt:      a.EndsAt.String(),
			Status:      a.Status,
			CreatedAt:   a.CreatedAt.String(),
			UpdatedAt:   a.UpdatedAt.String(),
		}
		alertResponses = append(alertResponses, alertResponse)
	}
	return alertResponses
}
