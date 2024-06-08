package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/messages/alertmanager"
	"net/http"
)

func AlertManagerHandler(as alerts.AlertsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := c.GetRawData()
		if len(body) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "request body is empty",
			})
		} else {
			//fmt.Println("Request Body", c.Request.Body)
			amr := &alertmanager.AlertManager{}
			err := json.Unmarshal(body, amr)
			if err != nil {
				c.JSON(http.StatusNotAcceptable, gin.H{
					"status":  "error",
					"message": err,
				})
				fmt.Println(err)
			} else {
				as.AddAlertManagerAlerts(amr.Alerts)
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			}
		}
	}
}
