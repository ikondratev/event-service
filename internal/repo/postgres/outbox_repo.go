package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type OutboxMessage struct {
	ID 		int64
	EventID int64
	Topic 	string
	Payload json.RawMessage
}

type OutboxRepo struct {
	db *sql.DB
}

func NewOutboxRepo(db *sql.DB) *OutboxRepo {
	return &OutboxRepo{db: db}
}

func (r *OutboxRepo) FetchPending(ctx context.Context, limit int) ([]OutboxMessage, error) {
	query := `
		SELECT id, event_id, topic, payload
		FROM outbox
		WHERE status = 'pending'
		ORDER BY created_at
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	msgs := make([]OutboxMessage, 0)
	for rows.Next() {
		var m OutboxMessage
		if err := rows.Scan(&m.ID, &m.EventID, &m.Topic, &m.Payload); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (r *OutboxRepo) MarkSent(ctx context.Context, id int64) error {
	query := `
		UPDATE outbox
		SET status = 'sent', sent_at = $2
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, time.Now())

	return err
}