package database

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// RunMigrations applies all pending migrations from the provided embed.FS against the target DSN.
func RunMigrations(migrationsFS fs.FS, dsn string) error {
	driver, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("database: failed to load embedded migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", driver, dsn)
	if err != nil {
		return fmt.Errorf("database: failed to initialize migrator: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		_ = srcErr
		_ = dbErr
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("database: migration up failed: %w", err)
	}

	return nil
}

// RunMigrationsDown rolls back n migration steps from the provided embed.FS.
func RunMigrationsDown(migrationsFS fs.FS, dsn string, steps int) error {
	driver, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("database: failed to load embedded migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", driver, dsn)
	if err != nil {
		return fmt.Errorf("database: failed to initialize migrator: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		_ = srcErr
		_ = dbErr
	}()

	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("database: migration down failed: %w", err)
	}

	return nil
}
