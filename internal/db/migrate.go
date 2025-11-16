package db

import (
	"fmt"
	"pr-reviewer-service/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(cfg *config.Config) error {
	m, err := migrate.New(
		cfg.DB.MigrationsPath,
		cfg.DB.URL,
	)
	if err != nil {
		return fmt.Errorf("migrate: failed to create migrations: %w", err)
	}

	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate: failed to run migrations: %w", err)
	}

	return nil
}
