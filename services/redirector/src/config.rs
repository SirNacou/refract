use envconfig::Envconfig;

#[derive(Envconfig)]
pub struct Config {
    #[envconfig(from = "REDIRECTOR_PORT", default = "3000")]
    pub port: i32,
    #[envconfig(from = "REDIRECTOR_BASE_URL", default = "http://localhost:3000")]
    pub base_url: String,
    #[envconfig(nested)]
    pub database: DatabaseConfig,
    #[envconfig(nested)]
    pub redis: RedisConfig,
    #[envconfig(nested)]
    pub events: EventsConfig,
    #[envconfig(from = "REDIRECTOR_L1_CACHE_CAPACITY", default = "1000")]
    pub l1_cache_capacity: i32,
    #[envconfig(from = "REDIRECTOR_WORKER_ID", default = "worker-1")]
    pub worker_id: String,
    #[envconfig(from = "GEOIP_DB_PATH", default = "/path/to/geoip.db")]
    pub geoip_db_path: String,
    #[envconfig(from = "UA_REGEXES_PATH", default = "./data/user_agent/regexes.yaml")]
    pub ua_regexes_path: String,
    #[envconfig(from = "LOG_LEVEL", default = "info")]
    pub log_level: String,
    #[envconfig(from = "LOG_FORMAT", default = "json")]
    pub log_format: String,
}

#[derive(Envconfig)]
pub struct DatabaseConfig {
    #[envconfig(from = "REDIRECTOR_DATABASE_URL")]
    pub url: String,
    #[envconfig(from = "REDIRECTOR_DATABASE_MAX_CONNECTIONS", default = "25")]
    pub max_connections: u32,
    #[envconfig(from = "REDIRECTOR_DATABASE_MIN_CONNECTIONS", default = "5")]
    pub min_connections: u32,
    #[envconfig(from = "REDIRECTOR_CONNECTION_MAX_IDLE_TIME", default = "5")]
    pub max_idle_time: u64,
    #[envconfig(from = "REDIRECTOR_CONNECTION_MAX_LIFETIME", default = "5")]
    pub max_lifetime: u64,
}

#[derive(Envconfig)]
pub struct RedisConfig {
    #[envconfig(from = "REDIS_HOST", default = "localhost")]
    pub host: String,
    #[envconfig(from = "REDIS_PORT", default = "6379")]
    pub port: i32,
    #[envconfig(from = "REDIS_PASSWORD", default = "")]
    pub password: String,
    #[envconfig(from = "REDIS_DB", default = "0")]
    pub db: i32,
}

impl RedisConfig {
    pub fn to_redis_url(&self) -> String {
        if self.password.is_empty() {
            format!("redis://{}:{}/{}", self.host, self.port, self.db)
        } else {
            format!(
                "redis://:{}@{}:{}/{}",
                self.password, self.host, self.port, self.db
            )
        }
    }
}

#[derive(Envconfig)]
pub struct EventsConfig {
    #[envconfig(from = "REDIS_STREAM_KEY", default = "refract:click_events")]
    pub stream_key: String,
    #[envconfig(from = "EVENTS_BATCH_SIZE", default = "100")]
    pub batch_size: usize,
    #[envconfig(from = "EVENTS_FLUSH_INTERVAL_MS", default = "1000")]
    pub flush_interval_ms: u64,
    #[envconfig(from = "EVENTS_MAX_STREAM_LEN", default = "1000000")]
    pub max_stream_len: usize,
    #[envconfig(from = "EVENTS_MAX_BUFFER_SIZE", default = "10000")]
    pub max_buffer_size: usize,
}
