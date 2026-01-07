package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SirNacou/refract/services/api/internal/config"
)

// Server manages the HTTP server lifecycle
type Server struct {
	cfg    *config.Config
	server *http.Server
	logger *slog.Logger
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.Port),
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Run starts the server and blocks until shutdown signal
func (s *Server) Run() error {
	// Channel for server errors
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		s.logger.Info("Server listening", slog.String("addr", s.server.Addr))
		serverErrors <- s.server.ListenAndServe()
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until error or signal
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server failed: %w", err)
		}
		return nil
	case sig := <-quit:
		s.logger.Info("Shutdown signal received", slog.String("signal", sig.String()))
		return s.Shutdown(context.Background())
	}
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.logger.Info("Shutting down server...")

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("Server stopped gracefully")
	return nil
}
