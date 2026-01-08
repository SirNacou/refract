use axum::{extract::State, Json};
use serde_json::{json, Value};
use std::sync::Arc;

use crate::repositories::{cache::CacheRepository, database::DatabaseRepository};

pub async fn health_handler(
    State((cache, db)): State<(Arc<CacheRepository>, Arc<DatabaseRepository>)>,
) -> Json<Value> {
    let cache_status = match cache.health_check().await {
        Ok(_) => "ok",
        Err(_) => "error",
    };

    let db_status = match db.health_check().await {
        Ok(_) => "ok",
        Err(_) => "error",
    };

    let overall_status = if cache_status == "ok" && db_status == "ok" {
        "healthy"
    } else {
        "degraded"
    };

    Json(json!({
        "status": overall_status,
        "checks": {
            "database": db_status,
            "cache": cache_status
        }
    }))
}
