package app

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	appconfig "github.com/ww1489/WarSpark/internal/config"
	inframysql "github.com/ww1489/WarSpark/internal/infra/mysql"
)

type MigrationOptions struct {
	ConfigPath     string
	MigrationsPath string
}

type MigrationVersion struct {
	Version uint
	Dirty   bool
}

func MigrateUp(opts MigrationOptions) error {
	migrator, err := newMigrator(opts)
	if err != nil {
		return err
	}
	defer closeMigrator(migrator)

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run up migrations: %w", err)
	}
	return nil
}

func MigrateDown(opts MigrationOptions, steps int) error {
	migrator, err := newMigrator(opts)
	if err != nil {
		return err
	}
	defer closeMigrator(migrator)

	if steps <= 0 {
		steps = 1
	}
	if err := migrator.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run down migrations: %w", err)
	}
	return nil
}

func MigrateForce(opts MigrationOptions, version int) error {
	migrator, err := newMigrator(opts)
	if err != nil {
		return err
	}
	defer closeMigrator(migrator)

	if err := migrator.Force(version); err != nil {
		return fmt.Errorf("force migration version: %w", err)
	}
	return nil
}

func CurrentMigrationVersion(opts MigrationOptions) (MigrationVersion, error) {
	migrator, err := newMigrator(opts)
	if err != nil {
		return MigrationVersion{}, err
	}
	defer closeMigrator(migrator)

	version, dirty, err := migrator.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return MigrationVersion{}, nil
	}
	if err != nil {
		return MigrationVersion{}, fmt.Errorf("read migration version: %w", err)
	}
	return MigrationVersion{Version: version, Dirty: dirty}, nil
}

func newMigrator(opts MigrationOptions) (*migrate.Migrate, error) {
	if opts.MigrationsPath == "" {
		opts.MigrationsPath = "migrations"
	}

	cfg, err := appconfig.Load(opts.ConfigPath, 0)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	sourceURL := "file://" + filepath.ToSlash(opts.MigrationsPath)
	migrator, err := migrate.New(sourceURL, inframysql.MigrationURL(cfg.MySQL))
	if err != nil {
		return nil, fmt.Errorf("init migrator: %w", err)
	}
	return migrator, nil
}

func closeMigrator(migrator *migrate.Migrate) {
	sourceErr, databaseErr := migrator.Close()
	_ = sourceErr
	_ = databaseErr
}
