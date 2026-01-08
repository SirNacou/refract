use chrono::Utc;
use multi_tier_cache::{CacheStrategy, CacheSystem};
use std::sync::Arc;
use std::time::Duration;
use tracing::{debug, info, warn};

use crate::errors::app_error::AppError;
use crate::models::url::UrlRecord;
use crate::repositories::database::DatabaseRepository;

pub struct UrlService {
    cache: Arc<CacheSystem>,
    db: Arc<DatabaseRepository>,
}

impl UrlService {
    pub async fn new(
        redis_url: &str,
        db: Arc<DatabaseRepository>,
        _l1_max_capacity: u64,
        _l1_ttl: Duration,
    ) -> Result<Self, AppError> {
        // Set Redis URL via environment variable (library expects REDIS_URL)
        std::env::set_var("REDIS_URL", redis_url);

        // Build cache system with L1 (Moka) + L2 (Redis/Valkey)
        // Note: L1 configuration uses library defaults (can be customized via MokaCacheConfig)
        let cache_system = CacheSystem::new()
            .await
            .map_err(|e| AppError::Internal(format!("Failed to initialize cache: {}", e)))?;

        info!(
            redis_url = redis_url,
            "Multi-tier cache initialized (L1: Moka, L2: Redis/Valkey)"
        );

        Ok(Self {
            cache: Arc::new(cache_system),
            db,
        })
    }

    pub async fn get_redirect_url(&self, short_code: &str) -> Result<String, AppError> {
        let start = std::time::Instant::now();
        let cache_key = Self::cache_key(short_code);

        // Use multi-tier-cache's get_or_compute_with for automatic:
        // - L1 check (Moka in-memory)
        // - L2 fallback (Redis/Valkey) with L1 promotion
        // - Stampede protection (DashMap coalescing)
        // - Database query on miss
        let db_clone = self.db.clone();
        let short_code_owned = short_code.to_string();

        let url = self
            .cache
            .cache_manager()
            .get_or_compute_with(&cache_key, CacheStrategy::Custom(Self::calculate_cache_ttl()), || async move {
                debug!(short_code = %short_code_owned, "Cache miss, querying database");

                // Query database
                let url_record = db_clone
                    .get_url(&short_code_owned)
                    .await
                    .map_err(|e| anyhow::anyhow!("Database error: {}", e))?
                    .ok_or_else(|| anyhow::anyhow!("Short code not found: {}", short_code_owned))?;

                // Validate (active + not expired)
                Self::validate_url(&url_record)
                    .map_err(|e| anyhow::anyhow!("{}", e))?;

                Ok::<serde_json::Value, anyhow::Error>(serde_json::json!(url_record.original_url))
            })
            .await
            .map_err(|e| {
                warn!(short_code = %short_code, error = %e, "Failed to get redirect URL");
                // Convert anyhow error back to AppError
                if e.to_string().contains("not found") {
                    AppError::NotFound(e.to_string())
                } else {
                    AppError::Internal(format!("Cache operation failed: {}", e))
                }
            })?;

        // Extract URL string from JSON value
        let url_string = url
            .as_str()
            .ok_or_else(|| AppError::Internal("Invalid cache data format".to_string()))?
            .to_string();

        let elapsed = start.elapsed();

        // Log with cache statistics
        let stats = self.cache.cache_manager().get_stats();
        info!(
            short_code = %short_code,
            original_url = %url_string,
            latency_ms = elapsed.as_millis() as u64,
            hit_rate_pct = format!("{:.1}", stats.hit_rate),
            l1_hits = stats.l1_hits,
            l2_hits = stats.l2_hits,
            misses = stats.misses,
            "Redirect successful"
        );

        Ok(url_string)
    }

    fn cache_key(short_code: &str) -> String {
        format!("url:redirect:{}", short_code)
    }

    fn validate_url(record: &UrlRecord) -> Result<(), AppError> {
        // Check active
        if !record.is_active {
            return Err(AppError::NotFound("URL is inactive".to_string()));
        }

        // Check expiration
        if record.expires_at <= Utc::now() {
            return Err(AppError::NotFound("URL has expired".to_string()));
        }

        Ok(())
    }

    fn calculate_cache_ttl() -> Duration {
        // Fixed 24-hour TTL for simplicity
        // multi-tier-cache handles per-tier TTL automatically
        Duration::from_secs(24 * 60 * 60)
    }

    pub fn get_cache_stats(&self) -> String {
        let stats = self.cache.cache_manager().get_stats();
        format!(
            "Hit rate: {:.2}%, L1: {} hits, L2: {} hits, Misses: {}, In-flight: {}",
            stats.hit_rate,
            stats.l1_hits,
            stats.l2_hits,
            stats.misses,
            stats.in_flight_requests
        )
    }
}
