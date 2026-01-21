use axum::{Json, http::StatusCode, response::IntoResponse};
use serde_json::json;

pub mod redirect;

pub enum AppError {
    NotFound(String),
    Expired(String),
    Validation(String),
    Internal(anyhow::Error),
}

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let (status, error_message) = match self {
            AppError::NotFound(e) => (StatusCode::NOT_FOUND, e),
            AppError::Validation(e) => (StatusCode::BAD_REQUEST, e),
            AppError::Internal(e) => (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()),
            AppError::Expired(e) => (StatusCode::GONE, e),
        };

        let body = Json(json!({
            "error": {
                "message": error_message
            },
        }));

        (status, body).into_response()
    }
}
