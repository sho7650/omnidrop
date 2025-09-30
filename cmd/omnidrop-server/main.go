package main

import (
	"log/slog"
	"os"

	"omnidrop/internal/app"
	"omnidrop/internal/observability"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Setup structured logging
	logger := observability.SetupLogger()
	
	application := app.NewWithVersion(Version, BuildTime)
	if err := application.Run(); err != nil {
		logger.Error("Application failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
