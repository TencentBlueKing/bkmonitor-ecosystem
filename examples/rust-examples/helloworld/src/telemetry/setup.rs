use std::collections::HashMap;

use opentelemetry::global;
use opentelemetry::trace::TracerProvider;
use opentelemetry_otlp::{Protocol, WithExportConfig, WithHttpConfig};
use opentelemetry_sdk::{
    logs::SdkLoggerProvider, metrics::SdkMeterProvider, propagation::TraceContextPropagator,
    trace::SdkTracerProvider, Resource,
};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use crate::config::AppConfig;
use crate::telemetry::metrics;

/// 持有各信号的 SDK Provider，确保退出时能刷新内存中尚未上报的数据。
pub struct TelemetryGuard {
    tracer_provider: SdkTracerProvider,
    meter_provider: SdkMeterProvider,
    logger_provider: SdkLoggerProvider,
}

impl TelemetryGuard {
    /// 按 Trace、Metric、Log 的顺序关闭。
    ///
    /// Log bridge 仍会记录前两个 Provider 的关闭过程，因此 Log Provider 必须最后关闭，
    /// 否则会产生“已关闭后仍写日志”的无效告警。
    pub fn shutdown(self) {
        let _ = self.tracer_provider.shutdown();
        let _ = self.meter_provider.shutdown();
        let _ = self.logger_provider.shutdown();
    }
}

pub fn setup(config: &AppConfig) -> Result<TelemetryGuard, Box<dyn std::error::Error>> {
    if config.otlp_exporter_type != "http" {
        return Err(format!(
            "此 Rust 示例只支持 OTLP HTTP/protobuf，OTLP_EXPORTER_TYPE 必须为 http，当前为 {}",
            config.otlp_exporter_type
        )
        .into());
    }
    // 使用 W3C TraceContext 格式在 HTTP 请求头中传递上下文。
    // 这会让内置 Client 的 Span 与 Server 的 Handle/HelloWorld Span 形成父子关系。
    global::set_text_map_propagator(TraceContextPropagator::new());

    // 三种信号共享同一 Resource，平台据此将数据归属到指定服务。
    // ❗❗【非常重要】SERVICE_NAME 必须与 APM 应用中的服务标识保持一致。
    let resource = Resource::builder()
        .with_service_name(config.service_name.clone())
        .build();
    // ❗❗【非常重要】Token 是 bk-collector 的鉴权凭证，不能写入代码或提交到仓库。
    // 每个 OTLP exporter 都必须携带相同的 x-bk-token 请求头。
    let headers = HashMap::from([("x-bk-token".to_owned(), config.token.clone())]);

    let tracer_provider = if config.enable_traces {
        // OTLP HTTP/protobuf 的完整上报路径为 /v1/traces。
        let exporter = opentelemetry_otlp::SpanExporter::builder()
            .with_http()
            .with_protocol(Protocol::HttpBinary)
            .with_endpoint(format!(
                "{}/v1/traces",
                config.otlp_endpoint.trim_end_matches('/')
            ))
            .with_headers(headers.clone())
            .build()?;
        // Batch processor 在后台线程批量发送 Span，不会让 HTTP 业务请求等待网络上报。
        SdkTracerProvider::builder()
            .with_batch_exporter(exporter)
            .with_resource(resource.clone())
            .build()
    } else {
        SdkTracerProvider::builder()
            .with_resource(resource.clone())
            .build()
    };

    let meter_provider = if config.enable_metrics {
        // 指标的 OTLP HTTP/protobuf 标准路径为 /v1/metrics。
        let exporter = opentelemetry_otlp::MetricExporter::builder()
            .with_http()
            .with_protocol(Protocol::HttpBinary)
            .with_endpoint(format!(
                "{}/v1/metrics",
                config.otlp_endpoint.trim_end_matches('/')
            ))
            .with_headers(headers.clone())
            .build()?;
        // Periodic reader 按周期采集并上报，不会阻塞记录指标的 add 调用。
        SdkMeterProvider::builder()
            .with_periodic_exporter(exporter)
            .with_resource(resource.clone())
            .build()
    } else {
        SdkMeterProvider::builder()
            .with_resource(resource.clone())
            .build()
    };

    let logger_provider = if config.enable_logs {
        // 日志的 OTLP HTTP/protobuf 标准路径为 /v1/logs。
        let exporter = opentelemetry_otlp::LogExporter::builder()
            .with_http()
            .with_protocol(Protocol::HttpBinary)
            .with_endpoint(format!(
                "{}/v1/logs",
                config.otlp_endpoint.trim_end_matches('/')
            ))
            .with_headers(headers)
            .build()?;
        // 日志先进入 batch 队列，再由后台线程发送给 collector。
        SdkLoggerProvider::builder()
            .with_batch_exporter(exporter)
            .with_resource(resource)
            .build()
    } else {
        SdkLoggerProvider::builder().with_resource(resource).build()
    };

    // 注册全局 Provider，使业务代码可通过 opentelemetry::global 获取统一实例。
    global::set_tracer_provider(tracer_provider.clone());
    global::set_meter_provider(meter_provider.clone());
    metrics::register_metrics_gauge_demo();
    // tracing Span 自动转换为 OTel Span；tracing 日志自动桥接为 OTel Log。
    let trace_layer =
        tracing_opentelemetry::layer().with_tracer(tracer_provider.tracer("helloworld"));
    let log_layer =
        opentelemetry_appender_tracing::layer::OpenTelemetryTracingBridge::new(&logger_provider);
    // fmt layer 同时保留本地控制台输出，便于未接 collector 时排查接入问题。
    tracing_subscriber::registry()
        .with(tracing_subscriber::fmt::layer())
        .with(trace_layer)
        .with(log_layer)
        .try_init()?;

    Ok(TelemetryGuard {
        tracer_provider,
        meter_provider,
        logger_provider,
    })
}
