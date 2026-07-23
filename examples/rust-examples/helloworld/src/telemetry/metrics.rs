//! 指标埋点。
//!
//! 指标记录先进入 SDK 的聚合器，再由周期性导出器批量发送；业务请求不会等待
//! 网络上报完成。

use std::sync::OnceLock;

use opentelemetry::{global, metrics::ObservableGauge, KeyValue};
use rand::Rng;

/// 调用分析指标使用的固定维度。
pub const CALL_ANALYSIS_ATTRIBUTE_KEYS: [&str; 13] = [
    "rpc_system",
    "scope_name",
    "instance",
    "namespace",
    "env_name",
    "caller_server",
    "caller_service",
    "caller_method",
    "callee_server",
    "callee_service",
    "callee_method",
    "code",
    "code_type",
];

static MEMORY_USAGE: OnceLock<ObservableGauge<f64>> = OnceLock::new();

/// 按国家记录请求总数。
pub fn metrics_counter_demo(country: &str) {
    global::meter("helloworld")
        .u64_counter("requests_total")
        .with_description("Total number of HTTP requests")
        .build()
        .add(1, &[KeyValue::new("country", country.to_owned())]);
}

/// 注册内存使用率观测值；每个采集周期生成一个示例值。
pub fn register_metrics_gauge_demo() {
    MEMORY_USAGE.get_or_init(|| {
        global::meter("helloworld")
            .f64_observable_gauge("memory_usage")
            .with_description("Memory usage")
            .with_callback(|observer| {
                observer.observe(0.1 + rand::rng().random_range(0.0..0.2), &[]);
            })
            .build()
    });
}

/// 记录单次任务执行耗时。
pub fn metrics_histogram_demo() {
    let started_at = std::time::Instant::now();
    std::thread::sleep(std::time::Duration::from_millis(
        rand::rng().random_range(10..110),
    ));
    global::meter("helloworld")
        .f64_histogram("task_execute_duration_seconds")
        .with_description("Task execute duration in seconds")
        .build()
        .record(started_at.elapsed().as_secs_f64(), &[]);
}

/// 记录调用双方的处理次数和耗时。
pub fn metrics_rpc_demo(role: &str) {
    let started_at = std::time::Instant::now();
    std::thread::sleep(std::time::Duration::from_millis(
        rand::rng().random_range(10..110),
    ));
    let attributes = [
        KeyValue::new("rpc_system", "custom"),
        KeyValue::new("scope_name", format!("{role}_metrics")),
        KeyValue::new("instance", "127.0.0.1"),
        KeyValue::new("namespace", "Development"),
        KeyValue::new("env_name", "dev"),
        KeyValue::new("caller_server", "helloworld"),
        KeyValue::new("caller_service", "helloworld.timer"),
        KeyValue::new("caller_method", "loopQueryHelloWorld"),
        KeyValue::new("callee_server", "helloworld"),
        KeyValue::new("callee_service", "helloworld.http"),
        KeyValue::new("callee_method", "/helloworld"),
        KeyValue::new("code", "200"),
        KeyValue::new("code_type", "success"),
    ];
    let meter = global::meter("helloworld");
    let duration = started_at.elapsed().as_secs_f64();
    meter
        .u64_counter(format!("rpc_{role}_handled_total"))
        .build()
        .add(1, &attributes);
    meter
        .f64_histogram(format!("rpc_{role}_handled_seconds"))
        .build()
        .record(duration, &attributes);
}
