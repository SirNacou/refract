CREATE TABLE IF NOT EXISTS refract.clicks ( 
  short_code String, 
  clicked_at DateTime64(3), -- FROM redirector (event time) 
  ip_address IPv6, 
  user_agent String DEFAULT '', 
  referer String DEFAULT '', 
  
  -- System fields 
  ingested_at DateTime64(3) DEFAULT now64(3), -- Auto-generated (processing time) 

  -- Derived fields 
  date Date DEFAULT toDate(clicked_at), 
  hour DateTime DEFAULT toStartOfHour(clicked_at) 
  ) 
ENGINE = MergeTree() 
PARTITION BY toYYYYMMDD(clicked_at)
ORDER BY (short_code, clicked_at, ip_address) TTL clicked_at + INTERVAL 30 DAY DELETE;

CREATE TABLE IF NOT EXISTS refract.url_daily_stats (
  date Date,
  short_code String,
  clicks AggregateFunction(sum, UInt64),
  unique_ips AggregateFunction(uniq, IPv6),
  first_click_at AggregateFunction(min, DateTime64(3)),
  last_click_at AggregateFunction(max, DateTime64(3))
) 
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (short_code, date)
SETTINGS index_granularity = 8192;

CREATE MATERIALIZED VIEW IF NOT EXISTS refract.url_daily_stats_mv
TO refract.url_daily_stats
AS SELECT
  toDate(clicked_at) as date,
  short_code,
  sumState(toUInt64(1)) as clicks,  -- Use -State suffix
  uniqState(ip_address) as unique_ips,
  minState(clicked_at) as first_click_at,
  maxState(clicked_at) as last_click_at
FROM refract.clicks
GROUP BY date, short_code;