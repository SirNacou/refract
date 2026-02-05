package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SirNacou/refract/api/internal/config"
	"github.com/SirNacou/refract/api/internal/features/urls/redirect"
	"github.com/SirNacou/refract/api/internal/infrastructure/cache"
	"github.com/SirNacou/refract/api/internal/infrastructure/persistence"
	"github.com/SirNacou/refract/api/internal/infrastructure/publisher"
	"github.com/SirNacou/refract/api/internal/infrastructure/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fatal("Config", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	db, err := persistence.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}
	defer db.Close()

	repo := repository.NewPostgresURLRepository(db.Querier)

	valkey, err := cache.NewCache(ctx, &cfg.Valkey)
	if err != nil {
		fatal("Valkey", err)
	}
	defer valkey.Close()

	clicksPublisher := publisher.NewClicksPublisher(valkey.Client(), cfg.Valkey.ClicksStreamKey)

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	redirectHandler := redirect.NewRedirectHandler(valkey, repo, clicksPublisher, cfg)
	r.Get("/{shortCode}", redirectHandler.Handle)

	srv := http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%v", cfg.RedirectorPort),
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start Server: %v", err)
	}

	<-ctx.Done()

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Failed to shut down server: %v", err)
	}
}

func fatal(name string, err error) {
	log.Fatalf("Failed to initialize %v: %v", name, err)
}
