package usecase

import (
	"context"

	"github.com/mosgor/Evently/backend/internal/entity"
)

type CartUseCase interface {
	GetCartItems(ctx context.Context, userId int) ([]*entity.Event, error)
	AddCartItem(ctx context.Context, userId, eventId int) error
	RemoveCartItem(ctx context.Context, userId, eventId int) error
}

type cartUseCase struct {
	cartRepo CartRepository
}

func NewCartUseCase(cartRepo CartRepository) CartUseCase {
	return &cartUseCase{
		cartRepo: cartRepo,
	}
}

func (uc *cartUseCase) GetCartItems(ctx context.Context, userId int) ([]*entity.Event, error) {
	return uc.cartRepo.GetCartItems(ctx, userId)
}

func (uc *cartUseCase) AddCartItem(ctx context.Context, userId, eventId int) error {
	return uc.cartRepo.AddCartItem(ctx, userId, eventId)
}

func (uc *cartUseCase) RemoveCartItem(ctx context.Context, userId, eventId int) error {
	return uc.cartRepo.RemoveCartItem(ctx, userId, eventId)
}
