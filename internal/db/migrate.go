package db

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

// RunMigrations запускает goose-миграции из каталога migrationsDir.
// Ожидает, что миграции в формате SQL.
func RunMigrations(ctx context.Context, database *sqlx.DB, migrationsDir string) error {
	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(nil)
	goose.SetDialect("postgres")

	// нормализуем путь (на случай относительных путей)
	abs, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("migrations path: %w", err)
	}
	// таймаут на прогон миграций
	mctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	ch := make(chan error, 1)
	go func() {
		ch <- goose.Up(database.DB, abs)
	}()
	select {
	case <-mctx.Done():
		return fmt.Errorf("migrations timeout: %w", mctx.Err())
	case err := <-ch:
		return err
	}
}
