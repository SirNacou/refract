package main

import (
	"log/slog"
	"os"

	"github.com/SirNacou/refract/services/api/internal/app"
	"github.com/SirNacou/refract/services/api/internal/config"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	// Create and initialize application
	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize application",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	// Run application (blocks until shutdown)
	if err := application.Run(); err != nil {
		logger.Error("Application error",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
}
