package postgres

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	pgxDriver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// RunMigrations executes all pending database migrations.
// It uses PostgreSQL advisory locks to ensure only one instance runs migrations at a time.
// Returns an error if migrations fail - caller should treat this as fatal.
func RunMigrations(pool *pgxpool.Pool, migrationsFS embed.FS, migrationsPath string) error {
	slog.Info("starting database migrations", "path", migrationsPath)

	// Create sub filesystem from embedded FS
	subFS, err := fs.Sub(migrationsFS, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to create sub filesystem: %w", err)
	}

	// Create source from embedded filesystem
	source, err := iofs.New(subFS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Convert pgxpool.Pool to *sql.DB for golang-migrate
	db := stdlib.OpenDBFromPool(pool)

	// Create pgx driver with default config
	driver, err := pgxDriver.WithInstance(db, &pgxDriver.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrator instance
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	// Get current version before migration
	currentVersion, dirty, _ := m.Version()
	if dirty {
		slog.Error("database is in dirty state",
			"version", currentVersion,
			"hint", "manually fix the issue and run: migrate force <version>",
		)
		return fmt.Errorf("database is in dirty state at version %d", currentVersion)
	}

	// Run all pending migrations
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			slog.Info("database schema is up to date", "version", currentVersion)
			return nil
		}

		// Log details for debugging
		version, dirty, _ := m.Version()
		slog.Error("migration failed",
			"error", err,
			"current_version", version,
			"dirty", dirty,
		)
		return fmt.Errorf("migration failed: %w", err)
	}

	// Log success
	newVersion, _, _ := m.Version()
	slog.Info("migrations completed successfully",
		"previous_version", currentVersion,
		"current_version", newVersion,
	)

	return nil
}
