//! 日志埋点。
//!
//! 本模块通过 `tracing` 写入结构化日志；初始化阶段注册的日志桥接层会将其
//! 转换为 OpenTelemetry Log，并由批处理导出器异步上报。

/// 写入请求日志和一条带自定义属性的业务日志。
pub fn logs_demo() {
    tracing::info!("收到请求：GET /helloworld");
    tracing::info!(
        method = "GET",
        k1 = "v1",
        k2 = 123,
        "上报带自定义属性的请求日志"
    );
}
