package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrLostLease = errors.New("outbox: lost lease")

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

func (r *OutboxRepo) MarkSent(
	ctx context.Context, 
	id int64, 
	instanceID string,
) error {
	query := `
		UPDATE outbox
		SET
			status    = 'sent',
			sent_at   = $3,
			locked_by = NULL,
			locked_at = NULL
		WHERE id = $1
		  AND status = 'processing'
		  AND locked_by = $2
	`
	res, err := r.db.ExecContext(ctx, query, id, instanceID, time.Now())
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("outbox %d: %w", id, ErrLostLease)
	}

	return nil
}

func (r *OutboxRepo) Claim(
	ctx context.Context,
	instanceID string,
	limit int,
	staleAfter time.Duration,
) ([]OutboxMessage, error) {
	query := `
		UPDATE outbox
		SET
			status    = 'processing',
			locked_by = $1,
			locked_at = now(),
			attempts  = attempts + 1
		WHERE id IN (
			SELECT id
			FROM outbox
			WHERE (
					status = 'pending'
					AND (next_retry_at IS NULL OR next_retry_at <= now())
				)
				OR (
					status = 'processing'
					AND locked_at < now() - ($2::bigint * interval '1 second')
				)
			ORDER BY created_at
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, event_id, topic, payload
	`
	rows, err := r.db.QueryContext(
		ctx,
		query,
		instanceID,
		int64(staleAfter.Seconds()),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	msgs := make([]OutboxMessage, 0)
	for rows.Next() {
		var msg OutboxMessage
		if err := rows.Scan(&msg.ID, &msg.EventID, &msg.Topic, &msg.Payload); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}

	return msgs, rows.Err()
}

func (r *OutboxRepo) Nack(
	ctx context.Context,
	id int64,
	instanceID string,
	lastError string,
	maxAttempts int,
	backoff time.Duration,
) error {
	query := `
		UPDATE outbox
		SET
			status = CASE
				WHEN attempts >= $4 THEN 'failed'
				ELSE 'pending'
			END,
			last_error = $3,
			next_retry_at = CASE
				WHEN attempts >= $4 THEN NULL
				ELSE now() + ($5::bigint * interval '1 second')
			END,
			locked_by = NULL,
			locked_at = NULL
		WHERE id = $1
		  AND status = 'processing'
		  AND locked_by = $2
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		id,
		instanceID,
		lastError,
		maxAttempts,
		int64(backoff.Seconds()),
	)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("outbox %d: %w", id, ErrLostLease)
	}

	return nil
}

