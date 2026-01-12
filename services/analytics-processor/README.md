# Analytics Processor (T040 Implementation)

The analytics processor is a background service that consumes click events from Redis Streams and processes them for analytics aggregation.

## Architecture

This service implements **T040: Create Redis Stream consumer** with the following design:

### Consumer Group Pattern
- Uses Redis Streams consumer groups for distributed processing
- Multiple processor instances can share load automatically
- At-least-once delivery semantics (messages not ACKed if processing fails)

### Components

1. **Config** (`internal/config/config.go`)
   - Loads configuration from environment variables
   - Supports all `.env.example` settings

2. **Consumer** (`internal/consumer/click_events.go`)
   - `StreamConsumer`: Main consumer loop with exponential backoff retry
   - `readBatch()`: Executes `XREADGROUP` to fetch batches
   - `decodeClickEvents()`: Parses JSON from `data` field into Go structs
   - `ack()`: Acknowledges processed messages

3. **Main** (`cmd/processor/main.go`)
   - Wires up consumer with graceful shutdown
   - Placeholder handler (T041 will add TimescaleDB insert)

## Key Design Decisions

### Start from `$` (new messages only)
- Default `ANALYTICS_STREAM_START=$` means only process new messages
- Prevents replay storms in dev/prod
- Can be changed to `0` to replay from beginning if needed

### Retry forever with backoff
- Redis errors trigger exponential backoff (250ms → 5s cap)
- Consumer never crashes on transient failures
- Logs all errors for observability

### Strict error handling (fail fast on schema issues)
- Missing `data` field → error, do NOT ACK
- Invalid JSON → error, do NOT ACK
- Messages remain in PEL (Pending Entries List) for manual inspection

### Batch processing
- Default batch size: 100 events
- Block time: 1000ms (1 second)
- Balances throughput vs latency for 5-second analytics SLA

## Environment Variables

See `.env.example` for full list. Key variables:

```bash
# Redis Stream
REDIS_STREAM_KEY=refract:click_events
ANALYTICS_CONSUMER_GROUP=analytics-processor
ANALYTICS_CONSUMER_NAME=processor-1  # Auto-set to hostname
ANALYTICS_BATCH_SIZE=100
ANALYTICS_BLOCK_MS=1000
ANALYTICS_STREAM_START=$  # $ = new messages, 0 = replay all

# Retry backoff
ANALYTICS_RETRY_MIN_BACKOFF_MS=250
ANALYTICS_RETRY_MAX_BACKOFF_MS=5000
```

## Running

### Local Development

```bash
# Start dependencies (Redis, PostgreSQL)
just up

# Run processor
cd services/analytics-processor
go run ./cmd/processor/
```

### Testing

```bash
cd services/analytics-processor
go test -v ./internal/consumer/
```

Tests cover:
- ✅ JSON decoding with all fields (required + optional)
- ✅ Missing `data` field error handling
- ✅ Invalid JSON error handling
- ✅ Empty batch handling
- ✅ Timeout error detection

## Event Schema

The consumer expects stream entries with a single field `data` containing JSON:

```json
{
  "event_id": "uuid-v7",
  "url_id": 42,
  "short_code": "abc123",
  "timestamp": "2024-01-13T12:00:00Z",
  "user_agent": "Mozilla/5.0",
  "ip_address": "192.168.1.1",
  "referrer": "https://example.com",  // optional
  "country_code": "US",               // optional
  "country_name": "United States",    // optional
  "city": "New York",                 // optional
  "latitude": 40.7128,                // optional
  "longitude": -74.0060,              // optional
  "device_type": "desktop",
  "browser": "Chrome",                // optional
  "operating_system": "Windows",      // optional
  "cache_tier": "l1",
  "latency_ms": 5.2,
  "request_id": "req-123"
}
```

This schema matches the Rust `ClickEvent` in `services/redirector/src/events/mod.rs`.

## How It Works

### Startup
1. Load config from environment
2. Connect to Redis (valkey-go client)
3. Create consumer group (idempotent: `XGROUP CREATE ... MKSTREAM`)
4. Start consumer loop

### Read Loop
1. `XREADGROUP GROUP <group> <consumer> COUNT 100 BLOCK 1000 STREAMS <stream> >`
2. Parse nested array response → extract `data` field
3. JSON unmarshal → `[]ClickEvent`
4. Call handler function (placeholder for now, T041 will add DB insert)
5. If handler succeeds → `XACK` all message IDs
6. If handler fails → do NOT ack (at-least-once semantics)

### Graceful Shutdown
- Listens for `SIGINT` / `SIGTERM`
- Cancels context → consumer loop exits cleanly
- In-flight batch is NOT acked (safe to replay)

## Next Steps (T041)

The handler is currently a placeholder that logs events. T041 will:
- Create TimescaleDB repository
- Batch insert events into `click_events` hypertable
- Handle duplicate IDs (at-least-once → idempotency)

## Operational Notes

### Monitoring
- All operations logged with structured JSON
- Key metrics: batch size, decode errors, handler failures, ack failures

### Pending Messages (PEL)
- If processor crashes after read but before ack, messages remain pending
- They won't be re-delivered automatically (reading with `>` gets new messages only)
- Later: add `XAUTOCLAIM` to reclaim old pending messages (follow-up task)

### Scaling
- Run multiple instances with different `ANALYTICS_CONSUMER_NAME`
- Redis distributes messages across consumers in the same group
- Each consumer processes different messages (no coordination needed)

## Troubleshooting

### "BUSYGROUP" error
- Harmless: consumer group already exists
- Consumer logs "consumer group already exists" and continues

### Messages not being consumed
- Check: Is redirector publishing? `redis-cli XRANGE refract:click_events - + COUNT 5`
- Check: Consumer group exists? `redis-cli XINFO GROUPS refract:click_events`
- Check: Any pending? `redis-cli XPENDING refract:click_events analytics-processor`

### Schema errors
- Look for "missing data field" or "bad click event json" in logs
- Check message: `redis-cli XRANGE refract:click_events <id> <id>`
- Fix schema, manually ack bad message: `redis-cli XACK refract:click_events analytics-processor <id>`

## References

- Spec: `specs/016-authenticated-url-shortener/spec.md` (FR-020 to FR-027)
- Plan: `specs/016-authenticated-url-shortener/plan.md` (Analytics Processor section)
- Tasks: `specs/016-authenticated-url-shortener/tasks.md` (T040)
- Constitution: `.specify/memory/constitution.md` (Principle IV: at-least-once delivery)
