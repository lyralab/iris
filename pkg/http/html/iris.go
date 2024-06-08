package html

import (
	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/alerts"
	"net/http"
)

func Index(as alerts.AlertsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		als, _ := as.GetFiringAlertsBySeverity()
		firing, _ := as.GetAlerts("firing", "critical", 10, 1)
		resolved, _ := as.GetAlerts("resolved", "", 10, 1)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"severity": &als,
			"alerts":   &firing,
			"resolved": &resolved,
		})
	}
}
