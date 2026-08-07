package postgres

import (
	"context"
	"database/sql"
	"microserice/internal/domain/event"
)

type EventRepo struct {
	db *sql.DB
}

func NewEventRepo(db *sql.DB) *EventRepo {
	return &EventRepo{db: db}
}

var _ event.EventRepo = (*EventRepo)(nil)

func (r *EventRepo) Create(ctx context.Context, e *event.Event) error {
	query := `
		INSERT INTO events (kind, data, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query, e.Kind, e.Data, e.Status).Scan(&e.ID, &e.CreatedAt)
}

func (r *EventRepo) List(ctx context.Context) ([]event.Event, error) {
	query := `
		SELECT id, kind, data, status, created_at
		FROM events
		ORDER BY id DESC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx,query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]event.Event, 0)
	for rows.Next() {
		var e event.Event
		if err := rows.Scan(
			&e.ID, 
			&e.Kind, 
			&e.Data, 
			&e.Status, 
			&e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, rows.Err()
}