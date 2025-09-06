package storage

import (
	migrationpostgresql "github.com/root-ali/iris/pkg/migration/postgresql"
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"go.uber.org/zap"
)

type Repos struct {
	Postgres *postgresql.Storage
}

func Init(logger *zap.SugaredLogger, pgConf postgresql.Postgres) (*Repos, error) {
	mp := migrationpostgresql.NewPostgresMigrate(logger, &pgConf)
	if err := mp.Migrate(); err != nil {
		logger.Panicw("migration failed", "error", err)
	}
	st := postgresql.New(logger, &pgConf)
	return &Repos{Postgres: st}, nil
}
