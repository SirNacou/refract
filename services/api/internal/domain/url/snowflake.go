package url

import (
	"errors"
	"sync"
	"time"
)

// Snowflake ID bit layout (64 bits total):
// - 1 bit: unused (always 0)
// - 41 bits: timestamp (milliseconds since custom epoch)
// - 10 bits: worker ID (0-1023)
// - 12 bits: sequence number (0-4095 per millisecond)

var (
	// Custom epoch: 2024-01-01 00:00:00 UTC (reduces ID size for ~10 years)
	customEpoch int64 = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli() // time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()

	// Bit allocation
	workerIDBits   = 10
	sequenceBits   = 12
	maxWorkerID    = -1 ^ (-1 << workerIDBits)   // 1023
	maxSequence    = -1 ^ (-1 << sequenceBits)   // 4095
	workerIDShift  = sequenceBits                // 12
	timestampShift = sequenceBits + workerIDBits // 22
)

var (
	ErrInvalidWorkerID   = errors.New("worker ID must be between 0 and 1023")
	ErrClockMovedBack    = errors.New("clock moved backwards, refusing to generate ID")
	ErrSequenceExhausted = errors.New("sequence exhausted for current millisecond")
)

// SnowflakeGenerator generates distributed unique IDs using the Snowflake algorithm
type SnowflakeGenerator struct {
	workerID   int64
	sequence   int64
	lastTimeMS int64
	mutex      sync.Mutex
}

// NewSnowflakeGenerator creates a new Snowflake ID generator
// workerID must be between 0-1023
// API service instances: 0-63
// Redirector service instances: 64-127
func NewSnowflakeGenerator(workerID int64) (*SnowflakeGenerator, error) {
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
func (g *SnowflakeGenerator) NextID() (uint64, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Get current timestamp in milliseconds since custom epoch
	nowMS := time.Now().UnixMilli() - customEpoch

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
				nowMS = time.Now().UnixMilli() - customEpoch
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

// WorkerID returns the generator's worker ID
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
	timestamp := ExtractTimestamp(id) + customEpoch
	return time.UnixMilli(timestamp)
}
