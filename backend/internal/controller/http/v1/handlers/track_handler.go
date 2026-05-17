package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/mosgor/Evently/backend/internal/usecase"
)

type TrackHandler struct {
	behaviorUC usecase.BehaviorUseCase
	logger     *slog.Logger
}

func NewTrackHandler(uc usecase.BehaviorUseCase, l *slog.Logger) *TrackHandler {
	return &TrackHandler{uc, l}
}

type trackReq struct {
	EventID int    `json:"event_id"`
	Type    string `json:"type"`
}

func (h *TrackHandler) TrackHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req trackReq
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, claims, _ := jwtauth.FromContext(r.Context())
		if claims == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		uid := int(claims["user_id"].(float64))
		if err := h.behaviorUC.LogInteraction(ctx, uid, req.EventID, req.Type); err != nil {
			h.logger.Error("track failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
