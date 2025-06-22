package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"

	_ "modernc.org/sqlite"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const MIGRATION_TABLE = "userservice_migrations"

//go:embed migrations
var migrationFiles embed.FS

func MigrateDB(driver string, dbURL string) error {
	_migrationFiles, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Migrating database", "dbURL", dbURL)

	sqlDB, err := sql.Open(driver, dbURL)
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}
	dbInstance, err := sqlite.WithInstance(sqlDB, &sqlite.Config{MigrationsTable: MIGRATION_TABLE})
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}

	//TODO: externalize in config
	m, err := migrate.NewWithInstance("iofs", _migrationFiles, "DUMMY", dbInstance)

	if err != nil {
		slog.Error("error", "err", err)
		return fmt.Errorf("failed creating new migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("error", "err", err)
		return fmt.Errorf("failed while migrating: %w", err)
	}

	return nil
}
