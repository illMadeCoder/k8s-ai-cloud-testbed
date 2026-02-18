use std::sync::LazyLock;

use prometheus::{CounterVec, Encoder, Gauge, Histogram, HistogramOpts, Opts, TextEncoder};

const BUCKETS: [f64; 11] = [
    0.000001, 0.000005, 0.00001, 0.00005, 0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1,
];

pub static WRITE_DURATION: LazyLock<Histogram> = LazyLock::new(|| {
    let h = Histogram::with_opts(
        HistogramOpts::new(
            "naivedb_write_duration_seconds",
            "Full write latency including fsync",
        )
        .buckets(BUCKETS.to_vec()),
    )
    .unwrap();
    prometheus::register(Box::new(h.clone())).unwrap();
    h
});

pub static READ_DURATION: LazyLock<Histogram> = LazyLock::new(|| {
    let h = Histogram::with_opts(
        HistogramOpts::new("naivedb_read_duration_seconds", "Read latency")
            .buckets(BUCKETS.to_vec()),
    )
    .unwrap();
    prometheus::register(Box::new(h.clone())).unwrap();
    h
});

pub static FSYNC_DURATION: LazyLock<Histogram> = LazyLock::new(|| {
    let h = Histogram::with_opts(
        HistogramOpts::new("naivedb_fsync_duration_seconds", "fsync-only latency")
            .buckets(BUCKETS.to_vec()),
    )
    .unwrap();
    prometheus::register(Box::new(h.clone())).unwrap();
    h
});

pub static BATCH_WRITE_DURATION: LazyLock<Histogram> = LazyLock::new(|| {
    let h = Histogram::with_opts(
        HistogramOpts::new(
            "naivedb_batch_write_duration_seconds",
            "Batch write latency including fsync",
        )
        .buckets(BUCKETS.to_vec()),
    )
    .unwrap();
    prometheus::register(Box::new(h.clone())).unwrap();
    h
});

pub static BATCH_SIZE: LazyLock<Histogram> = LazyLock::new(|| {
    let h = Histogram::with_opts(
        HistogramOpts::new("naivedb_batch_size", "Number of rows per batch write")
            .buckets(vec![1.0, 5.0, 10.0, 25.0, 50.0, 100.0, 250.0, 500.0, 1000.0]),
    )
    .unwrap();
    prometheus::register(Box::new(h.clone())).unwrap();
    h
});

pub static OPS_TOTAL: LazyLock<CounterVec> = LazyLock::new(|| {
    let c = CounterVec::new(
        Opts::new("naivedb_operations_total", "Operation count by type"),
        &["op"],
    )
    .unwrap();
    prometheus::register(Box::new(c.clone())).unwrap();
    c
});

pub static ROWS_TOTAL: LazyLock<Gauge> = LazyLock::new(|| {
    let g = Gauge::new("naivedb_rows_total", "Current row count").unwrap();
    prometheus::register(Box::new(g.clone())).unwrap();
    g
});

pub fn encode() -> String {
    let encoder = TextEncoder::new();
    let families = prometheus::gather();
    let mut buf = Vec::new();
    encoder.encode(&families, &mut buf).unwrap();
    String::from_utf8(buf).unwrap()
}
