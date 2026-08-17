package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/ikondratev/event-service/internal/application"
)

func main() {
	env := os.Getenv("APP_ENV")
	if strings.TrimSpace(env) == "" {
		env = "dev"
	}

	worker, err := application.NewWorker(env)
	if err != nil {
		slog.Error("start worker error", "error", err)
		os.Exit(1)
	}

	if err := worker.Run(); err != nil {
		slog.Error("woker error", "error", err)
		os.Exit(1)
	}

}