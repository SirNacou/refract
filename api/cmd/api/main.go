package main

import (
	"log"
	"log/slog"

	"github.com/SirNacou/refract/api/internal/config"
	"github.com/SirNacou/refract/api/internal/infrastructure/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	router := server.NewRouter(cfg.Port)

	log.Printf("Starting server on port %d", cfg.Port)

	if err := router.Run(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	slog.Info("Server shut down")
}
