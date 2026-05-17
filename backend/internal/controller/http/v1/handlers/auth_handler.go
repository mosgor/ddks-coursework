package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/jwtauth/v5"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/mosgor/Evently/backend/internal/entity"
	"github.com/mosgor/Evently/backend/internal/infrastracture/redis"
	"github.com/mosgor/Evently/backend/internal/usecase"
)

type AuthHandler struct {
	userUseCase usecase.UserUseCase
	logger      *slog.Logger
	jwtAuth     *jwtauth.JWTAuth
	redisClient *redis.Client
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	User  entity.User `json:"user"`
	Token string      `json:"token"`
}

// @Summary Вход в систему
// @Description Аутентифицирует пользователя и возвращает JWT-токен
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body loginRequest true "Учетные данные пользователя"
// @Success 200 {object} loginResponse "Успешная аутентификация"
// @Failure 400 {string} string "Некорректные данные запроса"
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *AuthHandler) LoginHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "users.handlers.NewLogin"
		logger := h.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var request loginRequest
		if err := render.DecodeJSON(r.Body, &request); err != nil {
			logger.Error("error decoding login request: ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user, err := h.userUseCase.Login(ctx, request.Email, request.Password)
		if errors.Is(err, entity.ErrorNotFound) {
			logger.Error("no user found: ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			logger.Error("error getting user: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, tokenString, err := h.jwtAuth.Encode(map[string]interface{}{"user_id": user.Id})
		if err != nil {
			logger.Error("error encoding token: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Info("successfully processed get user request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, loginResponse{
			*user,
			tokenString,
		})
	}
}

type registerResponse struct {
	User  entity.User `json:"user"`
	Token string      `json:"token"`
}

// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя и возвращает JWT-токен
// @Tags auth
// @Accept json
// @Produce json
// @Param user body entity.User true "Данные нового пользователя"
// @Success 200 {object} registerResponse "Успешная регистрация"
// @Failure 400 {string} string "Некорректные данные запроса"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /auth/register [post]
func (h *AuthHandler) RegisterHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "users.handlers.NewRegister"
		logger := h.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var user entity.User
		if err := render.DecodeJSON(r.Body, &user); err != nil {
			logger.Error("error decoding register request: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := h.userUseCase.Register(ctx, &user)
		if err != nil {
			logger.Error("error adding user: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, tokenString, err := h.jwtAuth.Encode(map[string]interface{}{"user_id": user.Id})
		if err != nil {
			logger.Error("error encoding token: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Info("successfully processed add user request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, registerResponse{
			user,
			tokenString,
		})
	}
}

// @Summary Выход из системы
// @Description Добавляет текущий JWT-токен в чёрный список
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {string} string "Успешный выход"
// @Router /auth/logout [post]
func (h *AuthHandler) LogoutHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handlers.Logout"
		logger := h.logger.With(slog.String("op", op))

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
			logger.Error("no bearer token in request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		token := authHeader[7:]

		if h.redisClient != nil {
			if err := h.redisClient.BlacklistToken(ctx, token); err != nil {
				logger.Error("failed to blacklist token", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		logger.Info("token blacklisted, user logged out")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]string{"message": "logged out"})
	}
}

func NewAuthHandler(userUseCase usecase.UserUseCase, logger *slog.Logger, jwtAuth *jwtauth.JWTAuth, redisClient *redis.Client) *AuthHandler {
	return &AuthHandler{
		userUseCase: userUseCase,
		logger:      logger,
		jwtAuth:     jwtAuth,
		redisClient: redisClient,
	}
}
