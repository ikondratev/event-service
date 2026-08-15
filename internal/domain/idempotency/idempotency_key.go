package idempotency

import "context"

type IdempotencyKey struct {
	Key 	    string
	RequestHash string
	EventID 	*int64
}

type IdempotencyRepo interface {
	TryInsert(ctx context.Context, key, requestHash string) (bool, error)
	Get(ctx context.Context, key string) (*IdempotencyKey, error)
	AttachEvent(ctx context.Context, key string, eventID int64) error
}