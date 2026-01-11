-- Migration: 00004_create_timescale_hypertables
-- Description: Create TimescaleDB hypertable for click events and continuous aggregates
-- Created: 2026-01-11
-- Requires: TimescaleDB extension

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create click_events table
CREATE TABLE click_events (
    time TIMESTAMPTZ NOT NULL,
    event_id UUID NOT NULL DEFAULT gen_random_uuid(),
    url_id BIGINT NOT NULL,
    referrer TEXT,
    user_agent TEXT NOT NULL,
    ip_address INET NOT NULL,
    country_code CHAR(2),
    country_name TEXT,
    city TEXT,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    device_type TEXT NOT NULL
        CHECK (device_type IN ('desktop', 'mobile', 'tablet', 'bot')),
    browser TEXT,
    operating_system TEXT
);

-- Convert to hypertable (partitioned by time, 1 day chunks)
SELECT create_hypertable('click_events', 'time', chunk_time_interval => INTERVAL '1 day');

-- Enable compression
ALTER TABLE click_events SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'url_id',
    timescaledb.compress_orderby = 'time DESC'
);

-- Add compression policy: compress chunks older than 7 days
SELECT add_compression_policy('click_events', INTERVAL '7 days');

-- Add retention policy: drop chunks older than 5 years
SELECT add_retention_policy('click_events', INTERVAL '5 years');

-- Indexes for common queries
CREATE INDEX idx_click_events_url_time ON click_events (url_id, time DESC);
CREATE INDEX idx_click_events_country ON click_events (country_code, time DESC);

-- Continuous Aggregate: Hourly summary
CREATE MATERIALIZED VIEW click_summary_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS hour,
    url_id,
    COUNT(*) AS total_clicks,
    COUNT(DISTINCT ip_address) AS unique_visitors,
    COUNT(*) FILTER (WHERE device_type = 'mobile') AS mobile_clicks,
    COUNT(*) FILTER (WHERE device_type = 'desktop') AS desktop_clicks,
    COUNT(*) FILTER (WHERE device_type = 'tablet') AS tablet_clicks,
    COUNT(*) FILTER (WHERE device_type = 'bot') AS bot_clicks
FROM click_events
GROUP BY hour, url_id
WITH NO DATA;

-- Refresh policy for hourly aggregate: every 5 minutes
SELECT add_continuous_aggregate_policy('click_summary_hourly',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '5 minutes');

-- Continuous Aggregate: Daily summary (rolls up from hourly)
CREATE MATERIALIZED VIEW click_summary_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', hour) AS day,
    url_id,
    SUM(total_clicks) AS total_clicks,
    SUM(unique_visitors) AS unique_visitors,
    SUM(mobile_clicks) AS mobile_clicks,
    SUM(desktop_clicks) AS desktop_clicks
FROM click_summary_hourly
GROUP BY day, url_id
WITH NO DATA;

-- Refresh policy for daily aggregate: once per day
SELECT add_continuous_aggregate_policy('click_summary_daily',
    start_offset => INTERVAL '30 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day');