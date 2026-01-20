use crate::{
    cache::{self, RedirectCache},
    geo::{self, lookup::GeoLookup},
    parser::{self, user_agent::UAParser},
    repository::{self, postgres::PostgresRepository},
};

pub struct AppState {
    db: PostgresRepository,
    cache: RedirectCache,
    geo_lookup: GeoLookup,
    ua_parser: UAParser,
}

impl AppState {
    pub fn new(
        db: PostgresRepository,
        cache: RedirectCache,
        geo_lookup: GeoLookup,
        ua_parser: UAParser,
    ) -> Self {
        AppState {
            db,
            cache,
            geo_lookup,
            ua_parser,
        }
    }
    pub fn db(&self) -> &PostgresRepository {
        &self.db
    }

    pub fn cache(&self) -> &RedirectCache {
        &self.cache
    }

    pub fn geo_lookup(&self) -> &GeoLookup {
        &self.geo_lookup
    }

    pub fn ua_parser(&self) -> &UAParser {
        &self.ua_parser
    }
}
