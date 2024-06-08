package http

import (
	"fmt"
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/health_check"
	"github.com/root-ali/iris/pkg/http/html"
	"github.com/root-ali/iris/pkg/http/rest"
	"time"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
)

func Handler(as alerts.AlertsService, hs health_check.HealthService, adminPassword string, ginMode string) *gin.Engine {
	router := gin.Default()

	if ginMode != "production" && ginMode != "test" {
		gin.SetMode(gin.DebugMode)
	} else if ginMode == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	m := ginmetrics.GetMonitor()

	m.SetMetricPath("/metrics")
	m.SetSlowTime(10)
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
	m.Use(router)
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())
	router.Use(helmet.Default())
	router.SetTrustedProxies(nil)
	router.LoadHTMLGlob("pkg/http/html/*")
	router.GET("/ready", rest.Ready(hs))
	router.GET("/healthy", rest.Healthy(hs))
	messageRouter := router.Group("v0/messages", gin.BasicAuth(gin.Accounts{
		"admin": adminPassword,
	}))
	alertRouter := router.Group("v0/alerts", gin.BasicAuth(gin.Accounts{
		"admin": adminPassword,
	}))
	messageRouter.POST("/alertmanager", rest.AlertManagerHandler(as))
	alertRouter.GET("/", rest.GetAlerts(as))
	alertRouter.GET("", rest.GetAlerts(as))
	alertRouter.GET("/firingCount", rest.GetFiringAlertsBySeverity(as))
	router.GET("/", html.Index(as))
	return router
}
