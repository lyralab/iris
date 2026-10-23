package main

import (
	"log"
	"os"

	"github.com/root-ali/iris/internal/bootstrap"
	"github.com/root-ali/iris/internal/config"
)

func main() {

	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if configFilePath == "" {
		configFilePath = "./config.yml"
	}
	cfg := config.New(configFilePath)
	app, err := bootstrap.Init(cfg)
	if err != nil {
		log.Fatalf("bootstrap init: %v", err)
	}

	// Run HTTP

	if err := app.Router.Run("0.0.0.0:" + cfg.Server.Port); err != nil {
		app.Logger.Fatalw("http server failed", "error", err)
	}
}
