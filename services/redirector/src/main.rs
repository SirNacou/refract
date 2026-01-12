use axum::Router;
use envconfig::Envconfig;
use tokio::signal;

use crate::config::Config;

mod cache;
mod config;
mod events;
mod geo;
mod parser;
mod repository;

#[tokio::main]
async fn main() {
    let config = Config::init_from_env().unwrap();

    let db = repository::postgres::PostgresRepository::new(&config.database)
        .await
        .expect("Failed to initialize Postgres repository");

    let cache = cache::RedirectCache::new(&config.redis)
        .await
        .expect("Failed to initialize Cache");

    let geo_lookup = geo::lookup::GeoLookup::new(&config.geoip_db_path)
        .expect("Failed to initialize Geo Lookup");

    let ua_parser = parser::user_agent::UAParser::new(&config.ua_regexes_path)
        .expect("Failed to initialize User Agent Parser");

    let app = Router::new();

    let addr = format!("{}:{}", "0.0.0.0", config.port);
    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .expect("Failed to bind to address");

    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await
        .unwrap();
}

async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }
}
