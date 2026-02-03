CREATE TABLE IF NOT EXISTS refract.urls (
  short_code String,
  original_url String,
  updated_at DateTime64(3) DEFAULT now64(3),
  is_deleted Bool DEFAULT false,  -- Soft delete flag
  
  -- Optional metadata for analytics
  created_by String DEFAULT '',
  tags Array(String) DEFAULT []
) 
ENGINE = ReplacingMergeTree(updated_at)  -- Keeps latest version based on updated_at
ORDER BY short_code
SETTINGS index_granularity = 8192;