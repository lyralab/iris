package http

import (
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/auth"
	"github.com/root-ali/iris/pkg/captcha"
	"github.com/root-ali/iris/pkg/groups"
	"github.com/root-ali/iris/pkg/health_check"
	"github.com/root-ali/iris/pkg/notifications"
	"github.com/root-ali/iris/pkg/user"
	"go.uber.org/zap"
)

type HttpHandler struct {
	AS            alerts.AlertsService
	HS            health_check.HealthService
	US            user.UserInterfaceService
	ATHS          auth.AuthServiceInterface
	GR            groups.GroupServiceInterface
	CS            captcha.CaptchaServiceInterface
	PS            notifications.ProviderServiceInterface
	AdminPassword string
	GinMode       string
	SignupEnabled bool
	Logger        *zap.SugaredLogger
}
