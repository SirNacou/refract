package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/SirNacou/refract/services/api/internal/database"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	log.Println("Connecting to database...")
	pool, err := database.Connect(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Database connected successfully")

	// TODO: Initialize HTTP server, routes, handlers

	log.Printf("API server ready to start on port %s", cfg.Port)
	log.Println("(HTTP server implementation pending)")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
