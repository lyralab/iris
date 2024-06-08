package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/health_check"
	"net/http"
)

func Healthy(hs health_check.HealthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		health, err := hs.Healthy()
		if err != nil {
			c.JSON(http.StatusInternalServerError, health)
		} else {
			c.JSON(http.StatusOK, health)
		}
	}
}

func Ready(hs health_check.HealthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := hs.Ready()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status": "OK",
			})
		}

	}
}
