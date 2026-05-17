// backend/cmd/app/main.go
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/mosgor/Evently/backend/cmd/cli"
	"github.com/mosgor/Evently/backend/config"
	_ "github.com/mosgor/Evently/backend/docs"
	"github.com/mosgor/Evently/backend/internal/app"
)

func main() {
	mode := flag.String("mode", "server", "Режим запуска: server, migrate, create-admin, clear-cache")
	email := flag.String("email", "", "Email для create-admin")
	password := flag.String("password", "", "Пароль для create-admin")
	configPath := flag.String("config", "", "Путь к конфигу")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	cfg := config.NewConfig(*configPath)

	ctx := context.Background()

	switch *mode {
	case "server":
		app.Run(ctx, cfg, logger)
	case "migrate":
		if err := cli.RunMigrations(ctx, cfg.Database, logger); err != nil {
			logger.Error("migration failed", "error", err)
			os.Exit(1)
		}
		logger.Info("migrations completed successfully")
	case "create-admin":
		if *email == "" || *password == "" {
			logger.Error("email and password are required for create-admin")
			os.Exit(1)
		}
		if err := cli.CreateAdminUser(ctx, cfg.Database, *email, *password, logger); err != nil {
			logger.Error("failed to create admin", "error", err)
			os.Exit(1)
		}
		logger.Info("admin user created", "email", *email)
	case "clear-cache":
		if err := cli.ClearRedisCache(ctx, cfg.Redis, logger); err != nil {
			logger.Error("cache clear failed", "error", err)
			os.Exit(1)
		}
		logger.Info("cache cleared successfully")
	default:
		logger.Error("unknown mode", "mode", *mode)
		os.Exit(1)
	}
}
