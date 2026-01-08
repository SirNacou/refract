use axum::{
    http::StatusCode,
    response::{Html, IntoResponse, Response},
};
use thiserror::Error;
use tracing::error;

#[derive(Debug, Error)]
pub enum AppError {
    #[error("Not found: {0}")]
    NotFound(String),

    #[error("Database error: {0}")]
    Database(#[from] sqlx::Error),

    #[error("Cache error: {0}")]
    Cache(#[from] redis::RedisError),

    #[error("Internal error: {0}")]
    Internal(String),
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, error_page) = match self {
            AppError::NotFound(ref msg) => {
                // Log for debugging but return generic 404 to user
                error!(error = %msg, "Not found error");
                (
                    StatusCode::NOT_FOUND,
                    include_str!("../templates/error_404.html"),
                )
            }
            AppError::Database(ref e) => {
                error!(error = %e, "Database error");
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    include_str!("../templates/error_500.html"),
                )
            }
            AppError::Cache(ref e) => {
                error!(error = %e, "Cache error");
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    include_str!("../templates/error_500.html"),
                )
            }
            AppError::Internal(ref msg) => {
                error!(error = %msg, "Internal error");
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    include_str!("../templates/error_500.html"),
                )
            }
        };

        (status, Html(error_page)).into_response()
    }
}
