package usecase

import (
	"context"
)

type PaymentUseCase interface {
	ProcessPayment(ctx context.Context, userID int, eventIDs []int) error
}

type paymentUseCase struct {
	cartRepo   CartRepository
	ticketRepo TicketsRepository
}

func NewPaymentUseCase(cartRepo CartRepository, ticketRepo TicketsRepository) PaymentUseCase {
	return &paymentUseCase{
		cartRepo:   cartRepo,
		ticketRepo: ticketRepo,
	}
}

func (uc *paymentUseCase) ProcessPayment(ctx context.Context, userID int, eventIDs []int) error {
	for _, eventID := range eventIDs {
		err := uc.ticketRepo.AddTicket(ctx, userID, eventID)
		if err != nil {
			return err
		}

		err = uc.cartRepo.RemoveCartItem(ctx, userID, eventID)
		if err != nil {
			return err
		}
	}

	return nil
}
