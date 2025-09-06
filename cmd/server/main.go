package main

import (
	"log"

	"github.com/root-ali/iris/internal/bootstrap"
	"github.com/root-ali/iris/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	app, err := bootstrap.Init(cfg)
	if err != nil {
		log.Fatalf("bootstrap init: %v", err)
	}

	// Run HTTP
	if err := app.Router.Run(":" + cfg.HTTP.Port); err != nil {
		app.Logger.Fatalw("http server failed", "error", err)
	}
}
