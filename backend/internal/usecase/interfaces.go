package usecase

import (
	"context"

	"github.com/mosgor/Evently/backend/internal/entity"
)

type EventRepository interface {
	GetAllEvents(ctx context.Context) ([]*entity.Event, error)
	GetRecommendedByVector(ctx context.Context, userVec []float32, userID, limit int) ([]*entity.Event, error)
	GetUserPreferenceVector(ctx context.Context, userID int) ([]float32, error)
}

type UserRepository interface {
	GetUser(ctx context.Context, id int) (*entity.User, error)
	AddUser(ctx context.Context, user *entity.User) error
	UpdateUser(ctx context.Context, user *entity.User) error
	GetByMail(ctx context.Context, email, password string) (*entity.User, error)
}

type TicketsRepository interface {
	GetTickets(ctx context.Context, userId int) ([]*entity.Event, error)
	AddTicket(ctx context.Context, userId, eventId int) error
}

type CartRepository interface {
	GetCartItems(ctx context.Context, userId int) ([]*entity.Event, error)
	AddCartItem(ctx context.Context, userId, eventId int) error
	RemoveCartItem(ctx context.Context, userId, eventId int) error
}

type BehaviorRepository interface {
	LogInteraction(ctx context.Context, userID, eventID int, eventType string) error
}

type Recommender interface {
	RunInference(ctx context.Context, inputVec []float32) ([]float32, error)
}
