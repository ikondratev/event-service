package main

import (
	"os"
	"log/slog"

	"microserice/lib/application"
)

const env = "dev"

func main() {
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