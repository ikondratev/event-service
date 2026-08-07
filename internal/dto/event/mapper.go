package eventdto

import "microserice/internal/domain/event"

func ToDomain(req CreateRequest) *event.Event {
	return &event.Event{
		Kind:   req.Kind,
		Data:   req.Data,
		Status: req.Status,
	}
}

func ToResponse(e event.Event) Response {
	return Response {
		ID:        e.ID,
		Kind:      e.Kind,
		Data:      e.Data,
		Status:    e.Status,
		CreatedAt: e.CreatedAt,
	}
}