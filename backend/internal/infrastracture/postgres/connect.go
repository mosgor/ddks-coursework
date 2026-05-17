package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mosgor/Evently/backend/config"
)

type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	logger *slog.Logger
	Pool   *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Database, l *slog.Logger) (*Postgres, error) {

	pg := &Postgres{
		maxPoolSize:  cfg.MaxPoolSize,
		connAttempts: cfg.ConnAttempts,
		connTimeout:  cfg.Timeout,
		logger:       l,
	}

	poolConfig, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	))
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %s", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			return nil, fmt.Errorf("error with config while creating new postgres pool: %s", err)
		}
		err = pg.Pool.Ping(ctx)
		if err == nil {
			break
		}

		// l.Warn("Postgres is trying to connect, attempts left: " + strconv.Itoa(pg.connAttempts))

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("can't connect to database: %s", err)
	}

	return pg, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
		p.logger.Info("postgres pool closed")
	}
}
