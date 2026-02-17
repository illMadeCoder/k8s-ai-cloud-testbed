use std::sync::Arc;
use std::time::Instant;

use axum::extract::{Path, State};
use axum::http::{header, StatusCode};
use axum::response::IntoResponse;
use axum::Json;
use serde::{Deserialize, Serialize};

use crate::metrics;
use crate::store::FileStore;

#[derive(Deserialize)]
pub struct WriteRequest {
    value: i32,
}

#[derive(Serialize)]
pub struct WriteResponse {
    row_id: u64,
}

#[derive(Serialize)]
pub struct ReadResponse {
    value: i32,
}

pub async fn write(
    State(store): State<Arc<FileStore>>,
    Json(req): Json<WriteRequest>,
) -> Result<Json<WriteResponse>, StatusCode> {
    let start = Instant::now();

    let result = tokio::task::spawn_blocking(move || store.write(req.value))
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    metrics::WRITE_DURATION.observe(start.elapsed().as_secs_f64());
    metrics::FSYNC_DURATION.observe(result.fsync_secs);
    metrics::OPS_TOTAL.with_label_values(&["write"]).inc();
    metrics::ROWS_TOTAL.set((result.row_id + 1) as f64);

    Ok(Json(WriteResponse {
        row_id: result.row_id,
    }))
}

pub async fn read(
    State(store): State<Arc<FileStore>>,
    Path(row_id): Path<u64>,
) -> Result<Json<ReadResponse>, StatusCode> {
    let start = Instant::now();

    let value = store.read(row_id).ok_or(StatusCode::NOT_FOUND)?;

    metrics::READ_DURATION.observe(start.elapsed().as_secs_f64());
    metrics::OPS_TOTAL.with_label_values(&["read"]).inc();

    Ok(Json(ReadResponse { value }))
}

pub async fn health() -> &'static str {
    "OK"
}

pub async fn ready() -> &'static str {
    "OK"
}

pub async fn serve_metrics() -> impl IntoResponse {
    (
        [(
            header::CONTENT_TYPE,
            "text/plain; version=0.0.4; charset=utf-8",
        )],
        metrics::encode(),
    )
}
