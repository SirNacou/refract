use std::time::Duration;

use axum::{
    extract::{Path, State},
    response::{Redirect, Result},
};
use axum_client_ip::ClientIp;
use axum_extra::{
    TypedHeader,
    headers::{Referer, UserAgent},
};
use multi_tier_cache::CacheStrategy;
use sqlx::types::chrono::{DateTime, Utc};
use tracing::error;

use crate::{cache, events::ClickEvent, handlers::AppError, state::AppState};

pub async fn handle(
    Path(short_code): Path<String>,
    TypedHeader(user_agent): TypedHeader<UserAgent>,
    referer: Option<TypedHeader<Referer>>,
    ClientIp(ip_address): ClientIp,
    State(state): State<std::sync::Arc<AppState>>,
) -> Result<Redirect, AppError> {
    let url = state
        .cache
        .get_cache_manager()
        .get(&get_redirect_cache_key(&short_code))
        .await
        .map_err(|err| AppError::Internal(err))?;

    if let Some(cached_url) = url {
        return Ok(Redirect::temporary(&cached_url.to_string()));
    }

    struct Url {
        snowflake_id: i64,
        destination_url: String,
        expires_at: Option<DateTime<Utc>>,
    }

    let url = sqlx::query_as!(
        Url,
        "
        select snowflake_id, destination_url, expires_at
        from urls 
        where short_code = $1 AND (status = 'active') AND (expires_at IS NULL OR expires_at > NOW())
        ",
        &short_code
    )
    .fetch_one(state.db.pool())
    .await
    .map_err(|e| {
        error!(
            "Error fetching redirect for short code {}: {}",
            short_code, e
        );
        AppError::NotFound(format!("Url not found for short code: {}", short_code))
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
        None => cache::DEFAULT_TTL,
    };

    let _ = state
        .cache
        .get_cache_manager()
        .set_with_strategy(
            &get_redirect_cache_key(&short_code),
            url.destination_url.clone().into(),
            CacheStrategy::Custom(Duration::from_secs(cache_expiration)),
        )
        .await
        .inspect_err(|e| {
            error!(
                "Error caching redirect for short code {}: {}",
                short_code, e
            );
        });

    let click_event = ClickEvent::new(
        url.snowflake_id,
        &short_code,
        user_agent.as_str(),
        ip_address,
        referer.map(|r| r.0.to_string()),
    );
    state.publisher.publish(click_event);

    Ok(Redirect::temporary(&url.destination_url))
}

fn get_redirect_cache_key(short_code: &str) -> String {
    format!("redirect:{}", short_code)
}
