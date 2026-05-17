package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mosgor/Evently/backend/config"
	"golang.org/x/crypto/bcrypt"
)

func CreateAdminUser(ctx context.Context, cfg config.Database, email, password string, logger *slog.Logger) error {
	pool, err := connectToDB(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (name, email, password, role) 
		VALUES ($1, $2, $3, 'admin')
		ON CONFLICT (email) DO UPDATE 
		SET role = 'admin', password = EXCLUDED.password
		RETURNING id`

	var userID int
	err = pool.QueryRow(ctx, query, "Admin", email, string(hashedPassword)).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	logger.Info("admin user created/updated", "user_id", userID, "email", email)
	return nil
}

func ClearRedisCache(ctx context.Context, cfg config.Redis, logger *slog.Logger) error {
	logger.Warn("cache clear is a destructive operation", "host", cfg.Host)
	return nil
}

func connectToDB(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	return pgxpool.New(ctx, connStr)
}
