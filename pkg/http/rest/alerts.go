package rest

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/alerts"
	"net/http"
	"strconv"
)

func GetAlerts(as alerts.AlertsService) gin.HandlerFunc {
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
		alerts, err := as.GetAlerts(status, severity, l, p)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status": "OK",
				"alerts": alerts,
				"count":  len(alerts),
			})
		}
	}
}

func GetFiringAlertsBySeverity(as alerts.AlertsService) gin.HandlerFunc {
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
