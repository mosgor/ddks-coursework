package usecase

import (
	"context"
	"log/slog"

	"github.com/mosgor/Evently/backend/internal/entity"
)

type EventUseCase interface {
	GetEvents(ctx context.Context) ([]*entity.Event, error)
	GetRecommended(ctx context.Context, userID, limit int) ([]*entity.Event, error)
}

type eventUseCase struct {
	eventsRepo  EventRepository
	recommender Recommender
	logger      *slog.Logger
}

func NewEventUseCase(eventsRepo EventRepository, recommender Recommender, logger *slog.Logger) EventUseCase {
	return &eventUseCase{
		eventsRepo:  eventsRepo,
		recommender: recommender,
		logger:      logger,
	}
}

func (uc *eventUseCase) GetEvents(ctx context.Context) ([]*entity.Event, error) {
	return uc.eventsRepo.GetAllEvents(ctx)
}

func (uc *eventUseCase) GetRecommended(ctx context.Context, userID, limit int) ([]*entity.Event, error) {
	avgVecF64, err := uc.eventsRepo.GetUserPreferenceVector(ctx, userID)
	if err != nil {
		return nil, err
	}
	if avgVecF64 == nil {
		return []*entity.Event{}, nil
	}

	avgVec := make([]float32, len(avgVecF64))
	for i, v := range avgVecF64 {
		avgVec[i] = float32(v)
	}

	userVec := avgVec

	// if uc.recommender != nil {
	// 	refined, err := uc.recommender.RunInference(ctx, avgVec)
	// 	if err != nil {
	// 		uc.logger.Warn("recommender inference failed, falling back to avg vector", "error", err)
	// 	} else if refined != nil {
	// 		userVec = refined
	// 		uc.logger.Debug("used refined user vector from ONNX")
	// 	}
	// }

	uc.logger.Info(
		"recommendation vectors",
		"userID", userID,
		"avg", avgVec[:5],
		"final", userVec[:5],
	)

	return uc.eventsRepo.GetRecommendedByVector(ctx, userVec, userID, limit)
}
