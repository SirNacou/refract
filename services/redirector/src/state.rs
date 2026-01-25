use std::sync::Arc;

use crate::{
    cache::Cache, events::publisher::ClickEventPublisher, geo::lookup::GeoLookup,
    parser::user_agent::UAParser, repository::postgres::Repository,
};

#[derive(Clone)]
pub struct AppState {
    pub db: Arc<dyn Repository + Send + Sync>,
    pub cache: Arc<dyn Cache + Send + Sync>,
    pub publisher: Arc<ClickEventPublisher>,
}

impl AppState {
    pub fn new(
        db: Arc<dyn Repository + Send + Sync>,
        cache: Arc<dyn Cache + Send + Sync>,
        publisher: Arc<ClickEventPublisher>,
    ) -> Self {
        AppState {
            db,
            cache,
            publisher,
        }
    }
}
