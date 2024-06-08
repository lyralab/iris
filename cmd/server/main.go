package main

// TODO add documantion for every fucntion and types
// TODO add project config
// TODO add start and exit service function
// TODO add database migratin

import (
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/health_check"
	"github.com/root-ali/iris/pkg/http"
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"go.uber.org/zap"
	"log"
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
}

func main() {
	var cfg Iris
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	l, _ := zap.NewDevelopment()
	postgres := &postgresql.Postgres{
		Host:     cfg.Postgres.Host,
		User:     cfg.Postgres.User,
		SSLMode:  cfg.Postgres.SSL,
		Password: cfg.Postgres.Pass,
		Port:     cfg.Postgres.Port,
		DBname:   cfg.Postgres.Name,
	}
	db := postgresql.New(l.Sugar(), postgres)
	as := alerts.NewAlertService(l.Sugar(), db)
	hs := health_check.NewHealthService(l.Sugar(), db)
	// start http server
	router := http.Handler(as, hs, cfg.Http.AdminPass, cfg.Go.Mode)

	log.Fatal(router.Run(":" + cfg.Http.Port))

}
