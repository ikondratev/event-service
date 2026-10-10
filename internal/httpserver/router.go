package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct {
	engine *mux.Router
}

func NewRouter(logger *slog.Logger) *Router {
	router := mux.NewRouter()
	router.Use(Logging(logger))

	return &Router{
		engine: router,
	}
}

func (r *Router) RegisterRotes() *mux.Router {
	r.engine.HandleFunc("/ping", Pong).Methods(http.MethodGet)

	return r.engine
}
