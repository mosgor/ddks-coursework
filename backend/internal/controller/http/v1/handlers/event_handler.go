package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/mosgor/Evently/backend/internal/entity"
	"github.com/mosgor/Evently/backend/internal/usecase"
)

type EventHandler struct {
	eventUseCase usecase.EventUseCase
	logger       *slog.Logger
}

type getEventsResponse struct {
	Events []*entity.Event `json:"events"`
}

// @Summary Получить список мероприятий
// @Description Возвращает список всех доступных мероприятий
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} getEventsResponse "Успешный ответ со списком мероприятий"
// @Failure 404 {string} string "Мероприятия не найдены"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /events [get]
func (h *EventHandler) GetEventsHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "events.handlers.NewGetEvents"
		logger := h.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		events, err := h.eventUseCase.GetEvents(ctx)
		if errors.Is(err, entity.ErrorNotFound) {
			logger.Error("no events found: ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			logger.Error("error getting events: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		pResponse := getEventsResponse{
			events,
		}

		logger.Info("successfully processed get events request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, pResponse)
	}
}

func (h *EventHandler) GetRecsHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "events.handlers.NewGetRecommendations"
		logger := h.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		_, claims, _ := jwtauth.FromContext(r.Context())
		if claims == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		uid := int(claims["user_id"].(float64))
		limit := 5

		events, err := h.eventUseCase.GetRecommended(ctx, uid, limit)
		if err != nil {
			logger.Error("recs failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		render.JSON(w, r, map[string]any{"events": events})
	}
}

func NewEventHandler(eventUseCase usecase.EventUseCase, logger *slog.Logger) *EventHandler {
	return &EventHandler{
		eventUseCase: eventUseCase,
		logger:       logger,
	}
}
