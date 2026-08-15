package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ikondratev/event-service/internal/domain/idempotency"
)

type IdempotencyRepo struct {
	db *sql.DB
}

func NewIdempotencyRepo(db *sql.DB) *IdempotencyRepo {
	return &IdempotencyRepo{db: db}
}

var _ idempotency.IdempotencyRepo = (*IdempotencyRepo)(nil)

func (r *IdempotencyRepo) TryInsert(ctx context.Context, key, requestHash string) (bool, error) {
	tx, err := requireTx(ctx)
	if err != nil {
		return false, err
	}

	query := `
		INSERT INTO idempotency_keys (key, request_hash)
		VALUES ($1, $2)
		ON CONFLICT (key) DO NOTHING
		RETURNING key
	`

	var insertedKey string
	err = tx.QueryRowContext(ctx, query, key, requestHash).Scan(&insertedKey)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *IdempotencyRepo) Get(ctx context.Context, key string) (*idempotency.IdempotencyKey, error) {
	tx, err := requireTx(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT key, request_hash, event_id
		FROM idempotency_keys
		WHERE key = $1
	`

	var k idempotency.IdempotencyKey
	var eventID sql.NullInt64
	err = tx.QueryRowContext(ctx, query, key).Scan(&k.Key, &k.RequestHash, &eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, idempotency.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if eventID.Valid {
		id := eventID.Int64
		k.EventID = &id
	}
	return &k, nil
}

func (r *IdempotencyRepo) AttachEvent(ctx context.Context, key string, eventID int64) error {
	tx, err := requireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE idempotency_keys
		SET event_id = $2
		WHERE key = $1
	`

	res, err := tx.ExecContext(ctx, query, key, eventID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return idempotency.ErrNotFound
	}
	return nil
}
