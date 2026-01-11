-- Rollback: 00004_create_timescale_hypertables
-- Warning: This will delete all click analytics data

-- Remove continuous aggregate policies first
SELECT remove_continuous_aggregate_policy('click_summary_daily', if_exists => true);
SELECT remove_continuous_aggregate_policy('click_summary_hourly', if_exists => true);

-- Drop continuous aggregates (order matters: daily depends on hourly)
DROP MATERIALIZED VIEW IF EXISTS click_summary_daily;
DROP MATERIALIZED VIEW IF EXISTS click_summary_hourly;

-- Remove hypertable policies
SELECT remove_retention_policy('click_events', if_exists => true);
SELECT remove_compression_policy('click_events', if_exists => true);

-- Drop the hypertable (this also drops chunks and indexes)
DROP TABLE IF EXISTS click_events;

-- Note: We don't drop the timescaledb extension as other tables may use it