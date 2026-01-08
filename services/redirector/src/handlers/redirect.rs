use axum::{
    extract::{Path, State},
    response::Redirect,
};
use std::sync::Arc;

use crate::errors::app_error::AppError;
use crate::services::url_service::UrlService;

pub async fn redirect_handler(
    Path(short_code): Path<String>,
    State(service): State<Arc<UrlService>>,
) -> Result<Redirect, AppError> {
    let original_url = service.get_redirect_url(&short_code).await?;
    Ok(Redirect::temporary(&original_url))
}
