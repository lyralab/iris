package main

// TODO add documantion for every function and types
// TODO add project config
// TODO add start and exit service function

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/auth"
	"github.com/root-ali/iris/pkg/captcha"
	"github.com/root-ali/iris/pkg/groups"
	"github.com/root-ali/iris/pkg/health_check"
	"github.com/root-ali/iris/pkg/http"
	migrationpostgresql "github.com/root-ali/iris/pkg/migration/postgresql"
	"github.com/root-ali/iris/pkg/roles"
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"github.com/root-ali/iris/pkg/user"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Iris struct {
	Postgres struct {
		Host string `env:"POSTGRES_HOST"`
		Port string `env:"POSTGRES_PORT"`
		Name string `env:"POSTGRES_DATABASE_NAME"`
		User string `env:"POSTGRES_USER"`
		Pass string `env:"POSTGRES_PASS"`
		SSL  bool   `env:"POSTGRES_SSL default=false"`
	}
	Http struct {
		Port      string `env:"HTTP_PORT" envDefault:"9090"`
		AdminPass string `env:"ADMIN_PASS"`
	}
	Go struct {
		Mode string `env:"GO_ENV" envDefault:"debug"`
	}

	Logger *zap.SugaredLogger

	JwtSecret            string `env:"JWT_SECRET"`
	SignupEnabled        bool   `env:"SIGNUP_ENABLED" envDefault=true`
	postgresRepositories *postgresql.Storage
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
	alertService := alerts.NewAlertService(i.Logger, i.postgresRepositories)
	healthService := health_check.NewHealthService(i.Logger, i.postgresRepositories)
	roleService := roles.NewRolesService(i.Logger, i.postgresRepositories)
	userService := user.NewUserService(i.postgresRepositories, roleService, i.Logger)
	authService := auth.NewAuthService([]byte(i.JwtSecret), roleService, i.Logger)
	groupService := groups.NewGroupService(i.Logger, i.postgresRepositories)
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
	i.postgresRepositories = postgresql.New(i.Logger, p)
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
