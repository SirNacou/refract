use sqlx::PgPool;

use crate::errors::app_error::AppError;
use crate::models::url::UrlRecord;

pub struct DatabaseRepository {
    pool: PgPool,
}

impl DatabaseRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }

    pub async fn get_url(&self, short_code: &str) -> Result<Option<UrlRecord>, AppError> {
        let record = sqlx::query_as::<_, UrlRecord>(
            r#"
            SELECT original_url, expires_at, is_active
            FROM urls
            WHERE short_code = $1
            LIMIT 1
            "#,
        )
        .bind(short_code)
        .fetch_optional(&self.pool)
        .await?;

        Ok(record)
    }

    pub async fn health_check(&self) -> Result<(), AppError> {
        sqlx::query("SELECT 1").execute(&self.pool).await?;
        Ok(())
    }
}
