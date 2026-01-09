package idgen

import (
	"fmt"
	"os"
	"strconv"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
)

// IDGenerator is the interface for generating unique IDs
type IDGenerator interface {
	NextID() (uint64, error)
}

// SnowflakeGenerator is the infrastructure implementation that wraps the domain Snowflake generator
type SnowflakeGenerator struct {
	generator *url.SnowflakeGenerator
	workerID  int64
}

// NewSnowflakeGenerator creates a new Snowflake ID generator from environment configuration
// Reads WORKER_ID environment variable (defaults to 0 if not set)
// Worker ID ranges:
//   - API service: 0-63 (recommended)
//   - Redirector service: 64-127 (recommended)
//   - Maximum: 0-1023
func NewSnowflakeGenerator() (*SnowflakeGenerator, error) {
	workerID, err := getWorkerIDFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker ID from environment: %w", err)
	}

	generator, err := url.NewSnowflakeGenerator(workerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake generator: %w", err)
	}

	return &SnowflakeGenerator{
		generator: generator,
		workerID:  workerID,
	}, nil
}

// NewSnowflakeGeneratorWithWorkerID creates a new Snowflake ID generator with explicit worker ID
// Useful for testing or when worker ID is provided programmatically
func NewSnowflakeGeneratorWithWorkerID(workerID int64) (*SnowflakeGenerator, error) {
	generator, err := url.NewSnowflakeGenerator(workerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake generator: %w", err)
	}

	return &SnowflakeGenerator{
		generator: generator,
		workerID:  workerID,
	}, nil
}

// NextID generates the next unique Snowflake ID
func (s *SnowflakeGenerator) NextID() (uint64, error) {
	return s.generator.NextID()
}

// WorkerID returns the generator's configured worker ID
func (s *SnowflakeGenerator) WorkerID() int64 {
	return s.workerID
}

// getWorkerIDFromEnv reads the WORKER_ID environment variable
// Returns 0 as default if not set or invalid
func getWorkerIDFromEnv() (int64, error) {
	workerIDStr := os.Getenv("WORKER_ID")
	if workerIDStr == "" {
		// Default to worker ID 0 for local development
		return 0, nil
	}

	workerID, err := strconv.ParseInt(workerIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid WORKER_ID environment variable '%s': %w", workerIDStr, err)
	}

	if workerID < 0 || workerID > 1023 {
		return 0, fmt.Errorf("WORKER_ID must be between 0 and 1023, got %d", workerID)
	}

	return workerID, nil
}
