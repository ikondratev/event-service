package eventdto

import (
	"encoding/json"

	"github.com/ikondratev/event-service/internal/domain/event"
)

func ToDomain(req CreateRequest) (*event.Event, error) {
	raw, err := json.Marshal(req.Data)
	if err != nil {
		return nil, err
	}

	return &event.Event{
		Kind:   req.Kind,
		Status: req.Status,
		Data:   string(raw),
	}, nil
}

func ToResponse(e event.Event) Response {
	return Response {
		ID:        e.ID,
		Kind:      e.Kind,
		Status:    e.Status,
		CreatedAt: e.CreatedAt,
	}
}