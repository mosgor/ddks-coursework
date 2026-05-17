package usecase

import (
	"context"

	"github.com/mosgor/Evently/backend/internal/entity"
)

type TicketUseCase interface {
	GetTickets(ctx context.Context, userId int) ([]*entity.Event, error)
	AddTicket(ctx context.Context, userId int, eventIds []int) error
}

type ticketUseCase struct {
	ticketsRepo TicketsRepository
}

func NewTicketUseCase(ticketsRepo TicketsRepository) TicketUseCase {
	return &ticketUseCase{
		ticketsRepo: ticketsRepo,
	}
}

func (uc *ticketUseCase) GetTickets(ctx context.Context, userId int) ([]*entity.Event, error) {
	return uc.ticketsRepo.GetTickets(ctx, userId)
}

func (uc *ticketUseCase) AddTicket(ctx context.Context, userId int, eventIds []int) error {
	for _, eventId := range eventIds {
		err := uc.ticketsRepo.AddTicket(ctx, userId, eventId)
		if err != nil {
			return err
		}
	}
	return nil
}
