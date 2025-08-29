package main

import (
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/auth"
	"github.com/root-ali/iris/pkg/cache"
	"github.com/root-ali/iris/pkg/captcha"
	"github.com/root-ali/iris/pkg/groups"
	"github.com/root-ali/iris/pkg/health_check"
	"github.com/root-ali/iris/pkg/http"
	migrationpostgresql "github.com/root-ali/iris/pkg/migration/postgresql"
	"github.com/root-ali/iris/pkg/notifications"
	"github.com/root-ali/iris/pkg/notifications/kavenegar"
	"github.com/root-ali/iris/pkg/notifications/smsir"
	"github.com/root-ali/iris/pkg/roles"
	"github.com/root-ali/iris/pkg/scheduler/cache_receptors"
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"github.com/root-ali/iris/pkg/user"
	"github.com/root-ali/iris/pkg/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Iris is the main configuration structure for the Iris application.
type Iris struct {

	// Postgres holds the configuration for the PostgreSQL database connection.
	Postgres struct {
		Host string `env:"POSTGRES_HOST"`
		Port string `env:"POSTGRES_PORT"`
		Name string `env:"POSTGRES_DATABASE_NAME"`
		User string `env:"POSTGRES_USER"`
		Pass string `env:"POSTGRES_PASS"`
		SSL  bool   `env:"POSTGRES_SSL default=false"`
	}

	// Http holds the configuration for the HTTP server.
	Http struct {
		Port      string `env:"HTTP_PORT" envDefault:"9090"`
		AdminPass string `env:"ADMIN_PASS"`
	}
	Go struct {
		Mode string `env:"GO_ENV" envDefault:"debug"`
	}
	Notifications struct {
		// Smsir is the configuration for the Smsir notification service.
		Smsir struct {
			// ApiKey is the API token for Smsir service.
			ApiKey string `env:"SMSIR_API_TOKEN"`
			// LineNumber is the phone number that will be used to send SMS messages.
			LineNumber int `env:"SMSIR_LINE_NUMBER" envDefault:"30007732911486"`
			// Enabled indicates whether the Smsir service is enabled.
			Enabled bool `env:"SMSIR_ENABLED" envDefault:"false"`
			// Priority is the priority of the Smsir service.
			Priority int `env:"SMSIR_PRIORITY" envDefault:"1"`
		}
		// Kavenegar is the configuration for the Kavenegar notification service.
		Kavenegar struct {
			// ApiToken is the API token for Kavenegar service.
			ApiToken string `env:"KAVENEGAR_API_TOKEN"`
			// Sender is the phone number that will be used to send SMS messages.
			Sender string `env:"KAVENEGAR_SENDER" envDefault:""`
			// Enabled indicates whether the Kavenegar service is enabled.
			Enabled bool `env:"KAVENEGAR_ENABLED" envDefault:"false"`
			// Priority is the priority of the Kavenegar service.
			Priority int `env:"KAVENEGAR_PRIORITY" envDefault:"2"`
		}
		// Email is the configuration for the email notification service.
		Email struct {
			Host     string `env:"EMAIL_HOST"`
			Port     string `env:"EMAIL_PORT"`
			User     string `env:"EMAIL_USER"`
			Password string `env:"EMAIL_PASSWORD"`
			From     string `env:"EMAIL_FROM"`
			Enabled  bool   `env:"EMAIL_ENABLED" envDefault:"false"`
		}
	}

	// Scheduler holds the configuration for the  schedulers.
	Scheduler struct {
		// MobileScheduler is the configuration for the mobile scheduler.
		MobileScheduler struct {
			// StartAt is the time when the scheduler should start.
			StartAt time.Duration `env:"MOBILE_SCHEDULER_START_AT" envDefault:"200"`
			// Interval is the interval at which the scheduler should run.
			Interval time.Duration `env:"MOBILE_SCHEDULER_INTERVAL" envDefault:"600s"`
			// Workers is the number of workers that should be used by the scheduler.
			Workers int `env:"MOBILE_SCHEDULER_WORKERS" envDefault:"1"`
			// QueueSize is the size of the queue that should be used by the scheduler.
			QueueSize int `env:"MOBILE_SCHEDULER_QUEUE_SIZE" envDefault:"1"`
			// CacheCapacity is the capacity of the cache that should be used by the scheduler.
			CacheCapacity int `env:"MOBILE_SCHEDULER_CACHE_CAPACITY" envDefault:"1000"`
		}
	}

	// Logger is the logger used throughout the application.
	Logger *zap.SugaredLogger

	// JwtSecret is the secret key used for signing JWT tokens.
	JwtSecret string `env:"JWT_SECRET"`

	// SignupEnabled indicates whether user signup is enabled.
	SignupEnabled bool `env:"SIGNUP_ENABLED" envDefault=true`

	// PostgresRepositories holds the storage layer for the application.
	PostgresRepositories *postgresql.Storage

	// NotificationServices holds the notification services used in the application.
	NotificationServices []notifications.NotificationInterface
	// ProviderService is the service for managing notification providers.
	ProviderService notifications.ProviderServiceInterface
}

func main() {
	var i Iris
	if err := env.Parse(&i); err != nil {
		fmt.Printf("%+v\n", err)
	}

	logger, err := configureZapLogger(i.Go.Mode)
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}(logger)
	i.Logger = logger.Sugar()

	if i.JwtSecret == "" {
		i.Logger.Panicw("JWT_SECRET environment variable not set")
	}
	postgres := &postgresql.Postgres{
		Host:     i.Postgres.Host,
		User:     i.Postgres.User,
		SSLMode:  i.Postgres.SSL,
		Password: i.Postgres.Pass,
		Port:     i.Postgres.Port,
		DBname:   i.Postgres.Name,
	}
	i.migratePostgresDatabase(postgres)
	i.initiateRepositories(postgres)
	i.initNotificationsService(i.Logger)
	i.initializeMobileScheduler()
	i.initializeCacheService()
	i.initializeProviderService()
	alertService := alerts.NewAlertService(i.Logger, i.PostgresRepositories)
	healthService := health_check.NewHealthService(i.Logger, i.PostgresRepositories)
	roleService := roles.NewRolesService(i.Logger, i.PostgresRepositories)
	userService := user.NewUserService(i.PostgresRepositories, roleService, i.Logger)
	authService := auth.NewAuthService([]byte(i.JwtSecret), roleService, i.Logger)
	groupService := groups.NewGroupService(i.Logger, i.PostgresRepositories)
	captchaService := captcha.NewCaptchaService(i.Logger)
	err = roleService.InitiateDefaultRoles()
	if err != nil {
		i.Logger.Panicw("Cannot create default roles", "error", err)
	}
	err = userService.CreateDefaultAdminUser()
	if err != nil {
		i.Logger.Panicw("Cannot create default admin user", "error", err)
	}
	hh := http.HttpHandler{
		AS:            alertService,
		HS:            healthService,
		US:            userService,
		ATHS:          authService,
		GR:            groupService,
		CS:            captchaService,
		PS:            i.ProviderService,
		AdminPassword: "admin",
		GinMode:       "debug",
		SignupEnabled: i.SignupEnabled,
		Logger:        i.Logger,
	}
	// start http server
	router := hh.Handler()

	log.Fatal(router.Run(":" + i.Http.Port))

}

func (i *Iris) migratePostgresDatabase(p *postgresql.Postgres) {
	mp := migrationpostgresql.NewPostgresMigrate(i.Logger, p)
	if err := mp.Migrate(); err != nil {
		i.Logger.Panicw("Something went wrong in migration please check logs", "Error", err)
	}
}

func (i *Iris) initiateRepositories(p *postgresql.Postgres) {
	i.PostgresRepositories = postgresql.New(i.Logger, p)
}

func configureZapLogger(mode string) (logger *zap.Logger, err error) {
	var config zap.Config

	if mode == "debug" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	config.Encoding = "json"

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err = config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func (i *Iris) initNotificationsService(logger *zap.SugaredLogger) {
	if i.Notifications.Smsir.Enabled {
		smsirService := smsir.NewSmsirService(i.Notifications.Smsir.ApiKey, 30007732911486, i.Notifications.Smsir.Priority, logger)
		i.NotificationServices = append(i.NotificationServices, smsirService)
		VerifySmsirService, err := smsirService.Verify()
		if err != nil {
			logger.Errorw("Failed to verify Smsir service", "error", err)
		} else {
			logger.Infow("Smsir service verified successfully", "response", VerifySmsirService)
		}
		i.Logger.Info("Smsir service initialized")
	}
	if i.Notifications.Kavenegar.Enabled {
		kavenegarService := kavenegar.NewKavenegarService(i.Notifications.Kavenegar.ApiToken, i.Notifications.Kavenegar.Priority, "", logger)
		i.NotificationServices = append(i.NotificationServices, kavenegarService)
		verifyKavenegarService, err := kavenegarService.Verify()
		if err != nil {
			logger.Errorw("Failed to verify Kavenegar service",
				"error", err)
		} else {
			logger.Infow("Kavenegar service verified successfully",
				"response", verifyKavenegarService)
		}
	}
}

func (i *Iris) initializeCacheService() {
	cacheService := cache.New[string, string](i.Logger,
		cache.WithCapacity(1000),
		cache.WithCleanupInterval(5),
		cache.WithJanitor(true))
	err := cacheService.Set("exampleKey", "exampleValue", 0)
	if err != nil {
		i.Logger.Errorw("Cache service failed to set value", "error", err)
		return
	}
	value, found := cacheService.Get("exampleKey")
	if found {
		i.Logger.Infow("Cache service initialized", "key", "exampleKey", "value", value)
	} else {
		i.Logger.Errorw("Cache service failed to retrieve value for key", "key", "exampleKey")
	}
	cacheService.Delete("exampleKey")
}

func (i *Iris) initializeMobileScheduler() {
	cacheService := cache.New[string, []string](i.Logger,
		cache.WithCapacity(i.Scheduler.MobileScheduler.CacheCapacity))

	cacheReceptorConfig := cache_receptors.Config{
		StartAt:   time.Now().Add(i.Scheduler.MobileScheduler.StartAt * time.Second),
		Interval:  i.Scheduler.MobileScheduler.Interval,
		Workers:   i.Scheduler.MobileScheduler.Workers,
		QueueSize: i.Scheduler.MobileScheduler.QueueSize,
	}
	cacheReceptorService, err := cache_receptors.NewCacheReceptorsScheduler(i.PostgresRepositories,
		cacheService, i.Logger, cacheReceptorConfig)
	if err != nil {
		i.Logger.Errorw("Failed to create cache receptor service", "error", err)
	}
	err = cacheReceptorService.Start()
	if err != nil {
		i.Logger.Errorw("Failed to start cache receptor service", "error", err)
	} else {
		i.Logger.Info("Cache receptor service started successfully")
	}
}

func (i *Iris) initializeProviderService() {
	ps := notifications.NewProvidersService(i.PostgresRepositories, i.Logger)
	i.ProviderService = ps
	for _, provider := range i.NotificationServices {
		id, err := util.NewUUIDv7()
		if err != nil {
			panic(err)
		}
		p := &notifications.Providers{
			ID:          id,
			Name:        provider.GetName(),
			Description: fmt.Sprintf("%s provider", provider.GetName()),
			Flag:        provider.GetFlag(),
			Priority:    provider.GetPriority(),
			Provider:    provider,
			Status:      true,
		}
		err = ps.AddProvider(p)
		if err != nil {
			i.Logger.Errorw("Failed to add provider", "provider", provider.GetName(), "error", err)
		} else {
			i.Logger.Infow("Provider added successfully", "provider", provider.GetName())
		}
	}
}
