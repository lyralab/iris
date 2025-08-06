package postgresql

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Postgres struct {
	Host     string
	User     string
	Password string
	DBname   string
	Port     string
	SSLMode  bool
}

type Storage struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func New(l *zap.SugaredLogger, p *Postgres) *Storage {
	gormLogger := logger.New(
		NewZapWriter(l),
		logger.Config{
			SlowThreshold:             time.Second, // report queries > 1s as slow
			LogLevel:                  logger.Info, // change to logger.Silent to disable
			IgnoreRecordNotFoundError: true,        // skip ErrRecordNotFound errors
			Colorful:                  false,       // disable color
		},
	)
	storage := Storage{}
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", p.Host, p.User, p.Password, p.DBname, p.Port)

	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
	if err != nil {
		l.Panic("Failed to connect to the Database")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return nil
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Second)
	l.Info("? Connected Successfully to the Database")
	storage.db = DB
	storage.logger = l
	return &storage

}

type zapWriter struct {
	sugar *zap.SugaredLogger
}

func NewZapWriter(sugar *zap.SugaredLogger) logger.Writer {
	return &zapWriter{sugar: sugar}
}

func (w *zapWriter) Printf(message string, data ...interface{}) {
	w.sugar.Debugf(message, data...)
}
