package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"microserice/internal/domain/event"
	eventdto "microserice/internal/dto/event"
)

type EventHandler struct {
	logger    *slog.Logger
	repo      event.EventRepo
}

func NewEventHandler(logger *slog.Logger, eventRepo event.EventRepo) * EventHandler {
	return &EventHandler{
		logger: logger,
		repo: 	eventRepo,
	}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req eventdto.CreateRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Kind == "" || req.Data == "" || req.Status == "" {
		http.Error(w, "kind, data, status are requeired", http.StatusBadRequest)
		return
	}

	domainEvent := eventdto.ToDomain(req)

	if err := h.repo.Create(r.Context(), domainEvent); err != nil {
		h.logger.Error("failed to create event", "error", err)
		http.Error(w, "failed to create event", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(eventdto.ToResponse(*domainEvent))
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.repo.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list events", "error", err)
		http.Error(w, "failed to get events", http.StatusInternalServerError)
		return
	}

	response := make([]eventdto.Response, 0, len(events))
	for _, e := range events {
		response = append(response, eventdto.ToResponse(e) )
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode", http.StatusInternalServerError)
	}
}