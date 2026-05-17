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
	"github.com/mosgor/Evently/backend/internal/usecase"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
	logger      *slog.Logger
}

// @Summary Получить данные текущего пользователя
// @Description Возвращает информацию о текущем аутентифицированном пользователе
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.User "Успешный ответ с данными пользователя"
// @Failure 401 {string} string "Неавторизованный доступ"
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /auth/me [get]
func (h *UserHandler) GetUserHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "users.handlers.NewGetUser"
		logger := h.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		_, claims, err :=
			jwtauth.FromContext(r.Context())
		if err != nil {
			logger.Error("error getting claims: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if claims == nil {
			logger.Error("no claims found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		idFromClaims := int(claims["user_id"].(float64))
		user, err := h.userUseCase.GetUser(ctx, idFromClaims)
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

		logger.Info("successfully processed get user request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, user)
	}
}

// @Summary Обновить данные пользователя
// @Description Обновляет информацию о текущем аутентифицированном пользователе
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body entity.User true "Данные пользователя для обновления"
// @Success 200 {object} entity.User "Успешное обновление пользователя"
// @Failure 400 {string} string "Некорректные данные запроса"
// @Failure 401 {string} string "Неавторизованный доступ"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /auth/me [put]
func (h *UserHandler) UpdateUserHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "users.handlers.NewUpdateUser"
		logger := h.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var user entity.User
		if err := render.DecodeJSON(r.Body, &user); err != nil {
			logger.Error("error decoding user: ", err)
		}
		_, claims, err :=
			jwtauth.FromContext(r.Context())
		if err != nil {
			logger.Error("error getting claims: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if claims == nil {
			logger.Error("no claims found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		user.Id = int(claims["user_id"].(float64))
		err = h.userUseCase.UpdateUser(ctx, &user)
		if err != nil {
			logger.Error("error updating user: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("successfully processed update user request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, user)
	}
}

func NewUserHandler(userUseCase usecase.UserUseCase, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
		logger:      logger,
	}
}
