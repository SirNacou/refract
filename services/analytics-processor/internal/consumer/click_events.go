package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/SirNacou/refract/services/analytics-processor/internal/config"
	"github.com/cenkalti/backoff/v5"
	"github.com/valkey-io/valkey-go"
)

// ClickEvent represents a single redirect click event (matches Rust publisher schema)
type ClickEvent struct {
	EventID         string   `json:"event_id"`
	URLID           int64    `json:"url_id"`
	ShortCode       string   `json:"short_code"`
	Timestamp       string   `json:"timestamp"` // TODO: parse to time.Time after validating format
	UserAgent       string   `json:"user_agent"`
	IPAddress       string   `json:"ip_address"` // TODO: parse to net.IP after validating format
	Referrer        *string  `json:"referrer,omitempty"`
	CountryCode     *string  `json:"country_code,omitempty"`
	CountryName     *string  `json:"country_name,omitempty"`
	City            *string  `json:"city,omitempty"`
	Latitude        *float64 `json:"latitude,omitempty"`
	Longitude       *float64 `json:"longitude,omitempty"`
	DeviceType      string   `json:"device_type"`
	Browser         *string  `json:"browser,omitempty"`
	OperatingSystem *string  `json:"operating_system,omitempty"`
	CacheTier       string   `json:"cache_tier"`
	LatencyMS       float64  `json:"latency_ms"`
	RequestID       string   `json:"request_id"`
}

// StreamEntry represents a single Redis Stream message
type StreamEntry struct {
	ID     string
	Fields map[string]string
}

// BatchHandler processes a batch of click events
// Returns error if processing fails (consumer will NOT ack)
type BatchHandler func(ctx context.Context, events []ClickEvent) error

// StreamConsumer consumes click events from Redis Streams using consumer groups
type StreamConsumer struct {
	client valkey.Client

	streamKey string
	group     string
	consumer  string

	batchSize int64
	blockMs   int64
	startID   string

	minBackoff time.Duration
	maxBackoff time.Duration
}

// NewStreamConsumer creates a new Redis Stream consumer
func NewStreamConsumer(
	client valkey.Client,
	cfg *config.Config,
) *StreamConsumer {
	return &StreamConsumer{
		client:     client,
		streamKey:  cfg.REDIS_STREAM_KEY,
		group:      cfg.ANALYTICS_CONSUMER_GROUP,
		consumer:   cfg.ANALYTICS_CONSUMER_NAME,
		batchSize:  cfg.ANALYTICS_BATCH_SIZE,
		blockMs:    cfg.ANALYTICS_BLOCK_MS,
		startID:    cfg.ANALYTICS_STREAM_START,
		maxBackoff: time.Duration(cfg.ANALYTICS_RETRY_MAX_BACKOFF_MS) * time.Millisecond,
	}
}

// EnsureGroup creates the consumer group if it doesn't exist (idempotent)
func (c *StreamConsumer) EnsureGroup(ctx context.Context) error {
	slog.Info("ensuring consumer group exists",
		"stream", c.streamKey,
		"group", c.group,
		"start_id", c.startID,
	)

	// XGROUP CREATE <stream> <group> <startID> MKSTREAM
	cmd := c.client.B().
		XgroupCreate().
		Key(c.streamKey).
		Group(c.group).
		Id(c.startID).
		Mkstream().
		Build()

	err := c.client.Do(ctx, cmd).Error()
	if err != nil {
		// BUSYGROUP means group already exists - that's okay
		if valkey.IsValkeyBusyGroup(err) {
			slog.Info("consumer group already exists", "group", c.group)
			return nil
		}
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	slog.Info("consumer group created successfully", "group", c.group)
	return nil
}

// Run starts the consumer loop (blocks until context cancelled)
func (c *StreamConsumer) Run(ctx context.Context, handler BatchHandler) error {
	slog.Info("starting consumer loop",
		"stream", c.streamKey,
		"group", c.group,
		"consumer", c.consumer,
		"batch_size", c.batchSize,
		"block_ms", c.blockMs,
	)

	for {
		select {
		case <-ctx.Done():
			slog.Info("consumer loop stopped", "reason", ctx.Err())
			return ctx.Err()
		default:
		}

		op := func() ([]StreamEntry, error) {
			return c.readBatch(ctx)
		}
		entries, err := backoff.Retry(ctx, op, backoff.WithMaxElapsedTime(c.maxBackoff))
		if err != nil {
			slog.Error("failed to read batch from stream",
				"error", err,
			)
			continue
		}

		if len(entries) == 0 {
			// No messages (block timeout) - continue immediately
			continue
		}

		slog.Debug("received batch from stream", "count", len(entries))

		events, ids, err := decodeClickEvents(entries)
		if err != nil {
			slog.Error("failed to decode click events (NOT acking)",
				"error", err,
				"entry_count", len(entries),
			)
			// Do NOT ack - leave messages pending so we notice schema issues
			time.Sleep(time.Second)
			continue
		}

		if len(events) == 0 {
			// All entries failed decode (already logged above)
			continue
		}

		slog.Info("decoded click events", "count", len(events))

		// Call handler (T041 will insert to DB)
		if err := handler(ctx, events); err != nil {
			slog.Error("handler failed (NOT acking)",
				"error", err,
				"event_count", len(events),
			)
			// Do NOT ack - at-least-once semantics
			time.Sleep(time.Second)
			continue
		}

		// Handler succeeded - ACK all messages
		if err := c.ack(ctx, ids); err != nil {
			slog.Error("failed to ack messages (duplicates possible)",
				"error", err,
				"id_count", len(ids),
			)
			// Continue anyway - messages were processed, duplicates are acceptable
		} else {
			slog.Debug("acked messages", "count", len(ids))
		}
	}
}

// readBatch reads a batch of messages using XREADGROUP
func (c *StreamConsumer) readBatch(ctx context.Context) ([]StreamEntry, error) {
	// XREADGROUP GROUP <group> <consumer> COUNT <batch> BLOCK <block> STREAMS <stream> >
	cmd := c.client.B().
		Xreadgroup().
		Group(c.group, c.consumer).
		Count(c.batchSize).
		Block(c.blockMs).
		Streams().
		Key(c.streamKey).
		Id(">").
		Build()

	result, err := c.client.Do(ctx, cmd).AsXRead()
	if err != nil {
		// Check for timeout (not an error, just no messages)
		if isTimeoutError(err) {
			return []StreamEntry{}, nil
		}
		return nil, fmt.Errorf("xreadgroup failed: %w", err)
	}

	// Parse the stream entries
	entries := make([]StreamEntry, 0, 128)
	for streamKey, streamEntries := range result {
		// Sanity check
		if streamKey != c.streamKey {
			return nil, fmt.Errorf("unexpected stream %q (expected %q)", streamKey, c.streamKey)
		}

		for _, entry := range streamEntries {
			fields := make(map[string]string, len(entry.FieldValues))
			for k, v := range entry.FieldValues {
				// valkey-go returns string values directly
				fields[k] = v
			}
			entries = append(entries, StreamEntry{
				ID:     entry.ID,
				Fields: fields,
			})
		}
	}

	return entries, nil
}

// ack acknowledges processed messages
func (c *StreamConsumer) ack(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// XACK <stream> <group> <id1> <id2> ...
	cmd := c.client.B().Xack().Key(c.streamKey).Group(c.group).Id(ids...).Build()
	return c.client.Do(ctx, cmd).Error()
}

// decodeClickEvents decodes stream entries into ClickEvent structs
// Returns decoded events + corresponding IDs for ACK
func decodeClickEvents(entries []StreamEntry) ([]ClickEvent, []string, error) {
	events := make([]ClickEvent, 0, len(entries))
	ids := make([]string, 0, len(entries))

	for _, e := range entries {
		data, ok := e.Fields["data"]
		if !ok || data == "" {
			// Strict fail during learning - we need to notice schema issues
			return nil, nil, fmt.Errorf("missing data field for stream entry id=%s", e.ID)
		}

		var evt ClickEvent
		if err := json.Unmarshal([]byte(data), &evt); err != nil {
			return nil, nil, fmt.Errorf("bad click event json id=%s: %w", e.ID, err)
		}

		events = append(events, evt)
		ids = append(ids, e.ID)
	}

	return events, ids, nil
}

// isTimeoutError checks if error is a Redis timeout (BLOCK expired)
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "nil") || strings.Contains(errStr, "timeout")
}
