package idgen

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

// SnowflakeGenerator generates distributed unique IDs using Twitter's Snowflake algorithm
//
// Snowflake ID format (64 bits):
// - 1 bit:  unused (always 0)
// - 41 bits: timestamp in milliseconds since epoch
// - 10 bits: node/machine ID (0-1023)
// - 12 bits: sequence number (0-4095)
//
// This provides:
// - Time-ordered IDs (sortable by creation time)
// - Distributed generation (multiple nodes won't collide)
// - 4096 IDs per millisecond per node
// - No database dependency for ID generation
type SnowflakeGenerator struct {
	node *snowflake.Node
	mu   sync.RWMutex
}

var (
	instance *SnowflakeGenerator
	once     sync.Once
)

// NewSnowflakeGenerator creates a new Snowflake ID generator
// nodeID must be between 0 and 1023
func NewSnowflakeGenerator(nodeID int64) (*SnowflakeGenerator, error) {
	if nodeID < 0 || nodeID > 1023 {
		return nil, fmt.Errorf("node ID must be between 0 and 1023, got %d", nodeID)
	}

	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %w", err)
	}

	return &SnowflakeGenerator{
		node: node,
	}, nil
}

// GetInstance returns the singleton instance of SnowflakeGenerator
// If not initialized, it will create one with the provided nodeID
func GetInstance(nodeID int64) (*SnowflakeGenerator, error) {
	var err error
	once.Do(func() {
		instance, err = NewSnowflakeGenerator(nodeID)
	})
	return instance, err
}

// Generate creates a new unique Snowflake ID
// This method is thread-safe and never fails (unless system clock goes backwards)
func (g *SnowflakeGenerator) Generate() int64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return int64(g.node.Generate())
}

// ParseID extracts metadata from a Snowflake ID
func (g *SnowflakeGenerator) ParseID(id int64) SnowflakeMetadata {
	sfID := snowflake.ID(id)
	return SnowflakeMetadata{
		ID:        id,
		Timestamp: sfID.Time(),
		NodeID:    sfID.Node(),
		Step:      sfID.Step(),
	}
}

// SnowflakeMetadata contains parsed information from a Snowflake ID
type SnowflakeMetadata struct {
	ID        int64 // The original ID
	Timestamp int64 // Unix timestamp in seconds
	NodeID    int64 // The node/machine ID that generated this ID
	Step      int64 // The sequence number within the millisecond
}
