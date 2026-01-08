use chrono::{DateTime, Utc};
use sqlx::FromRow;

#[derive(Debug, Clone, FromRow)]
pub struct UrlRecord {
    pub original_url: String,
    pub expires_at: DateTime<Utc>,
    pub is_active: bool,
}
