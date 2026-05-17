package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/mosgor/Evently/backend/internal/entity"
)

type EventRepository struct {
	db *Postgres
}

func (repo *EventRepository) GetAllEvents(ctx context.Context) ([]*entity.Event, error) {
	q := `SELECT id, title, date, description, image, price FROM events`
	query, err := repo.db.Pool.Query(ctx, q)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrorNotFound
	}
	if err != nil {
		return nil, errors.New("error in getting all events: " + err.Error())
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
			return nil, errors.New("error in getting all events: " + err.Error())
		}
		events = append(events, &event)
	}
	return events, nil
}

func (repo *EventRepository) GetUserPreferenceVector(ctx context.Context, userID int) ([]float32, error) {
	rows, err := repo.db.Pool.Query(ctx, `
		SELECT e.embedding::text, ub.event_type
		FROM user_behavior ub
		JOIN events e ON ub.event_id = e.id
		WHERE ub.user_id = $1 AND e.embedding IS NOT NULL
		ORDER BY ub.created_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch behavior: %w", err)
	}
	defer rows.Close()

	weights := map[string]float32{"view": 1.0, "cart_add": 2.0, "purchase": 3.0}
	var sumVec []float32
	var totalWeight float32
	var hasData bool

	for rows.Next() {
		var vecStr string
		var eventType string
		if err := rows.Scan(&vecStr, &eventType); err != nil {
			return nil, err
		}

		var embeddingF64 []float64
		if err := json.Unmarshal([]byte(vecStr), &embeddingF64); err != nil {
			return nil, fmt.Errorf("parse vector: %w", err)
		}

		embedding := make([]float32, len(embeddingF64))
		for i, v := range embeddingF64 {
			embedding[i] = float32(v)
		}

		w := weights[eventType]
		if w == 0 {
			w = 1.0
		}

		if !hasData {
			sumVec = make([]float32, len(embedding))
			hasData = true
		}

		for i := range embedding {
			sumVec[i] += embedding[i] * w
		}
		totalWeight += w
	}

	if !hasData || totalWeight == 0 {
		return nil, nil
	}

	for i := range sumVec {
		sumVec[i] /= totalWeight
	}

	var norm float32
	for _, v := range sumVec {
		norm += v * v
	}

	norm = float32(math.Sqrt(float64(norm)))

	if norm > 0 {
		for i := range sumVec {
			sumVec[i] /= norm
		}
	}

	return sumVec, nil
}

func (repo *EventRepository) GetRecommendedByVector(ctx context.Context, userVec []float32, userID, limit int) ([]*entity.Event, error) {
	vecParts := make([]string, len(userVec))
	for i, v := range userVec {
		vecParts[i] = fmt.Sprintf("%f", v)
	}
	vecStr := "[" + strings.Join(vecParts, ",") + "]"

	q := `
		SELECT id, title, date, description, image, price
		FROM events
		WHERE embedding IS NOT NULL
		  AND id NOT IN (SELECT event_id FROM user_behavior WHERE user_id = $1 AND event_type = 'purchase')
		  AND date >= NOW()
		ORDER BY embedding <=> $2::vector
		LIMIT $3;
	`

	rows, err := repo.db.Pool.Query(ctx, q, userID, vecStr, limit)
	if err != nil {
		return nil, fmt.Errorf("recommendation query failed: %w", err)
	}
	defer rows.Close()

	var events []*entity.Event
	for rows.Next() {
		var ev entity.Event
		if err := rows.Scan(&ev.Id, &ev.Title, &ev.Date, &ev.Description, &ev.Image, &ev.Price); err != nil {
			return nil, err
		}
		events = append(events, &ev)
	}
	return events, nil
}

func NewEventPostgresRepo(db *Postgres) *EventRepository {
	return &EventRepository{db: db}
}
