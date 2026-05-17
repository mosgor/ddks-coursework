package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/mosgor/Evently/backend/internal/usecase"
)

type CartHandler struct {
	cartUseCase usecase.CartUseCase
	logger      *slog.Logger
}

type deleteCartItemResponse struct {
	Success bool `json:"success"`
}

// @Summary Удалить элемент из корзины
// @Description Удаляет мероприятие из корзины текущего пользователя по ID
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID элемента корзины"
// @Success 200 {object} deleteCartItemResponse "Успешное удаление элемента"
// @Failure 400 {string} string "Некорректный ID"
// @Failure 401 {string} string "Неавторизованный доступ"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /cart/{id} [delete]
func (h *CartHandler) DeleteCartItemHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "users.handlers.NewDeleteCartItem"
		logger := h.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			logger.Error("error getting id")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Error("error converting id to int: ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
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
		idFromClaims := int(claims["user_id"].(float64))
		err = h.cartUseCase.RemoveCartItem(ctx, idFromClaims, id)
		if err != nil {
			logger.Error("error deleting cart item: ", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		logger.Info("successfully processed delete cart item request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, deleteCartItemResponse{true})
	}
}

type addCartItemRequest struct {
	Id int `json:"eventId"`
}

type addCartItemResponse struct {
	Success bool `json:"success"`
	ItemId  int  `json:"item_id"`
}

// @Summary Добавить элемент в корзину
// @Description Добавляет мероприятие в корзину текущего пользователя
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cartItem body addCartItemRequest true "Данные для добавления в корзину"
// @Success 200 {object} addCartItemResponse "Успешное добавление элемента"
// @Failure 400 {string} string "Некорректные данные запроса"
// @Failure 401 {string} string "Неавторизованный доступ"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /cart [post]
func (h *CartHandler) AddCartItemHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "users.handlers.NewAddCartItem"
		logger := h.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req addCartItemRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			logger.Error("error decoding ids: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
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
		idFromClaims := int(claims["user_id"].(float64))
		err = h.cartUseCase.AddCartItem(ctx, idFromClaims, req.Id)
		if err != nil {
			logger.Error("error adding cart item: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("successfully processed add cart item request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, addCartItemResponse{true, req.Id})
	}
}

// @Summary Получить элементы корзины
// @Description Возвращает все элементы корзины текущего аутентифицированного пользователя
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.Event "Успешный ответ со списком элементов корзины"
// @Failure 401 {string} string "Неавторизованный доступ"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /cart [get]
func (h *CartHandler) GetCartItemsHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "users.handlers.NewGetCartItems"
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
		cartItems, err := h.cartUseCase.GetCartItems(ctx, idFromClaims)
		if err != nil {
			logger.Error("error getting cart items: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("successfully processed get cart items request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, cartItems)
	}
}

func NewCartHandler(cartUseCase usecase.CartUseCase, logger *slog.Logger) *CartHandler {
	return &CartHandler{
		cartUseCase: cartUseCase,
		logger:      logger,
	}
}
