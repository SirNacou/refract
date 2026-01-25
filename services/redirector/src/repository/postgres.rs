use std::time::Duration;

use sqlx::{PgPool, postgres::PgPoolOptions};

use crate::config::DatabaseConfig;

pub trait Repository {
    fn pool(&self) -> &PgPool;
}

pub struct PostgresRepository {
    pool: PgPool,
}

impl PostgresRepository {
    pub async fn new(cfg: &DatabaseConfig) -> Result<Self, sqlx::Error> {
        print!("Connecting to Postgres at {}", cfg.url);
        let pool = PgPoolOptions::new()
            .max_connections(cfg.max_connections)
            .min_connections(cfg.min_connections)
            .acquire_timeout(Duration::from_secs(5))
            .idle_timeout(Duration::from_secs(cfg.max_idle_time))
            .max_lifetime(Duration::from_secs(cfg.max_lifetime))
            .connect(&cfg.url)
            .await?;

        Ok(Self { pool })
    }
}

impl Repository for PostgresRepository {
    fn pool(&self) -> &PgPool {
        &self.pool
    }
}
