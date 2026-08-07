package event

import (
	"time"
	"context"
) 

type Event struct {
	ID 		  int64
	Kind 	  string
	Data 	  string
	Status 	  string
	CreatedAt time.Time
}


type EventRepo interface {
	Create(ctx context.Context, e *Event) error
	List(ctx context.Context) ([]Event, error)
}


