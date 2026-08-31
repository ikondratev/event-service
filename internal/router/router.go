package router

import (
	"log/slog"
	"net/http"

	"github.com/ikondratev/event-service/internal/auth"
	"github.com/ikondratev/event-service/internal/handlers"
	"github.com/ikondratev/event-service/internal/middleware"
	"github.com/ikondratev/event-service/internal/service/event"

	"github.com/gorilla/mux"
)

type Router struct {
	engine   *mux.Router
	logger   *slog.Logger
	service  *eventsvc.Service
	verifier *auth.Verifier
}

func New(logger *slog.Logger, service *eventsvc.Service, verifier *auth.Verifier) *Router {
	router := mux.NewRouter()
	router.Use(middleware.Logging(logger))

	return &Router{
		engine:   router,
		logger:   logger,
		service:  service,
		verifier: verifier,
	}
}

func (r *Router) RegisterRotes() *mux.Router {
	eventHandler := handlers.NewEventHandler(r.logger, r.service)
	r.engine.HandleFunc("/ping", handlers.Pong).Methods(http.MethodGet)

	api := r.engine.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.Authenticate(r.verifier))

	write := api.NewRoute().Subrouter()
	write.Use(middleware.RequireScope("events:write"))
	write.HandleFunc("/event", eventHandler.CreateEvent).Methods(http.MethodPost)
	
	read := api.NewRoute().Subrouter()
	read.Use(middleware.RequireScope("events:read"))
	read.HandleFunc("/events", eventHandler.GetEvents).Methods(http.MethodGet)

	return r.engine
}