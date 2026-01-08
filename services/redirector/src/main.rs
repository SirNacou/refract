use axum::{routing::get, Router};
use sqlx::postgres::PgPoolOptions;
use std::sync::Arc;
use std::time::Duration;
use tower_http::timeout::TimeoutLayer;
use tracing_subscriber::EnvFilter;
use axum::http::StatusCode;

mod config;
mod errors;
mod handlers;
mod models;
mod repositories;
mod services;

use config::Config;
use repositories::{cache::CacheRepository, database::DatabaseRepository};
use services::url_service::UrlService;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Initialize logging
    tracing_subscriber::fmt()
        .with_env_filter(
            EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info")),
        )
        .json()
        .init();

    tracing::info!("Starting redirector service");

    // Load configuration
    let config = Config::from_env()?;
    tracing::info!("Configuration loaded successfully");

    // Initialize database pool
    let db_pool = PgPoolOptions::new()
        .max_connections(config.database.max_connections)
        .connect(&config.database_url())
        .await?;
    tracing::info!("Database connection pool initialized");

    // Initialize cache
    let cache = Arc::new(CacheRepository::new(&config.cache_url()).await?);
    tracing::info!("Cache connection established");

    // Initialize repositories
    let db_repo = Arc::new(DatabaseRepository::new(db_pool));

    // Initialize URL service with multi-tier cache
    let redis_url = config.redis_url();
    let l1_max_capacity = config.cache.l1_max_capacity;
    let l1_ttl = Duration::from_secs(config.cache.l1_ttl_seconds);

    let url_service = Arc::new(
        UrlService::new(&redis_url, db_repo.clone(), l1_max_capacity, l1_ttl).await?,
    );

    tracing::info!(
        redis_url = %redis_url,
        l1_max_capacity = l1_max_capacity,
        l1_ttl_seconds = l1_ttl.as_secs(),
        "URL service initialized with multi-tier cache"
    );

    // Build router
    let app = Router::new()
        .route("/health", get(handlers::health::health_handler))
        .with_state((cache.clone(), db_repo.clone()))
        .route("/:short_code", get(handlers::redirect::redirect_handler))
        .with_state(url_service)
        .layer(TimeoutLayer::with_status_code(
            StatusCode::GATEWAY_TIMEOUT,
            Duration::from_secs(1),
        ));

    // Start server
    let addr = format!("{}:{}", config.server.host, config.server.port);
    let listener = tokio::net::TcpListener::bind(&addr).await?;

    tracing::info!("Redirector service listening on {}", addr);

    axum::serve(listener, app).await?;

    Ok(())
}
