package bootstrap

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/internal/config"
	"github.com/root-ali/iris/internal/logging"
	"github.com/root-ali/iris/internal/schedulers"
	"github.com/root-ali/iris/internal/server"
	"github.com/root-ali/iris/internal/storage"
	"github.com/root-ali/iris/pkg/cache"
	"github.com/root-ali/iris/pkg/message"
	"github.com/root-ali/iris/pkg/notifications"
	"github.com/root-ali/iris/pkg/notifications/asiatech"
	"github.com/root-ali/iris/pkg/notifications/kavenegar"
	"github.com/root-ali/iris/pkg/notifications/mail"
	"github.com/root-ali/iris/pkg/notifications/mattermost"
	"github.com/root-ali/iris/pkg/notifications/smsir"
	"github.com/root-ali/iris/pkg/notifications/telegram"
	"github.com/root-ali/iris/pkg/scheduler/message_status"
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"github.com/root-ali/iris/pkg/util"

	"go.uber.org/zap"
)

type App struct {
	Logger          *zap.SugaredLogger
	Repos           *storage.Repos
	ProviderService notifications.ProviderServiceInterface
	Services        []notifications.NotificationInterface
	Router          *gin.Engine
}

func Init(cfg *config.Config) (*App, error) {
	// logging
	zl, err := logging.New(cfg.Go.Mode)
	if err != nil {
		return nil, fmt.Errorf("logger init: %w", err)
	}
	logger := zl.Sugar()

	if cfg.JwtSecret == "" {
		logger.Panicw("JWT_SECRET not set")
	}

	// storage & migration
	repos, err := storage.Init(logger, postgresql.Postgres{
		Host:     cfg.Postgres.Host,
		User:     cfg.Postgres.User,
		SSLMode:  cfg.Postgres.SSL,
		Password: cfg.Postgres.Pass,
		Port:     cfg.Postgres.Port,
		DBname:   cfg.Postgres.Name,
	})
	if err != nil {
		return nil, err
	}

	// notifications (providers + schedulers)
	allServices := make([]notifications.NotificationInterface, 0)
	deactiveProviders := make([]string, 0)

	cacheReceptorsSchedulerStartAt, err := time.ParseDuration(cfg.Scheduler.MobileScheduler.StartAt)
	cacheReceptorsSchedulerInterval, err := time.ParseDuration(cfg.Scheduler.MobileScheduler.Interval)
	if err != nil {
		return nil, fmt.Errorf("incorrect cache receptors scheduler config: %w", err)
	}

	// Start cache receptors
	cr, err := schedulers.StartCacheReceptors(
		logger,
		repos.Postgres,
		cacheReceptorsSchedulerStartAt,
		cacheReceptorsSchedulerInterval,
		cfg.Scheduler.MobileScheduler.Workers,
		cfg.Scheduler.MobileScheduler.QueueSize,
		cfg.Scheduler.MobileScheduler.CacheCapacity,
	)
	if err != nil {
		logger.Errorw("cache receptors start failed", "error", err)
		return nil, err
	}
	// smsir notification provider
	if cfg.Notifications.Smsir.Enabled {
		smsirSvc := smsir.NewSmsirService(
			cfg.Notifications.Smsir.ApiKey,
			cfg.Notifications.Smsir.LineNumber,
			cfg.Notifications.Smsir.Priority,
			logger,
		)
		allServices = append(allServices, smsirSvc)
		if v, err := smsirSvc.Verify(); err != nil {
			logger.Errorw("smsir verify failed", "error", err)
		} else {
			logger.Infow("smsir verified", "response", v)
		}
	} else {
		deactiveProviders = append(deactiveProviders, "Smsir")
	}

	// kavenegar notification provider
	if cfg.Notifications.Kavenegar.Enabled {
		kv := kavenegar.NewKavenegarService(
			cfg.Notifications.Kavenegar.ApiToken,
			cfg.Notifications.Kavenegar.Priority,
			cfg.Notifications.Kavenegar.Sender,
			logger,
		)
		allServices = append(allServices, kv)
		if v, err := kv.Verify(); err != nil {
			logger.Errorw("kavenegar verify failed", "error", err)
		} else {
			logger.Infow("kavenegar verified", "response", v)
		}
	} else {
		deactiveProviders = append(deactiveProviders, "Kavenegar")
	}

	// telegram notification provider
	if cfg.Notifications.Telegram.Enabled {
		telegramSvc, err := telegram.NewTelegramService(
			cfg.Notifications.Telegram.BotToken,
			cfg.Notifications.Telegram.Proxy,
			logger,
		)
		if err != nil {
			logger.Errorw("telegram init failed", "error", err)
		} else {
			allServices = append(allServices, telegramSvc)
		}
	} else {
		deactiveProviders = append(deactiveProviders, "Telegram")
	}

	// mail server provider initialized
	if cfg.Notifications.Mail.Enabled {
		mailConfig := mail.Config{
			Username:    cfg.Notifications.Mail.Username,
			Password:    cfg.Notifications.Mail.Password,
			FromAddress: cfg.Notifications.Mail.FromAddress,
			FromName:    cfg.Notifications.Mail.FromName,
			SMTPServer:  cfg.Notifications.Mail.SMTPHost,
		}

		mailServer := mail.NewService(mailConfig, "Mail", 10, logger)
		if mailServer == nil {
			logger.Errorw("mail server init failed")
		} else {
			logger.Infow("mail server initialized")
			allServices = append(allServices, mailServer)
		}
	} else {
		deactiveProviders = append(deactiveProviders, "Mail")
	}

	// Initialize mattermost notification provider
	if cfg.Notifications.Mattermost.Enabled {
		logger.Infow("Initializing mattermost notification provider",
			"bot token", cfg.Notifications.Mattermost.BotToken, "url", cfg.Notifications.Mattermost.Url)
		mattermostSvc := mattermost.NewService(
			mattermost.Config{
				Url:      cfg.Notifications.Mattermost.Url,
				BotToken: cfg.Notifications.Mattermost.BotToken,
				Priority: cfg.Notifications.Mattermost.Priority,
			},
			logger,
		)
		allServices = append(allServices, mattermostSvc)
	} else {
		deactiveProviders = append(deactiveProviders, "Mattermost")
	}

	// Initialize asiatech notification provider
	if cfg.Notifications.Asiatech.Enabled {
		asiatechCache := cache.New[string, string](logger, cache.WithCapacity(1))
		asiatechSvc := asiatech.NewService(
			cfg.Notifications.Asiatech.Username,
			cfg.Notifications.Asiatech.Password,
			cfg.Notifications.Asiatech.Scope,
			cfg.Notifications.Asiatech.Host,
			cfg.Notifications.Asiatech.Sender,
			cfg.Notifications.Asiatech.Priority,
			asiatechCache,
			logger,
		)
		allServices = append(allServices, asiatechSvc)
		if v, err := asiatechSvc.Verify(); err != nil {
			logger.Errorw("asiatech verify failed", "error", err)
		} else {
			logger.Infow("asiatech verified", "response", v)
		}
	} else {
		deactiveProviders = append(deactiveProviders, "Asiatech")
	}

	// provider registry
	providerCache := cache.New[string, *[]notifications.Providers](logger, cache.WithCapacity(3))
	providerService := notifications.NewProvidersService(repos.Postgres, allServices, providerCache, logger)
	for _, p := range allServices {
		id, err := util.NewUUIDv7()
		if err != nil {
			return nil, err
		}
		if err := providerService.AddProvider(&notifications.Providers{
			ID:          id,
			Name:        p.GetName(),
			Description: fmt.Sprintf("%s provider", p.GetName()),
			Flag:        p.GetFlag(),
			Priority:    p.GetPriority(),
			Provider:    p,
			Status:      true,
		}); err != nil {
			logger.Errorw("provider add failed", "provider", p.GetName(), "error", err)
		} else {
			logger.Infow("provider added", "provider", p.GetName())
		}
	}
	// Disable deactive providers
	for _, p := range deactiveProviders {
		err := providerService.DisableProvider(p)
		if err != nil {
			logger.Errorw("disable provider failed", "provider", p, "error", err)
			panic(err)
		}
	}

	messageService := message.NewService(repos.Postgres, logger)
	alertSchedulerInterval, err := time.ParseDuration(cfg.Scheduler.AlertScheduler.Interval)
	if err != nil {
		return nil, fmt.Errorf("incorrect alert scheduler config: %w", err)
	}
	alertCache := cache.New[string, []string](logger, cache.WithCapacity(3))
	err = schedulers.StartAlertScheduler(logger,
		repos.Postgres,
		cr,
		alertCache,
		providerService,
		messageService,
		alertSchedulerInterval,
		cfg.Scheduler.AlertScheduler.Workers,
		cfg.Scheduler.AlertScheduler.QueueSize)

	messageStatusSchedulerStartAt, err := time.ParseDuration(cfg.Scheduler.MessageStatus.StartAt)
	messageStatusSchedulerInterval, err := time.ParseDuration(cfg.Scheduler.MessageStatus.Interval)
	if err != nil {
		return nil, fmt.Errorf("incorrect message status scheduler config: %w", err)
	}
	messageStatusConfig := message_status.Config{
		StartAt:   messageStatusSchedulerStartAt,
		Interval:  messageStatusSchedulerInterval,
		Workers:   10,
		QueueSize: 100,
	}
	messageStatusScheduler, err := message_status.NewMessageStatusMessageService(
		messageService,
		providerService,
		messageStatusConfig,
		logger)
	if err != nil {
		return nil, fmt.Errorf("message status scheduler init: %w", err)
	}
	err = messageStatusScheduler.Start()
	if err != nil {
		return nil, fmt.Errorf("message status scheduler start: %w", err)
	}

	// HTTP router (and default data bootstraps like roles/admin)
	router := server.RegisterRoutes(server.Deps{
		Logger:          logger,
		Repos:           repos.Postgres,
		JWTSecret:       []byte(cfg.JwtSecret),
		SignupEnabled:   cfg.SignupEnabled,
		ProviderService: providerService,
		AdminPass:       cfg.HTTP.AdminPass,
		GinMode:         cfg.Go.Mode, // reuse
	})

	return &App{
		Logger:          logger,
		Repos:           repos,
		ProviderService: providerService,
		Services:        allServices,
		Router:          router,
	}, nil
}
