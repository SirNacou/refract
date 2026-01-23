use std::{net::SocketAddr, sync::Arc};

use axum::{Router, routing::get};
use axum_client_ip::ClientIpSource;
use redirector::{
    cache,
    config::{self},
    events::publisher,
    geo, handlers, parser, repository,
    state::AppState,
};
use tokio::signal;
use tracing::{error, info};

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();
    let config = config::get_config();

    let db = Arc::new(
        repository::postgres::PostgresRepository::new(&config.database)
            .await
            .expect("Failed to initialize Postgres repository"),
    );

    let cache = Arc::new(
        cache::RedirectCache::new(&config.redis)
            .await
            .expect("Failed to initialize Cache"),
    );

    let geo_lookup = Arc::new(
        geo::lookup::GeoLookup::new(&config.geoip_db_path)
            .expect("Failed to initialize Geo Lookup"),
    );

    let ua_parser = Arc::new(
        parser::user_agent::UAParser::new(&config.ua_regexes_path)
            .expect("Failed to initialize User Agent Parser"),
    );

    let click_event_publisher = Arc::new(
        publisher::ClickEventPublisher::new(&config.redis, &config.events)
            .await
            .inspect_err(|e| error!("Failed to initialize ClickEventPublisher: {}", e))
            .unwrap(),
    );

    click_event_publisher.start_flush_task();

    let app_state = Arc::new(AppState::new(
        db,
        cache,
        click_event_publisher,
        geo_lookup,
        ua_parser,
    ));

    let ip_source = match config.ip_source {
        config::IpSource::CfConnectingIp => ClientIpSource::CfConnectingIp,
        config::IpSource::XRealIp => ClientIpSource::XRealIp,
        config::IpSource::XForwardedFor => ClientIpSource::RightmostXForwardedFor,
        config::IpSource::ConnectInfo => ClientIpSource::ConnectInfo,
    };
    let app = Router::new()
        .route("/{short_code}", get(handlers::redirect::handle))
        .with_state(app_state)
        .layer(ip_source.into_extension());

    let addr = format!("{}:{}", "0.0.0.0", config.port);
    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .expect("Failed to bind to address");

    info!("Redirector service running on port {}", config.port);

    axum::serve(
        listener,
        app.into_make_service_with_connect_info::<SocketAddr>(),
    )
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
