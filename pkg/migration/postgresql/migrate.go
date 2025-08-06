package migrationpostgresql

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/root-ali/iris/pkg/storage/postgresql"
	"go.uber.org/zap"
)

func NewPostgresMigrate(l *zap.SugaredLogger, p *postgresql.Postgres) *MigrationPostgresql {
	return &MigrationPostgresql{
		logger: l,
		p:      p,
	}
}

func (mp *MigrationPostgresql) Migrate() error {
	db, err := sql.Open("postgres",
		"postgres://"+mp.p.User+":"+
			mp.p.Password+"@"+mp.p.Host+":"+mp.p.Port+"/"+mp.p.DBname+"?sslmode=disable")
	if err != nil {
		mp.logger.Panicw("Cannot connect to postgres database for migration",
			"Error", err)
	}
	defer db.Close()
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		mp.logger.Panicw("Cannot connect to postgres database for migrations",
			"Error", err)
	}

	// Get migrations path - try environment variable first, then relative path
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		// Default to migrations directory in project root
		wd, _ := os.Getwd()
		migrationsPath = filepath.Join(wd, "migrations")
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		mp.logger.Panicw("failed to create migration instance", "Error", err)
	}
	err = m.Up()
	mp.logger.Infow("migrate finished", "err ", err)
	if errors.Is(err, migrate.ErrNoChange) || errors.Is(err, nil) {
		version, _, _ := m.Version()
		mp.logger.Infow("Everything is up to date let's continue", "Version", version)
	} else {
		mp.logger.Errorw("Something went wrong while migrating postgresql ", "Error", err)
		return err
	}
	return nil
}
