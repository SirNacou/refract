use serde::Deserialize;

#[derive(Debug, Clone, Deserialize)]
pub struct Config {
    pub server: ServerConfig,
    pub database: DatabaseConfig,
    pub cache: CacheConfig,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ServerConfig {
    #[serde(default = "default_host")]
    pub host: String,
    #[serde(default = "default_port")]
    pub port: u16,
}

#[derive(Debug, Clone, Deserialize)]
pub struct DatabaseConfig {
    pub host: String,
    pub port: u16,
    pub user: String,
    pub password: String,
    pub name: String,
    #[serde(default = "default_max_connections")]
    pub max_connections: u32,
}

#[derive(Debug, Clone, Deserialize)]
pub struct CacheConfig {
    // L2 cache (Valkey/Redis)
    pub host: String,
    pub port: u16,
    pub password: String,
    #[serde(default)]
    pub db: u8,
    
    // L1 cache (Moka in-memory)
    #[serde(default = "default_l1_max_capacity")]
    pub l1_max_capacity: u64,
    #[serde(default = "default_l1_ttl_seconds")]
    pub l1_ttl_seconds: u64,
}

fn default_l1_max_capacity() -> u64 {
    10_000  // 10K entries ≈ 1-2MB
}

fn default_l1_ttl_seconds() -> u64 {
    60  // 1 minute
}

fn default_host() -> String {
    "0.0.0.0".to_string()
}

fn default_port() -> u16 {
    8081
}

fn default_max_connections() -> u32 {
    10
}

impl Config {
    pub fn from_env() -> Result<Self, config::ConfigError> {
        let server = config::Config::builder()
            .add_source(
                config::Environment::default()
                    .prefix("REDIRECTOR")
                    .separator("_"),
            )
            .build()?
            .try_deserialize::<ServerConfig>()?;

        let database = config::Config::builder()
            .add_source(config::Environment::default().prefix("DB").separator("_"))
            .build()?
            .try_deserialize::<DatabaseConfig>()?;

        let cache = config::Config::builder()
            .add_source(
                config::Environment::default()
                    .prefix("VALKEY")
                    .separator("_"),
            )
            .build()?
            .try_deserialize::<CacheConfig>()?;

        Ok(Config {
            server,
            database,
            cache,
        })
    }

    pub fn database_url(&self) -> String {
        format!(
            "postgres://{}:{}@{}:{}/{}",
            self.database.user,
            self.database.password,
            self.database.host,
            self.database.port,
            self.database.name
        )
    }

    pub fn cache_url(&self) -> String {
        format!(
            "redis://:{}@{}:{}/{}",
            self.cache.password, self.cache.host, self.cache.port, self.cache.db
        )
    }
    
    pub fn redis_url(&self) -> String {
        // Reuse cache_url() for multi-tier-cache compatibility
        self.cache_url()
    }
}
