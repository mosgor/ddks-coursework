package usecase

import (
	"context"
)

type BehaviorUseCase interface {
	LogInteraction(ctx context.Context, userID, eventID int, eventType string) error
}

type behaviorUseCase struct{ repo BehaviorRepository }

func NewBehaviorUseCase(repo BehaviorRepository) BehaviorUseCase {
	return &behaviorUseCase{repo: repo}
}

func (uc *behaviorUseCase) LogInteraction(ctx context.Context, userID, eventID int, eventType string) error {
	return uc.repo.LogInteraction(ctx, userID, eventID, eventType)
}
