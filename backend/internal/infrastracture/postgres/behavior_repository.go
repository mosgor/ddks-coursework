package postgres

import (
	"context"
)

type BehaviorRepository struct{ db *Postgres }

func (r *BehaviorRepository) LogInteraction(ctx context.Context, userID, eventID int, eventType string) error {
	q := `INSERT INTO user_behavior(user_id, event_id, event_type) VALUES ($1, $2, $3)`
	_, err := r.db.Pool.Exec(ctx, q, userID, eventID, eventType)
	return err
}

func NewBehaviorRepository(db *Postgres) *BehaviorRepository {
	return &BehaviorRepository{db: db}
}
