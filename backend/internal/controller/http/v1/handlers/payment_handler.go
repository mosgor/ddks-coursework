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

type PaymentHandler struct {
	paymentUseCase usecase.PaymentUseCase
	logger         *slog.Logger
}

type payRequest struct {
	Ids []int `json:"eventIds"`
}

type payResponse struct {
	Success bool `json:"success"`
}

// @Summary Обработать оплату
// @Description Обрабатывает оплату выбранных мероприятий для текущего пользователя
// @Tags payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment body payRequest true "Данные для оплаты"
// @Success 200 {object} payResponse "Успешная обработка оплаты"
// @Failure 400 {string} string "Некорректные данные запроса"
// @Failure 401 {string} string "Неавторизованный доступ"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /payment [post]
func (h *PaymentHandler) PayHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "users.handlers.NewPay"
		logger := h.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req payRequest
		var ids []int
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			logger.Error("error decoding ids: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ids = req.Ids
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
		err = h.paymentUseCase.ProcessPayment(ctx, idFromClaims, ids)
		if err != nil {
			logger.Error("error adding tickets: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("successfully processed add tickets request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, payResponse{true})
	}
}

func NewPaymentHandler(paymentUseCase usecase.PaymentUseCase, logger *slog.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentUseCase: paymentUseCase,
		logger:         logger,
	}
}
