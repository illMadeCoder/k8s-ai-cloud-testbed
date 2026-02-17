use std::sync::Arc;

use axum::{
    routing::{get, post},
    Router,
};
use tracing::info;

mod handlers;
mod metrics;
mod store;

use store::FileStore;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let data_dir = std::env::var("DATA_DIR").unwrap_or_else(|_| "/data".to_string());
    let port = std::env::var("PORT").unwrap_or_else(|_| "8080".to_string());

    let store = Arc::new(FileStore::open_or_create(&data_dir).expect("failed to open store"));
    info!(rows = store.row_count(), "store ready");
    metrics::ROWS_TOTAL.set(store.row_count() as f64);

    let app = Router::new()
        .route("/write", post(handlers::write))
        .route("/read/{row_id}", get(handlers::read))
        .route("/health", get(handlers::health))
        .route("/ready", get(handlers::ready))
        .route("/metrics", get(handlers::serve_metrics))
        .with_state(store);

    let addr = format!("0.0.0.0:{port}");
    info!(%addr, "listening");
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
