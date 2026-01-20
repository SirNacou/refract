use anyhow::Ok;
use axum::{
    extract::{Path, State},
    response::{Redirect, Result},
};

use crate::{handlers::AppError, state::AppState};

pub async fn handle(
    Path(short_code): Path<String>,
    State(state): State<std::sync::Arc<AppState>>,
) -> Result<Redirect, AppError> {
    let url = state
        .cache()
        .get_cache_manager()
        .get(&get_redirect_cache_key(&short_code))
        .await
        .map_err(|err| AppError(anyhow::format_err!(err)))?;

    Ok(Redirect::temporary(&url.unwrap().to_string()))
}

fn get_redirect_cache_key(short_code: &str) -> String {
    format!("redirect:{}", short_code)
}
