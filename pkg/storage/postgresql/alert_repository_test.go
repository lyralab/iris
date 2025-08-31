package postgresql

import (
	"context"
	"github.com/root-ali/iris/pkg/alerts"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"sync"
	"testing"
	"time"
)

type testLogger struct{ t *testing.T }

func (l testLogger) Errorf(format string, args ...any) { l.t.Logf("ERR: "+format, args...) }

// Minimal model mapping to your existing alerts table.
// If your real struct differs, it’s fine; we only use this in test setup/queries.
type alertRow struct {
	Id        string `gorm:"column:id;primaryKey"`
	SendNotif bool   `gorm:"column:send_notif"`
	Receptor  string `gorm:"column:receptor"`
}

// TableName ensures GORM uses the exact table.
func (alertRow) TableName() string { return "alerts" }

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Ensure the table exists with needed columns (id TEXT PK, send_notif BOOL)
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			send_notif BOOLEAN NOT NULL DEFAULT false,
			receptor TEXT
		)`).
		Error; err != nil {
		t.Fatalf("migrate alerts: %v", err)
	}
	return db
}

func resetData(t *testing.T, db *gorm.DB, rows []alertRow) {
	t.Helper()
	if err := db.Exec(`TRUNCATE alerts`).Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}
	for _, r := range rows {
		if err := db.Create(&r).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func newStorage(t *testing.T, db *gorm.DB) *Storage {
	t.Helper()
	return &Storage{
		db:     db,                   // If your field is named `db`, adjust here.
		logger: zap.NewNop().Sugar(), // If your field is `logger`, adjust here.
	}
}

func TestConcurrentWorkers_ProcessQueueWithoutDuplicates(t *testing.T) {
	db := openDB(t)
	resetData(t, db, []alertRow{
		{Id: "q1", SendNotif: false},
		{Id: "q2", SendNotif: false},
		{Id: "q3", SendNotif: false},
	})
	s := newStorage(t, db)

	worker := func(wg *sync.WaitGroup) {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			var a alerts.Alert
			id, err := s.GetUnsentAlertID(a)
			if err != nil {
				// transient tx error; backoff and retry a bit
				time.Sleep(20 * time.Millisecond)
				continue
			}
			if id == "" { // nothing left
				return
			}
			if err := s.MarkAlertAsSent(id); err != nil {
				// retry path (very unlikely with proper locking)
				time.Sleep(20 * time.Millisecond)
				continue
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go worker(&wg)
	go worker(&wg)
	wg.Wait()

	// Assert every row is marked exactly once; i.e., all send_notif = true
	var cnt int64
	if err := db.Model(&alertRow{}).Where("send_notif = ?", true).Count(&cnt).Error; err != nil {
		t.Fatalf("count true: %v", err)
	}
	if cnt != 3 {
		t.Fatalf("want 3 processed rows, got %d", cnt)
	}
}
