package repository

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/SirNacou/refract/services/analytics-processor/internal/consumer"
	"github.com/SirNacou/refract/services/analytics-processor/internal/geo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TimescaleRepository handles batch insertion of click events into TimescaleDB
type TimescaleRepository struct {
	pool        *pgxpool.Pool
	geoLookup   *geo.GeoLookup // optional: nil if geo enrichment disabled
	enrichGeo   bool           // track if we should attempt enrichment
	enriched    int64          // counter for enriched events (for metrics)
	notEnriched int64          // counter for events that already had geo data
}

// NewTimescaleRepository creates a new TimescaleDB repository
// geoLookup is optional - pass nil to disable geo enrichment fallback
func NewTimescaleRepository(pool *pgxpool.Pool, geoLookup *geo.GeoLookup) *TimescaleRepository {
	return &TimescaleRepository{
		pool:      pool,
		geoLookup: geoLookup,
		enrichGeo: geoLookup != nil,
	}
}

// InsertClickEvents inserts a batch of click events into the click_events hypertable
// Returns error if any event is invalid or insertion fails (fail-fast semantics)
func (r *TimescaleRepository) InsertClickEvents(ctx context.Context, events []consumer.ClickEvent) error {
	if len(events) == 0 {
		return nil
	}

	slog.Debug("inserting click events batch", "count", len(events))

	// Validate and build rows
	rows, err := r.buildRows(events)
	if err != nil {
		return fmt.Errorf("failed to build rows: %w", err)
	}

	// Use CopyFrom for efficient batch insert
	columns := []string{
		"time", "event_id", "url_id", "referrer", "user_agent", "ip_address",
		"country_code", "country_name", "city", "latitude", "longitude",
		"device_type", "browser", "operating_system",
	}

	copyCount, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"click_events"},
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("failed to copy rows: %w", err)
	}

	slog.Info("inserted click events batch",
		"count", copyCount,
		"expected", len(events),
	)

	if copyCount != int64(len(events)) {
		return fmt.Errorf("inserted row count mismatch: expected %d, got %d", len(events), copyCount)
	}

	return nil
}

// buildRows converts click events to row format for pgx.CopyFrom
// Validates required fields, parses timestamps/IPs, and enriches geo data if missing
func (r *TimescaleRepository) buildRows(events []consumer.ClickEvent) ([][]any, error) {
	rows := make([][]any, 0, len(events))

	for i, evt := range events {
		// Optionally enrich geo data if missing (fallback)
		if r.enrichGeo && evt.CountryCode == nil {
			if geoInfo := r.geoLookup.Lookup(evt.IPAddress); geoInfo != nil {
				evt.CountryCode = geoInfo.CountryCode
				evt.CountryName = geoInfo.CountryName
				evt.City = geoInfo.City
				evt.Latitude = geoInfo.Latitude
				evt.Longitude = geoInfo.Longitude
				r.enriched++
			}
		} else if evt.CountryCode != nil {
			r.notEnriched++
		}

		row, err := r.eventToRow(evt)
		if err != nil {
			return nil, fmt.Errorf("event %d (event_id=%s): %w", i, evt.EventID, err)
		}
		rows = append(rows, row)
	}

	// Log enrichment stats if geo is enabled
	if r.enrichGeo && r.enriched+r.notEnriched > 0 {
		slog.Debug("geo enrichment stats",
			"enriched", r.enriched,
			"already_had_geo", r.notEnriched,
		)
	}

	return rows, nil
}

// eventToRow converts a single ClickEvent to a row slice
func (r *TimescaleRepository) eventToRow(evt consumer.ClickEvent) ([]any, error) {
	// Parse timestamp (required, RFC3339)
	timestamp, err := time.Parse(time.RFC3339, evt.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp %q: %w", evt.Timestamp, err)
	}

	// Parse event_id (required, UUID)
	eventID, err := uuid.Parse(evt.EventID)
	if err != nil {
		return nil, fmt.Errorf("invalid event_id %q: %w", evt.EventID, err)
	}

	// Validate url_id (required, non-zero)
	if evt.URLID == 0 {
		return nil, fmt.Errorf("url_id is required")
	}

	// Parse IP address (required, inet)
	ipAddr := net.ParseIP(evt.IPAddress)
	if ipAddr == nil {
		return nil, fmt.Errorf("invalid ip_address %q", evt.IPAddress)
	}

	// Validate user_agent (required, non-empty)
	if evt.UserAgent == "" {
		return nil, fmt.Errorf("user_agent is required")
	}

	// Validate device_type (required, constrained values)
	if !isValidDeviceType(evt.DeviceType) {
		return nil, fmt.Errorf("invalid device_type %q (must be desktop|mobile|tablet|bot)", evt.DeviceType)
	}

	// Optional: validate country_code is 2 chars if present
	if evt.CountryCode != nil && len(*evt.CountryCode) != 2 {
		return nil, fmt.Errorf("invalid country_code %q (must be 2 characters)", *evt.CountryCode)
	}

	// Build row with nullable fields
	row := []any{
		timestamp,                     // time
		eventID,                       // event_id
		evt.URLID,                     // url_id
		ptrToAny(evt.Referrer),        // referrer (nullable)
		evt.UserAgent,                 // user_agent
		ipAddr.String(),               // ip_address (as string for inet)
		ptrToAny(evt.CountryCode),     // country_code (nullable)
		ptrToAny(evt.CountryName),     // country_name (nullable)
		ptrToAny(evt.City),            // city (nullable)
		ptrToFloat64(evt.Latitude),    // latitude (nullable)
		ptrToFloat64(evt.Longitude),   // longitude (nullable)
		evt.DeviceType,                // device_type
		ptrToAny(evt.Browser),         // browser (nullable)
		ptrToAny(evt.OperatingSystem), // operating_system (nullable)
	}

	return row, nil
}

// isValidDeviceType checks if device type matches migration constraint
func isValidDeviceType(dt string) bool {
	switch dt {
	case "desktop", "mobile", "tablet", "bot":
		return true
	default:
		return false
	}
}

// ptrToAny converts *string to any (nil or string value)
func ptrToAny(ptr *string) any {
	if ptr == nil {
		return nil
	}
	return *ptr
}

// ptrToFloat64 converts *float64 to any (nil or float64 value)
func ptrToFloat64(ptr *float64) any {
	if ptr == nil {
		return nil
	}
	return *ptr
}
