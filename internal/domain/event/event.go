package event

import (
	"context"
	"time"
)

type Event struct {
	ID        int64
	Kind      string
	Data      string
	CreatedAt time.Time
}

type EventRepo interface {
	Create(ctx context.Context, e *Event) error
	List(ctx context.Context) ([]Event, error)
	GetByID(ctx context.Context, id int64) (*Event, error)
}
