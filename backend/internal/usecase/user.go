package usecase

import (
	"context"

	"github.com/mosgor/Evently/backend/internal/entity"
)

type UserUseCase interface {
	GetUser(ctx context.Context, id int) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, error)
	Register(ctx context.Context, user *entity.User) error
	UpdateUser(ctx context.Context, user *entity.User) error
}

type userUseCase struct {
	userRepo UserRepository
}

func NewUserUseCase(userRepo UserRepository) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
}

func (uc *userUseCase) GetUser(ctx context.Context, id int) (*entity.User, error) {
	return uc.userRepo.GetUser(ctx, id)
}

func (uc *userUseCase) Login(ctx context.Context, email, password string) (*entity.User, error) {
	return uc.userRepo.GetByMail(ctx, email, password)
}

func (uc *userUseCase) Register(ctx context.Context, user *entity.User) error {
	return uc.userRepo.AddUser(ctx, user)
}

func (uc *userUseCase) UpdateUser(ctx context.Context, user *entity.User) error {
	return uc.userRepo.UpdateUser(ctx, user)
}
