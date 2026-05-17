package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/mosgor/Evently/backend/internal/entity"
)

type TicketsRepository struct {
	db *Postgres
}

func (tr *TicketsRepository) GetTickets(ctx context.Context, userId int) ([]*entity.Event, error) {
	q := `
		  SELECT ev.id, ev.title, ev.date, ev.description, ev.image, ev.price 
		  FROM tickets AS tk JOIN events AS ev ON tk.event_id = ev.id 
		  WHERE tk.user_id = $1
    `
	query, err := tr.db.Pool.Query(ctx, q, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrorNotFound
	}
	if err != nil {
		return nil, errors.New("error in getting all tickets: " + err.Error())
	}
	var events []*entity.Event
	for query.Next() {
		var event entity.Event
		err = query.Scan(
			&event.Id, &event.Title,
			&event.Date, &event.Description,
			&event.Image, &event.Price,
		)
		if err != nil {
			return nil, errors.New("error in getting all tickets: " + err.Error())
		}
		events = append(events, &event)
	}
	return events, nil
}

func (tr *TicketsRepository) AddTicket(ctx context.Context, userId, eventId int) error {
	q := `INSERT INTO tickets(user_id, event_id) VALUES ($1, $2)`
	_, err := tr.db.Pool.Exec(ctx, q, userId, eventId)
	if err != nil {
		return fmt.Errorf("error in adding ticket: %w", err)
	}
	return nil
}

func NewTicketsRepository(db *Postgres) *TicketsRepository {
	return &TicketsRepository{db: db}
}
