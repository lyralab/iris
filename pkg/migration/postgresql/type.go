package migrationpostgresql

import (
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"go.uber.org/zap"
)

type MigrationPostgresql struct {
	logger *zap.SugaredLogger
	p      *postgresql.Postgres
}
