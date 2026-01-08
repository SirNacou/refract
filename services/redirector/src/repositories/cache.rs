use redis::{aio::ConnectionManager, AsyncCommands, Client};
use std::time::Duration;

use crate::errors::app_error::AppError;

pub struct CacheRepository {
    conn: ConnectionManager,
}

impl CacheRepository {
    pub async fn new(redis_url: &str) -> Result<Self, AppError> {
        let client = Client::open(redis_url)?;
        let conn = ConnectionManager::new(client).await?;
        Ok(Self { conn })
    }

    pub async fn get(&self, key: &str) -> Result<Option<String>, AppError> {
        let mut conn = self.conn.clone();
        let value: Option<String> = conn.get(key).await?;
        Ok(value)
    }

    pub async fn set(&self, key: &str, value: &str, ttl: Duration) -> Result<(), AppError> {
        let mut conn = self.conn.clone();
        conn.set_ex::<_, _, ()>(key, value, ttl.as_secs()).await?;
        Ok(())
    }

    pub async fn health_check(&self) -> Result<(), AppError> {
        let mut conn = self.conn.clone();
        redis::cmd("PING").query_async::<()>(&mut conn).await?;
        Ok(())
    }
}
