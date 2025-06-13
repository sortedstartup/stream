package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

const MIGRATION_TABLE = "userservice_migrations"

//go:embed migrations
var migrationFiles embed.FS

func MigrateDB(driver, url string) error {
	slog.Info("Starting database migration", "driver", driver, "url", url)

	// Create source driver from embedded files
	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create source driver: %w", err)
	}

	// Open database connection
	sqlDB, err := sql.Open(driver, url)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create database driver
	dbInstance, err := sqlite.WithInstance(sqlDB, &sqlite.Config{MigrationsTable: MIGRATION_TABLE})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "DUMMY", dbInstance)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migration
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration: %w", err)
	}

	slog.Info("Database migration completed successfully")
	return nil
}
