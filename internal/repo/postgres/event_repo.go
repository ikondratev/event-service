package postgres

import (
	"fmt"
	"context"
	"database/sql"
	"encoding/json"
	
	"microserice/internal/domain/event"
	"microserice/internal/settings"
)

type EventRepo struct {
	db 		   *sql.DB
	settings   *settings.Settings
}

func NewEventRepo(db *sql.DB, settings *settings.Settings) *EventRepo {
	return &EventRepo{
		db: 		db,
		settings:   settings,
	}
}

var _ event.EventRepo = (*EventRepo)(nil)

func (r *EventRepo) Create(ctx context.Context, e *event.Event) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("TX error: %w", err)
	}
	defer tx.Rollback()

	queryEvent := `
		INSERT INTO events (kind, data, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	if err := tx.QueryRowContext(ctx, queryEvent, e.Kind, e.Data, e.Status).
	Scan(&e.ID, &e.CreatedAt); err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]any{
		"id": 		  e.ID,
		"kind": 	  e.Kind,
		"data": 	  e.Data,
		"status": 	  e.Status,
		"created_at": e.CreatedAt,
	})
	if err != nil {
		return err
	}

	queryOutbox := `
		INSERT INTO outbox (event_id, topic, payload, status)
		VALUES ($1, $2, $3, 'pending')
	`
	if _, err := tx.ExecContext(ctx, queryOutbox, e.ID, r.settings.Kafka.Topics["events_created"], payload); err != nil {
		return err
	}
	

	return tx.Commit()	
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