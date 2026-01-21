use std::time::Duration;

use axum::{
    extract::{Path, State},
    response::{Redirect, Result},
};
use sqlx::types::chrono::{DateTime, Utc};
use tracing::error;

use crate::{handlers::AppError, state::AppState};

pub async fn handle(
    Path(short_code): Path<String>,
    State(state): State<std::sync::Arc<AppState>>,
) -> Result<Redirect, AppError> {
    let url = state
        .cache()
        .get_cache_manager()
        .get(&get_redirect_cache_key(&short_code))
        .await
        .map_err(|err| AppError::Internal(err))?;

    if let Some(cached_url) = url {
        if let serde_json::Value::String(destination_url) = cached_url {
            return Ok(Redirect::temporary(&destination_url));
        }
    }

    #[derive(sqlx::FromRow)]
    struct Url {
        destination_url: String,
        expires_at: Option<DateTime<Utc>>,
    }

    let url = sqlx::query_as::<_, Url>(
        "
        select destination_url, expires_at
        from urls 
        where short_code = $1 AND (status = 'active') AND (expires_at IS NULL OR expires_at > NOW())
        ",
    )
    .bind(&short_code)
    .fetch_one(state.db().pool())
    .await
    .map_err(|e| {
        error!(
            "Error fetching redirect for short code {}: {}",
            short_code, e
        );
        AppError::Expired(format!("Expired redirect for short code: {}", short_code).to_string())
    })?;

    let cache_expiration = match url.expires_at {
        Some(exp) => {
            let now = Utc::now();
            let duration = exp.signed_duration_since(now);
            if duration.num_seconds() <= 0 {
                return Err(AppError::Expired(format!("Expired url")));
            }
            duration.num_seconds() as u64
        }
        None => 3600, // Default to 1 hour if no expiration is set
    };

    let _ = state
        .cache()
        .get_cache_manager()
        .set_with_strategy(
            &get_redirect_cache_key(&short_code),
            serde_json::Value::String(url.destination_url.clone()),
            multi_tier_cache::CacheStrategy::Custom(Duration::from_secs(cache_expiration)),
        )
        .await
        .inspect_err(|e| {
            error!(
                "Error caching redirect for short code {}: {}",
                short_code, e
            );
        });

    Ok(Redirect::temporary(&url.destination_url))
}

fn get_redirect_cache_key(short_code: &str) -> String {
    format!("redirect:{}", short_code)
}
