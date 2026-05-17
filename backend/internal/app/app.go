package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/mosgor/Evently/backend/config"
	v1 "github.com/mosgor/Evently/backend/internal/controller/http/v1"
	"github.com/mosgor/Evently/backend/internal/infrastracture/postgres"
	"github.com/mosgor/Evently/backend/internal/infrastracture/recommender"
	"github.com/mosgor/Evently/backend/internal/infrastracture/redis"
	"github.com/mosgor/Evently/backend/internal/usecase"
)

func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger.Info("Trying to connect to PostgreSQL as " + cfg.Database.User + " user")
	pg, err := postgres.New(ctx, cfg.Database, logger)
	if err != nil {
		logger.Info("Failed to connect to PostgreSQL switching to SQLite")
		// log.Fatalf("error while connecting to the database: %s", err)
	}
	defer pg.Close()

	logger.Info("Trying to connect to Redis")
	redisClient, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		logger.Warn("Failed to connect to Redis, continuing without it", "error", err)
	}
	defer redisClient.Close()

	embeddingDimStr := os.Getenv("EMBEDDING_DIM")
	var embeddingDim int64 = 384
	if d, err := strconv.ParseInt(embeddingDimStr, 10, 64); err == nil && d > 0 {
		embeddingDim = d
	}

	modelPath := os.Getenv("REC_MODEL_PATH")
	recService, err := recommender.New(modelPath, logger, embeddingDim)
	if err != nil {
		logger.Warn("recommender init failed, using pure PostgreSQL fallback", "error", err)
		recService = nil
	}
	if recService != nil {
		defer recService.Close()
	}

	eventsRepo := postgres.NewEventPostgresRepo(pg)
	userRepo := postgres.NewUserRepository(pg)
	ticketRepo := postgres.NewTicketsRepository(pg)
	cartRepo := postgres.NewCartItemsRepository(pg)
	behaviorRepo := postgres.NewBehaviorRepository(pg)

	eventUseCase := usecase.NewEventUseCase(eventsRepo, recService, logger)
	userUseCase := usecase.NewUserUseCase(userRepo)
	ticketUseCase := usecase.NewTicketUseCase(ticketRepo)
	cartUseCase := usecase.NewCartUseCase(cartRepo)
	paymentUseCase := usecase.NewPaymentUseCase(cartRepo, ticketRepo)
	behaviorUseCase := usecase.NewBehaviorUseCase(behaviorRepo)

	router := v1.NewEventlyRouter(
		ctx, cfg.Server, cfg.Redis,
		logger, eventUseCase,
		userUseCase, ticketUseCase,
		cartUseCase, paymentUseCase, behaviorUseCase, redisClient,
	)

	httpServer := v1.NewServer(cfg.Server, router)
	logger.Info("starting server on port " + cfg.Server.Port)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		logger.Info(s.String() + " signal received")

		v1.IsShuttingDown.Store(true)
	case err = <-httpServer.Notify():
		logger.Error("shutdown via http server error: " + err.Error())
	case <-ctx.Done():
		logger.Info("context canceled")
	}

	logger.Info("shutting down...")
	cancel()

	err = httpServer.Shutdown()
	if err != nil {
		logger.Error("error while shutting down the http server: " + err.Error())
	}
	logger.Info("shut down http server")
}
