package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mosgor/Evently/backend/config"
)

func RunMigrations(ctx context.Context, cfg config.Database, logger *slog.Logger) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	migrationsPath := "file:///app/migrations"
	if _, err := os.Stat("/app/migrations"); os.IsNotExist(err) {
		migrationsPath = "file://./backend/migrations"
	}

	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("failed to init migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	logger.Info("migrations applied", "version", getMigrationVersion(m))
	return nil
}

func getMigrationVersion(m *migrate.Migrate) uint {
	v, _, _ := m.Version()
	return v
}
