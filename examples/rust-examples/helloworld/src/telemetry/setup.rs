// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

use std::collections::HashMap;

use opentelemetry::global;
use opentelemetry::trace::TracerProvider;
use opentelemetry_otlp::{Protocol, WithExportConfig, WithHttpConfig};
use opentelemetry_sdk::{
    logs::SdkLoggerProvider, metrics::SdkMeterProvider, propagation::TraceContextPropagator,
    trace::SdkTracerProvider, Resource,
};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use crate::{config::AppConfig, http::server::metrics_gauge_demo};

/// TelemetryGuard 定义 OTLP 服务的核心结构。
pub struct TelemetryGuard {
    tracer_provider: SdkTracerProvider,
    meter_provider: SdkMeterProvider,
    logger_provider: SdkLoggerProvider,
}

impl TelemetryGuard {
    /// shutdown 停止 OTLP 服务并清理资源。
    pub fn shutdown(self) {
        let _ = self.tracer_provider.shutdown();
        let _ = self.meter_provider.shutdown();
        let _ = self.logger_provider.shutdown();
    }
}

/// setup 启动 OTLP 服务并初始化各项功能。
pub fn setup(config: &AppConfig) -> Result<TelemetryGuard, Box<dyn std::error::Error>> {
    if config.otlp_exporter_type != "http" {
        return Err(format!(
            "此 Rust 示例只支持 OTLP HTTP/protobuf，OTLP_EXPORTER_TYPE 必须为 http，当前为 {}",
            config.otlp_exporter_type
        )
        .into());
    }
    global::set_text_map_propagator(TraceContextPropagator::new());

    let resource = Resource::builder()
        // ❗❗【非常重要】应用服务唯一标识
        .with_service_name(config.service_name.clone())
        .build();
    let headers = HashMap::from([("x-bk-token".to_owned(), config.token.clone())]);

    let tracer_provider = if config.enable_traces {
        let exporter = opentelemetry_otlp::SpanExporter::builder()
            .with_http()
            .with_protocol(Protocol::HttpBinary)
            // ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
            // 格式为 ip:port 或 domain:port，不要带 schema
            .with_endpoint(format!(
                "http://{}/v1/traces",
                config.otlp_endpoint.trim_end_matches('/')
            ))
            // ❗❗【非常重要】请传入应用 Token
            .with_headers(headers.clone())
            .build()?;
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
        let exporter = opentelemetry_otlp::MetricExporter::builder()
            .with_http()
            .with_protocol(Protocol::HttpBinary)
            // ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
            // 格式为 ip:port 或 domain:port，不要带 schema
            .with_endpoint(format!(
                "http://{}/v1/metrics",
                config.otlp_endpoint.trim_end_matches('/')
            ))
            // ❗❗【非常重要】请传入应用 Token
            .with_headers(headers.clone())
            .build()?;
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
        let exporter = opentelemetry_otlp::LogExporter::builder()
            .with_http()
            .with_protocol(Protocol::HttpBinary)
            // ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
            // 格式为 ip:port 或 domain:port，不要带 schema
            .with_endpoint(format!(
                "http://{}/v1/logs",
                config.otlp_endpoint.trim_end_matches('/')
            ))
            // ❗❗【非常重要】请传入应用 Token
            .with_headers(headers)
            .build()?;
        SdkLoggerProvider::builder()
            .with_batch_exporter(exporter)
            .with_resource(resource)
            .build()
    } else {
        SdkLoggerProvider::builder().with_resource(resource).build()
    };

    global::set_tracer_provider(tracer_provider.clone());
    global::set_meter_provider(meter_provider.clone());
    metrics_gauge_demo();
    let trace_layer =
        tracing_opentelemetry::layer().with_tracer(tracer_provider.tracer("helloworld"));
    let log_layer =
        opentelemetry_appender_tracing::layer::OpenTelemetryTracingBridge::new(&logger_provider);
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
