package main

import (
	"os"
	"log/slog"

	"github.com/ikondratev/event-service/internal/application"
)

func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	
	app, err := application.New(env)
	if err != nil {
		slog.Error("app load with error", "error", err)
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		slog.Error("app run with error", "error", err)
		os.Exit(1)
	}
}