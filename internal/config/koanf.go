package config

import (
	"log"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

func New(path string) *Config {
	var instance Config

	k := koanf.New(".")

	// load configuration from file
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		log.Printf("error loading config.yml: %s", err)
	}

	// Load env configurations with prefix "IRIS_". For example, "IRIS_SERVER_PORT" will be mapped to "server.port" in the config struct.
	err := k.Load(env.Provider("IRIS_", ".", func(s string) string {
		s = strings.TrimPrefix(s, "APP_")
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "_", ".")
		return s
	}), nil)
	if err != nil {
		log.Printf("error loading environment variables: %s", err)
	}

	if err := k.Unmarshal("iris", &instance); err != nil {
		log.Fatalf("error unmarshalling config: %s", err)
	}

	return &instance
}
