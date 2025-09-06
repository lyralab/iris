package server

import (
	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/auth"
	"github.com/root-ali/iris/pkg/captcha"
	"github.com/root-ali/iris/pkg/groups"
	"github.com/root-ali/iris/pkg/health_check"
	"github.com/root-ali/iris/pkg/http"
	"github.com/root-ali/iris/pkg/notifications"
	"github.com/root-ali/iris/pkg/roles"
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"github.com/root-ali/iris/pkg/user"
	"go.uber.org/zap"
)

type Deps struct {
	Logger          *zap.SugaredLogger
	Repos           *postgresql.Storage
	JWTSecret       []byte
	SignupEnabled   bool
	ProviderService notifications.ProviderServiceInterface
	AdminPass       string
	GinMode         string
}

func RegisterRoutes(d Deps) *gin.Engine {
	alertService := alerts.NewAlertService(d.Logger, d.Repos)
	healthService := health_check.NewHealthService(d.Logger, d.Repos)
	roleService := roles.NewRolesService(d.Logger, d.Repos)
	userService := user.NewUserService(d.Repos, roleService, d.Logger)
	authService := auth.NewAuthService(d.JWTSecret, roleService, d.Logger)
	groupService := groups.NewGroupService(d.Logger, d.Repos)
	captchaSvc := captcha.NewCaptchaService(d.Logger)

	if err := roleService.InitiateDefaultRoles(); err != nil {
		d.Logger.Panicw("Cannot create default roles", "error", err)
	}
	if err := userService.CreateDefaultAdminUser(); err != nil {
		d.Logger.Panicw("Cannot create default admin user", "error", err)
	}

	h := http.HttpHandler{
		AS:            alertService,
		HS:            healthService,
		US:            userService,
		ATHS:          authService,
		GR:            groupService,
		CS:            captchaSvc,
		PS:            d.ProviderService,
		AdminPassword: d.AdminPass,
		GinMode:       d.GinMode,
		SignupEnabled: d.SignupEnabled,
		Logger:        d.Logger,
	}

	return h.Handler()
}
