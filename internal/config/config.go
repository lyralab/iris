package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

type Postgres struct {
	Host string `env:"POSTGRES_HOST"`
	Port string `env:"POSTGRES_PORT"`
	Name string `env:"POSTGRES_DATABASE_NAME"`
	User string `env:"POSTGRES_USER"`
	Pass string `env:"POSTGRES_PASS"`
	SSL  bool   `env:"POSTGRES_SSL" envDefault:"false"`
}

type HTTP struct {
	Port      string `env:"HTTP_PORT" envDefault:"9090"`
	AdminPass string `env:"ADMIN_PASS"`
}

type GoEnv struct {
	Mode string `env:"GO_ENV" envDefault:"debug"`
}

type Notifications struct {
	Smsir struct {
		ApiKey     string `env:"SMSIR_API_TOKEN"`
		LineNumber string `env:"SMSIR_LINE_NUMBER" envDefault:"30007732911486"`
		Enabled    bool   `env:"SMSIR_ENABLED" envDefault:"false"`
		Priority   int    `env:"SMSIR_PRIORITY" envDefault:"2"`
	}
	Kavenegar struct {
		ApiToken string `env:"KAVENEGAR_API_TOKEN"`
		Sender   string `env:"KAVENEGAR_SENDER" envDefault:""`
		Enabled  bool   `env:"KAVENEGAR_ENABLED" envDefault:"false"`
		Priority int    `env:"KAVENEGAR_PRIORITY" envDefault:"1"`
	}
	Email struct {
		Host     string `env:"EMAIL_HOST"`
		Port     string `env:"EMAIL_PORT"`
		User     string `env:"EMAIL_USER"`
		Password string `env:"EMAIL_PASSWORD"`
		From     string `env:"EMAIL_FROM"`
		Enabled  bool   `env:"EMAIL_ENABLED" envDefault:"false"`
	}
}

type Scheduler struct {
	MobileScheduler struct {
		StartAt       time.Duration `env:"MOBILE_SCHEDULER_START_AT" envDefault:"200s"`
		Interval      time.Duration `env:"MOBILE_SCHEDULER_INTERVAL" envDefault:"600s"`
		Workers       int           `env:"MOBILE_SCHEDULER_WORKERS" envDefault:"1"`
		QueueSize     int           `env:"MOBILE_SCHEDULER_QUEUE_SIZE" envDefault:"1"`
		CacheCapacity int           `env:"MOBILE_SCHEDULER_CACHE_CAPACITY" envDefault:"1000"`
	}
	AlertScheduler struct {
		Interval  time.Duration `env:"ALERT_SCHEDULER_INTERVAL" envDefault:"10s"`
		Workers   int           `env:"ALERT_SCHEDULER_WORKERS" envDefault:"1"`
		QueueSize int           `env:"ALERT_SCHEDULER_QUEUE_SIZE" envDefault:"10"`
	}
}

type Config struct {
	Postgres      Postgres
	HTTP          HTTP
	Go            GoEnv
	Notifications Notifications
	Scheduler     Scheduler
	JwtSecret     string `env:"JWT_SECRET"`
	SignupEnabled bool   `env:"SIGNUP_ENABLED" envDefault:"true"`
}

func Load() (*Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
