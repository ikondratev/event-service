package router

import (
	"log/slog"
	"net/http"

	"github.com/ikondratev/event-service/internal/service/event"
	"github.com/ikondratev/event-service/internal/handlers"
	"github.com/ikondratev/event-service/internal/middleware"

	"github.com/gorilla/mux"
)

type Router struct {
	engine  *mux.Router
	logger  *slog.Logger
	service *eventsvc.Service
}

func New(logger *slog.Logger, service *eventsvc.Service) *Router {
	router := mux.NewRouter()
	router.Use(middleware.Logging(logger))

	return &Router{
		engine:  router,
		logger:  logger,
		service: service,
	}
}

func (r *Router) RegisterRotes() *mux.Router {
	eventHandler := handlers.NewEventHandler(r.logger, r.service)
	r.engine.HandleFunc("/ping", handlers.Pong).Methods(http.MethodGet)
	api := r.engine.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/event", eventHandler.CreateEvent).Methods(http.MethodPost)
	api.HandleFunc("/events", eventHandler.GetEvents).Methods(http.MethodGet)

	return r.engine
}