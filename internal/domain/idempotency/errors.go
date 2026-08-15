package idempotency

import "errors"

var (
    ErrMissingKey      = errors.New("idempotency key is required")
    ErrInvalidKey      = errors.New("idempotency key is invalid")
    ErrPayloadMismatch = errors.New("idempotency key reused with different payload")
    ErrNotFound        = errors.New("idempotency key not found")
)