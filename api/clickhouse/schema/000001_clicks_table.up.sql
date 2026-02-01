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
ORDER BY (short_code, clicked_at, ip_address) TTL clicked_at + INTERVAL 90 DAY DELETE;