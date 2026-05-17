package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/mosgor/Evently/backend/internal/usecase"
)

type TicketHandler struct {
	ticketUseCase usecase.TicketUseCase
	logger        *slog.Logger
}

// @Summary Получить билеты пользователя
// @Description Возвращает список всех билетов текущего аутентифицированного пользователя
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.Event "Успешный ответ со списком билетов"
// @Failure 401 {string} string "Неавторизованный доступ"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /tickets [get]
func (h *TicketHandler) GetTicketsHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "users.handlers.NewGetTickets"
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
		tickets, err := h.ticketUseCase.GetTickets(ctx, idFromClaims)
		if err != nil {
			logger.Error("error getting cart items: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("successfully processed get tickets items request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, tickets)
	}
}

func NewTicketHandler(ticketUseCase usecase.TicketUseCase, logger *slog.Logger) *TicketHandler {
	return &TicketHandler{
		ticketUseCase: ticketUseCase,
		logger:        logger,
	}
}
