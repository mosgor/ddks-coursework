package v1

import (
	"context"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/mosgor/Evently/backend/config"
	"github.com/mosgor/Evently/backend/internal/controller/http/v1/handlers"
	"github.com/mosgor/Evently/backend/internal/infrastracture/redis"
	"github.com/mosgor/Evently/backend/internal/usecase"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func JWTBlacklistMiddleware(redisClient *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token := authHeader[7:]

				if redisClient != nil {
					if blacklisted, _ := redisClient.IsTokenBlacklisted(r.Context(), token); blacklisted {
						http.Error(w, "Token revoked", http.StatusUnauthorized)
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func NewJSONLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				logger.InfoContext(r.Context(), "HTTP request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"bytes_sent", ww.BytesWritten(),
					"duration_ms", time.Since(start).Milliseconds(),
					"request_id", middleware.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

var IsShuttingDown atomic.Bool

func ShutdownMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsShuttingDown.Load() {
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"Service is shutting down"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// @title Evently API
// @version 1.0
// @description REST API для системы мероприятий Evently

// @contact.name mos_gor Support

// @host localhost:8090
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате "Bearer {token}"
func NewEventlyRouter(
	ctx context.Context, cfg config.Server, redisCfg config.Redis,
	logger *slog.Logger, eventUseCase usecase.EventUseCase,
	userUseCase usecase.UserUseCase,
	ticketUseCase usecase.TicketUseCase,
	cartUseCase usecase.CartUseCase,
	paymentUseCase usecase.PaymentUseCase,
	behaviorUseCase usecase.BehaviorUseCase,
	redisClient *redis.Client,
) *chi.Mux {
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"https://localhost:3000",
			"http://0.0.0.0:3000",
			"https://0.0.0.0:3000",
			"http://51.250.44.86:3000",
			"https://51.250.44.86:3000",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
		ExposedHeaders:   []string{"Link", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	router.Use(ShutdownMiddleware)
	router.Use(middleware.RequestID)
	router.Use(NewJSONLogger(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Timeout(cfg.Timeout))
	router.Use(middleware.RequestSize(350))

	jwtToken := jwtauth.New("HS256", []byte(cfg.JwtSecret), nil)

	eventHandler := handlers.NewEventHandler(eventUseCase, logger)
	authHandler := handlers.NewAuthHandler(userUseCase, logger, jwtToken, redisClient)
	userHandler := handlers.NewUserHandler(userUseCase, logger)
	cartHandler := handlers.NewCartHandler(cartUseCase, logger)
	ticketHandler := handlers.NewTicketHandler(ticketUseCase, logger)
	paymentHandler := handlers.NewPaymentHandler(paymentUseCase, logger)
	trackHandler := handlers.NewTrackHandler(behaviorUseCase, logger)

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8090/swagger/doc.json"),
	))

	// Events routes
	router.Route("/events", func(r chi.Router) {
		r.Get("/", eventHandler.GetEventsHandler(ctx))
	})

	// Auth routes
	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.LoginHandler(ctx))
		r.Post("/register", authHandler.RegisterHandler(ctx))
	})

	router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(jwtToken))
		r.Use(jwtauth.Authenticator(jwtToken))
		r.Use(JWTBlacklistMiddleware(redisClient))

		// User routes
		r.Get("/auth/me", userHandler.GetUserHandler(ctx))
		r.Put("/auth/me", userHandler.UpdateUserHandler(ctx))
		r.Post("/auth/logout", authHandler.LogoutHandler(ctx))
		r.Get("/recommendations", eventHandler.GetRecsHandler(ctx))

		// Cart routes
		r.Route("/cart", func(r chi.Router) {
			r.Get("/", cartHandler.GetCartItemsHandler(ctx))
			r.Post("/", cartHandler.AddCartItemHandler(ctx))
			r.Delete("/{id}", cartHandler.DeleteCartItemHandler(ctx))
		})

		// Payment routes
		r.Route("/payment", func(r chi.Router) {
			r.Post("/", paymentHandler.PayHandler(ctx))
		})

		// Tickets routes
		r.Route("/tickets", func(r chi.Router) {
			r.Get("/", ticketHandler.GetTicketsHandler(ctx))
		})

		r.Post("/track", trackHandler.TrackHandler(ctx))
	})

	return router
}
