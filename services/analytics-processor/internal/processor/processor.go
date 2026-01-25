package processor

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/SirNacou/refract/services/analytics-processor/internal/config"
	"github.com/SirNacou/refract/services/analytics-processor/internal/domain"
	"github.com/SirNacou/refract/services/analytics-processor/internal/geo"
	"github.com/SirNacou/refract/services/analytics-processor/internal/repository"
	"github.com/SirNacou/refract/services/analytics-processor/internal/useragent"
	"github.com/valkey-io/valkey-go"
)

const flushTimeout = 5 * time.Second

type Processor struct {
	valkey       valkey.Client
	repo         *repository.PostgresRepository
	geo          *geo.GeoLookup
	uap          *useragent.UserAgentParser
	batchSize    int64
	blockMs      int64
	streamKey    string
	groupName    string
	consumerName string
	salt         string
}

func NewProcessor(valkeyClient valkey.Client, repo *repository.PostgresRepository, geoLookup *geo.GeoLookup, uap *useragent.UserAgentParser, cfg *config.Config) *Processor {
	return &Processor{
		valkey:       valkeyClient,
		repo:         repo,
		geo:          geoLookup,
		uap:          uap,
		batchSize:    cfg.ANALYTICS_BATCH_SIZE,
		blockMs:      cfg.ANALYTICS_BLOCK_MS,
		streamKey:    cfg.REDIS_STREAM_KEY,
		groupName:    cfg.ANALYTICS_CONSUMER_GROUP,
		consumerName: cfg.ANALYTICS_CONSUMER_NAME,
		salt:         cfg.SECURITY_HMAC_SECRET,
	}
}

func (p *Processor) processEvent(ctx context.Context, event ClickEvent) (*domain.ClickEvent, error) {
	geoInfo := p.geo.Lookup(event.IPAddress)
	ua := p.uap.Parse(event.UserAgent)
	browser := ua.Browser.String()
	os := ua.OS.String()
	deviceType := ua.Device.String()

	parsedIP := net.ParseIP(event.IPAddress)
	if parsedIP == nil {
		slog.WarnContext(ctx, "invalid IP address", "ip_address", event.IPAddress)
		return nil, fmt.Errorf("invalid IP address: %s", event.IPAddress)
	}

	visitorHash := getVisitorID(event.IPAddress, p.salt)

	slog.DebugContext(ctx, "processed event",
		"event_id", event.EventID,
		"url_id", event.URLID,
		"timestamp", event.Timestamp,
		"short_code", event.ShortCode,
		"referrer", event.Referrer,
		"visitor_hash", visitorHash,
		"country_code", geoInfo.CountryCode,
		"country_name", geoInfo.CountryName,
		"city", geoInfo.City,
		"browser", browser,
		"operating_system", os,
		"device_type", deviceType,
		"user_agent", event.UserAgent,
		"latitude", geoInfo.Latitude,
		"longitude", geoInfo.Longitude,
	)

	return &domain.ClickEvent{
		EventID:         event.EventID,
		URLID:           event.URLID,
		Timestamp:       event.Timestamp,
		ShortCode:       event.ShortCode,
		Referrer:        event.Referrer,
		VisitorHash:     visitorHash,
		CountryCode:     geoInfo.CountryCode,
		CountryName:     geoInfo.CountryName,
		City:            geoInfo.City,
		Browser:         &browser,
		OperatingSystem: &os,
		DeviceType:      deviceType,
		UserAgent:       event.UserAgent,
		Latitude:        geoInfo.Latitude,
		Longitude:       geoInfo.Longitude,
	}, nil
}

func getVisitorID(ip, salt string) string {
	h := hmac.New(sha256.New, []byte(salt))
	h.Write([]byte(ip))
	return hex.EncodeToString(h.Sum(nil))
}

func (p *Processor) Run(ctx context.Context) error {
	slog.InfoContext(ctx, "processor starting",
		"stream_key", p.streamKey,
		"consumer_group", p.groupName,
		"consumer_name", p.consumerName,
		"batch_size", p.batchSize)

	if err := p.ensureGroupExists(ctx); err != nil {
		return fmt.Errorf("Failed to ensure consumer group existed: %v", err)
	}

	batch := &BatchData{}
	ticker := time.NewTicker(flushTimeout)
	defer ticker.Stop()
	for {
		slog.DebugContext(ctx, "reading batch")

		select {
		case <-ctx.Done():
			if len(batch.events) > 0 {
				if err := p.flushAndAckBatch(ctx, batch); err != nil {
					return fmt.Errorf("failed to flush final batch: %v", err)
				}
			}
			return ctx.Err()

		case <-ticker.C:
			if len(batch.events) > 0 {
				if err := p.flushAndAckBatch(ctx, batch); err != nil {
					return fmt.Errorf("failed to flush batch on ticker: %v", err)
				}
				batch = nil
			}
		default:
		}

		entries, err := p.readBatch(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to read batch: %v", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if len(entries) == 0 {
			// No entries, continue
			continue
		}

		for _, entry := range entries {
			event, err := p.processEvent(ctx, entry.clickEvent)
			if err != nil {
				return fmt.Errorf("failed to process event: %v", err)
			}
			batch.events = append(batch.events, *event)
			batch.ids = append(batch.ids, entry.entryID)
		}

		// If batch size reached or timeout exceeded, flush batch
		if int64(len(batch.events)) >= p.batchSize {
			if err := p.flushAndAckBatch(ctx, batch); err != nil {
				slog.ErrorContext(ctx, "failed to flush batch: %v", "error", err)
				continue
			}

			// Reset batch
			batch = &BatchData{}
		}
	}

}

func (p *Processor) ensureGroupExists(ctx context.Context) error {
	cmd := p.valkey.B().
		XgroupCreate().
		Key(p.streamKey).
		Group(p.groupName).
		Id("$").
		Mkstream().
		Build()

	err := p.valkey.Do(ctx, cmd).Error()

	if err != nil {
		if strings.Contains(err.Error(), "BUSYGROUP") {
			return nil
		}

		return err
	}

	return nil
}

func (p *Processor) flushAndAckBatch(ctx context.Context, batch *BatchData) error {
	if err := p.repo.InsertClickEvents(ctx, batch.events); err != nil {
		return fmt.Errorf("Failed to insert click events: %v", err)
	}

	return p.ackEntries(ctx, batch.ids)
}

func (p *Processor) readBatch(ctx context.Context) ([]StreamEntry, error) {
	cmd := p.valkey.B().
		Xreadgroup().
		Group(p.groupName, p.consumerName).
		Count(p.batchSize).
		Block(p.blockMs).
		Streams().
		Key(p.streamKey).
		Id(">").
		Build()

	res, err := p.valkey.Do(ctx, cmd).AsXRead()
	if err != nil {
		return nil, err
	}

	entries := res[p.streamKey]
	if len(entries) == 0 {
		return nil, nil
	}

	var events []StreamEntry
	for _, entry := range entries {
		event, err := parseClickEvent(entry)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse click event: %v", err)
		}
		events = append(events, *event)
	}

	return events, nil
}

func parseClickEvent(entry valkey.XRangeEntry) (*StreamEntry, error) {
	data, existed := entry.FieldValues["data"]
	if !existed {
		return nil, fmt.Errorf("missing \"data\" field in stream entry")
	}

	var clickEvent ClickEvent
	err := json.Unmarshal([]byte(data), &clickEvent)
	if err != nil {
		return nil, err
	}

	res := &StreamEntry{
		entryID:    entry.ID,
		clickEvent: clickEvent,
	}

	return res, nil
}

func (p *Processor) ackEntries(ctx context.Context, ids []string) error {
	cmd := p.valkey.B().
		Xack().
		Key(p.streamKey).
		Group(p.groupName).
		Id(ids...).
		Build()

	return p.valkey.Do(ctx, cmd).Error()
}
