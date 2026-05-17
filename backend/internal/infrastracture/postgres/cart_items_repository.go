package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/mosgor/Evently/backend/internal/entity"
)

type CartItemsRepository struct {
	db *Postgres
}

func (cir *CartItemsRepository) GetCartItems(ctx context.Context, userId int) ([]*entity.Event, error) {
	q := `
		  SELECT ev.id, ev.title, ev.date, ev.description, ev.image, ev.price 
		  FROM cart_items AS ci JOIN events AS ev ON ci.event_id = ev.id 
		  WHERE ci.user_id = $1
    `
	query, err := cir.db.Pool.Query(ctx, q, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrorNotFound
	}
	if err != nil {
		return nil, errors.New("error in getting all cartItems: " + err.Error())
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
			return nil, errors.New("error in getting all cartItems: " + err.Error())
		}
		events = append(events, &event)
	}
	return events, nil
}

func (cir *CartItemsRepository) AddCartItem(ctx context.Context, userId, eventId int) error {
	q := `INSERT INTO cart_items(user_id, event_id) VALUES ($1, $2)`
	_, err := cir.db.Pool.Exec(ctx, q, userId, eventId)
	if err != nil {
		return fmt.Errorf("error in adding cartItem: %w", err)
	}
	return nil
}

func (cir *CartItemsRepository) RemoveCartItem(ctx context.Context, userId, eventId int) error {
	q := `DELETE FROM cart_items WHERE event_id = $1 AND user_id = $2`
	_, err := cir.db.Pool.Exec(ctx, q, eventId, userId)
	if err != nil {
		return fmt.Errorf("error in removing cartItem: %w", err)
	}
	return nil
}

func NewCartItemsRepository(db *Postgres) *CartItemsRepository {
	return &CartItemsRepository{db: db}
}
