package idgen

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// ==================== SNOWFLAKE ID GENERATION (T014) ====================

// Snowflake ID bit layout (64 bits total):
// - 1 bit: unused (always 0)
// - 41 bits: timestamp (milliseconds since custom epoch)
// - 10 bits: worker ID (0-1023)
// - 12 bits: sequence number (0-4095 per millisecond)

const (
	// Custom epoch: 2026-01-01 00:00:00 UTC (in seconds)
	customEpochSeconds int64 = 1767200400

	// Bit allocation
	workerIDBits   = 10
	sequenceBits   = 12
	maxWorkerID    = -1 ^ (-1 << workerIDBits)   // 1023
	maxSequence    = -1 ^ (-1 << sequenceBits)   // 4095
	workerIDShift  = sequenceBits                // 12
	timestampShift = sequenceBits + workerIDBits // 22
)

var (
	// ErrInvalidWorkerID indicates worker ID is out of valid range (0-1023)
	ErrInvalidWorkerID = errors.New("worker ID must be between 0 and 1023")
	// ErrClockMovedBack indicates system clock moved backwards
	ErrClockMovedBack = errors.New("clock moved backwards, refusing to generate ID")
)

// IDGenerator is the interface for generating unique IDs
type IDGenerator interface {
	NextID() (uint64, error)
	WorkerID() int64
}

// SnowflakeGenerator generates distributed unique IDs using the Snowflake algorithm
type SnowflakeGenerator struct {
	workerID   int64
	sequence   int64
	lastTimeMS int64
	mutex      sync.Mutex
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

	return NewSnowflakeGeneratorWithWorkerID(workerID)
}

// NewSnowflakeGeneratorWithWorkerID creates a new Snowflake ID generator with explicit worker ID
// workerID must be between 0-1023
// API service instances: 0-63 (recommended)
// Redirector service instances: 64-127 (recommended)
func NewSnowflakeGeneratorWithWorkerID(workerID int64) (*SnowflakeGenerator, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, ErrInvalidWorkerID
	}

	return &SnowflakeGenerator{
		workerID:   workerID,
		sequence:   0,
		lastTimeMS: -1,
	}, nil
}

// NextID generates the next unique Snowflake ID
// Thread-safe: uses mutex to prevent race conditions
// Returns error if clock moves backwards
func (g *SnowflakeGenerator) NextID() (uint64, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Get current timestamp in milliseconds since custom epoch
	nowMS := time.Now().UnixMilli() - (customEpochSeconds * 1000)

	// Clock moved backwards - this should rarely happen
	if nowMS < g.lastTimeMS {
		return 0, ErrClockMovedBack
	}

	if nowMS == g.lastTimeMS {
		// Same millisecond - increment sequence
		g.sequence = (g.sequence + 1) & maxSequence

		if g.sequence == 0 {
			// Sequence exhausted - wait for next millisecond
			for nowMS <= g.lastTimeMS {
				time.Sleep(time.Millisecond)
				nowMS = time.Now().UnixMilli() - (customEpochSeconds * 1000)
			}
		}
	} else {
		// New millisecond - reset sequence
		g.sequence = 0
	}

	g.lastTimeMS = nowMS

	// Construct the Snowflake ID:
	// [timestamp: 41 bits] [workerID: 10 bits] [sequence: 12 bits]
	id := (nowMS << timestampShift) | (g.workerID << workerIDShift) | g.sequence

	return uint64(id), nil
}

// WorkerID returns the generator's configured worker ID
func (g *SnowflakeGenerator) WorkerID() int64 {
	return g.workerID
}

// ExtractTimestamp extracts the timestamp from a Snowflake ID
// Returns milliseconds since custom epoch
func ExtractTimestamp(id uint64) int64 {
	return int64(id >> timestampShift)
}

// ExtractWorkerID extracts the worker ID from a Snowflake ID
func ExtractWorkerID(id uint64) int64 {
	return int64((id >> workerIDShift) & uint64(maxWorkerID))
}

// ExtractSequence extracts the sequence number from a Snowflake ID
func ExtractSequence(id uint64) int64 {
	return int64(id & uint64(maxSequence))
}

// IDToTime converts a Snowflake ID to the actual timestamp
func IDToTime(id uint64) time.Time {
	timestampMS := ExtractTimestamp(id)
	timestampSeconds := (timestampMS / 1000) + customEpochSeconds
	return time.Unix(timestampSeconds, (timestampMS%1000)*1000000)
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

// ==================== BASE62 ENCODING (T015) ====================

// Base62 alphabet for URL-safe encoding
// Uses 0-9, a-z, A-Z (62 characters total)
const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// EncodeBase62 converts a Snowflake ID (uint64) to a Base62 string
// Produces URL-safe short codes (8-11 characters for typical Snowflake IDs)
// Example: 1234567890123456 -> "5Ezg7yb1S"
func EncodeBase62(id uint64) string {
	if id == 0 {
		return "0"
	}

	encoded := make([]byte, 0, 11) // Max 11 chars for uint64
	base := uint64(62)

	for id > 0 {
		remainder := id % base
		encoded = append(encoded, base62Alphabet[remainder])
		id /= base
	}

	// Reverse the string (encoded is backwards)
	for i, j := 0, len(encoded)-1; i < j; i, j = i+1, j-1 {
		encoded[i], encoded[j] = encoded[j], encoded[i]
	}

	return string(encoded)
}

// DecodeBase62 converts a Base62 string back to a Snowflake ID (uint64)
// Returns error if string contains invalid characters
// Example: "5Ezg7yb1S" -> 1234567890123456
func DecodeBase62(encoded string) (uint64, error) {
	if encoded == "" {
		return 0, fmt.Errorf("empty Base62 string")
	}

	var id uint64
	base := uint64(62)

	for _, char := range encoded {
		// Find character position in alphabet
		pos := -1
		for i, c := range base62Alphabet {
			if c == char {
				pos = i
				break
			}
		}

		if pos == -1 {
			return 0, fmt.Errorf("invalid Base62 character: %c", char)
		}

		// Check for overflow before multiplication
		if id > (^uint64(0) / base) {
			return 0, fmt.Errorf("Base62 string too long, would overflow uint64")
		}

		id = id*base + uint64(pos)
	}

	return id, nil
}
