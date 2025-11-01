package http

import (
	"time"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
	"github.com/root-ali/iris/pkg/http/middlewares"
	"github.com/root-ali/iris/pkg/http/rest"
)

func (ht *HttpHandler) Handler() *gin.Engine {
	router := gin.Default()
	//ht.GinMode = "production"
	if ht.GinMode != "production" && ht.GinMode != "test" {
		gin.SetMode(gin.DebugMode)
	} else if ht.GinMode == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	m := ginmetrics.GetMonitor()

	m.SetMetricPath("/metrics")
	m.SetSlowTime(10)
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
	m.Use(router)
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			ht.Logger.Infow("Received HTTP request",
				"IP", param.ClientIP,
				"Timestamps", param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
				"Method", param.Method,
				"Path", param.Path,
				"Protocol", param.Request.Proto,
				"StatusCode", param.StatusCode,
				"Latency", param.Latency,
				"UserAgent", param.Request.UserAgent(),
				"BodySize", param.BodySize,
				"Error", param.ErrorMessage,
			)
			return ""
		},
	}))
	config := cors.DefaultConfig()
	config.AllowAllOrigins = false
	config.AllowCredentials = true
	config.AllowedOrigins = []string{"http://localhost:3000"}
	config.AllowedMethods = []string{"*"}
	config.AllowedHeaders = []string{"*"}
	config.MaxAge = 24 * time.Hour
	router.Use(cors.New(config))
	router.Use(gin.Recovery())
	router.Use(helmet.Default())
	err := router.SetTrustedProxies(nil)
	if err != nil {
		return nil
	}
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})
	router.GET("/ready", rest.Ready(ht.HS))
	router.GET("/healthy", rest.Healthy(ht.HS))

	// Captcha handler routes
	captchaRouter := router.Group("/v0/captcha")
	captchaRouter.GET("/generate",
		rest.GenerateCaptchaHandler(ht.CS, ht.Logger))
	// captchaRouter.POST("/verify",
	// 	middlewares.CheckContentTypeHeader("application/json", ht.Logger),
	// 	rest.VerifyCaptchaHandler(ht.CS, ht.Logger))

	// Message handler routes
	messageRouter := router.Group("v1/messages",
		middlewares.BasicAuth("admin", "admin"))
	messageRouter.POST("/alertmanager",
		rest.AlertManagerHandler(ht.AS))

	// Alerts handler routes
	alertRouter := router.Group("v0/alerts")
	alertRouter.GET("/",
		middlewares.ValidateJWTToken(ht.ATHS, "admin,viewer", ht.Logger),
		rest.GetAlerts(ht.AS))
	alertRouter.GET("",
		middlewares.ValidateJWTToken(ht.ATHS, "admin,viewer", ht.Logger),
		rest.GetAlerts(ht.AS))
	alertRouter.GET("/firingCount",
		middlewares.ValidateJWTToken(ht.ATHS, "admin,viewer", ht.Logger),
		rest.GetFiringAlertsBySeverity(ht.AS))

	// User handler routes
	userRouter := router.Group("v0/users")
	userRouter.POST("",
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		middlewares.CheckContentTypeHeader("application/json", ht.Logger),
		rest.AddUserHandler(ht.US, ht.Logger))
	userRouter.POST("/signin",
		middlewares.CheckContentTypeHeader("application/json", ht.Logger),
		middlewares.VerifyCaptchaMiddleware(ht.CS, ht.Logger),
		rest.LoginUserHandler(ht.US, ht.ATHS, ht.Logger),
	)
	userRouter.PUT("/verify",
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		middlewares.CheckContentTypeHeader("application/json", ht.Logger),
		rest.VerifyUserHandler(ht.US, ht.Logger),
	)
	userRouter.PUT("",
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		middlewares.CheckContentTypeHeader("application/json", ht.Logger),
		rest.UpdateUserHandler(ht.US, ht.Logger),
	)
	userRouter.GET("",
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		rest.GetAllUsersHandler(ht.US, ht.Logger),
	)
	userRouter.GET("/me",
		middlewares.ValidateJWTToken(ht.ATHS, "", ht.Logger),
		rest.GetUserInfoHandler(ht.US, ht.Logger),
	)
	// TODO: implement GetUserByID and GetUserByEmail handlers
	userRouter.GET("/:user_id",
		middlewares.ValidateJWTToken(ht.ATHS, "admin,viewer", ht.Logger),
		rest.GetUserByIDHandler(ht.US, ht.Logger),
	)
	// userRouter.GET("/email/:email",
	// 	middlewares.ValidateJWTToken(ht.ATHS, "admin,viewer", ht.Logger),
	// 	rest.GetUserByEmailHandler(ht.US, ht.Logger),
	// )
	userRouter.GET("/:user_id/groups",
		middlewares.ValidateJWTToken(ht.ATHS, "", ht.Logger),
		rest.GetUserGroupsHandler(ht.GR, ht.Logger),
	)

	// Group handler routes
	groupRouter := router.Group("v0/groups")
	groupRouter.POST("",
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		middlewares.CheckContentTypeHeader("application/json", ht.Logger),
		rest.CreateGroupHandler(ht.GR, ht.Logger))
	groupRouter.GET("",
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		rest.GetAllGroupHandler(ht.GR, ht.Logger))
	groupRouter.GET("/:group_id",
		middlewares.ValidateJWTToken(ht.ATHS, "admin,viewer", ht.Logger),
		rest.GetGroupHandler(ht.GR, ht.Logger))
	groupRouter.DELETE("",
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		middlewares.CheckContentTypeHeader("application/json", ht.Logger),
		rest.DeleteGroupHandler(ht.GR, ht.Logger))
	groupRouter.POST("/:group_id/users",
		middlewares.CheckContentTypeHeader("application/json", ht.Logger),
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		rest.AddUserToGroupHandler(ht.GR, ht.Logger))
	groupRouter.GET("/:group_id/users",
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		rest.GetUsersInGroupHandler(ht.GR, ht.US, ht.Logger),
	)

	providerRouter := router.Group("v0/providers")
	providerRouter.GET("",
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		rest.GetProvidersHandler(ht.PS))
	providerRouter.PUT("",
		middlewares.ValidateJWTToken(ht.ATHS, "admin", ht.Logger),
		middlewares.CheckContentTypeHeader("application/json", ht.Logger),
		rest.ModifyProviderHandler(ht.PS))

	// Serve static files from web/build
	router.Use(static.Serve("/", static.LocalFile("./web/build", true)))

	// Catch-all route to serve index.html for React Router (must be last)
	router.NoRoute(func(c *gin.Context) {
		c.File("./web/build/index.html")
	})

	return router
}
