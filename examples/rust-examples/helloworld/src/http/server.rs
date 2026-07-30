// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//! HTTP 服务的 HelloWorld 接口实现。

use std::sync::OnceLock;

use actix_web::{web, HttpResponse};
use opentelemetry::{
    global,
    metrics::ObservableGauge,
    trace::{Status, TraceContextExt},
    Context, KeyValue,
};
use rand::Rng;
use tracing::Instrument;
use tracing_opentelemetry::OpenTelemetrySpanExt;

static MEMORY_USAGE: OnceLock<ObservableGauge<f64>> = OnceLock::new();

const COUNTRIES: [&str; 10] = [
    "United States",
    "Canada",
    "United Kingdom",
    "Germany",
    "France",
    "Japan",
    "Australia",
    "China",
    "India",
    "Brazil",
];

const ERROR_MESSAGES: [&str; 4] = [
    "mysql connect timeout",
    "user not found",
    "network unreachable",
    "file not found",
];

pub fn configure_routes(config: &mut web::ServiceConfig) {
    config
        .route("/helloworld", web::get().to(hello_world))
        .route("/healthz", web::get().to(healthz));
}

async fn healthz() -> HttpResponse {
    HttpResponse::NoContent().finish()
}

// hello_world 处理 HTTP 请求并返回问候语
async fn hello_world() -> HttpResponse {
    let span = tracing::info_span!("Handle/HelloWorld");
    let _entered = span.enter();

    // Logs（日志）
    logs_demo();

    let mut rng = rand::rng();
    let country = choice_country(&mut rng);
    tracing::info!(country, "选择国家");

    // Metrics（指标） - Counter 类型
    metrics_counter_demo(country);
    // Metrics（指标） - Histograms 类型
    metrics_histogram_demo();
    // Metrics（指标） - 调用分析场景
    metrics_rpc_demo("server");
    metrics_rpc_demo("client");

    // Traces（调用链）- 自定义 Span
    traces_custom_span_demo();
    // Traces（调用链）- 在当前 Span 上设置自定义属性
    traces_set_custom_span_attributes();
    // Traces（调用链）- Span 事件
    traces_span_event_demo();
    // Traces（调用链）- Span Links
    traces_span_links_demo();
    // Traces（调用链）- 模拟错误
    if let Err(error) = traces_random_error_demo(&mut rng) {
        return HttpResponse::InternalServerError().body(error.to_string());
    }

    HttpResponse::Ok().body(generate_greeting(country))
}

fn do_something(max_ms: u64) {
    let sleep_ms = 10 + rand::rng().random_range(0..max_ms);
    std::thread::sleep(std::time::Duration::from_millis(sleep_ms));
}

fn choice_country(rng: &mut impl Rng) -> &'static str {
    COUNTRIES[rng.random_range(0..COUNTRIES.len())]
}

fn choice_error(rng: &mut impl Rng) -> std::io::Error {
    std::io::Error::other(ERROR_MESSAGES[rng.random_range(0..ERROR_MESSAGES.len())])
}

fn generate_greeting(country: &str) -> String {
    format!("Hello World, {country}!")
}

fn random_error(rng: &mut impl Rng, error_rate: f64) -> Option<std::io::Error> {
    (rng.random::<f64>() < error_rate).then(|| choice_error(rng))
}

// logs_demo Logs（日志）打印日志
// Refer: https://opentelemetry.io/docs/languages/rust/getting-started/#logs
fn logs_demo() {
    // 上报日志
    tracing::info!("收到请求：GET /helloworld");

    // 添加自定义属性
    tracing::info!(
        method = "GET",
        k1 = "v1",
        k2 = 123,
        "上报带自定义属性的请求日志"
    );
}

// metrics_counter_demo Metrics（指标）- 使用 Counter 类型指标
// Refer: https://opentelemetry.io/docs/specs/otel/metrics/api/#counter
fn metrics_counter_demo(country: &str) {
    global::meter("helloworld")
        .u64_counter("requests_total")
        .with_description("Total number of HTTP requests")
        .build()
        .add(1, &[KeyValue::new("country", country.to_owned())]);
}

// metrics_gauge_demo Metrics（指标）- 使用 Gauge 类型指标
// Refer: https://docs.rs/opentelemetry/0.32.0/opentelemetry/metrics/struct.ObservableGauge.html
pub(crate) fn metrics_gauge_demo() {
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

// metrics_histogram_demo Metrics（指标）- 使用 Histogram 类型指标
// Refer: https://docs.rs/opentelemetry/0.32.0/opentelemetry/metrics/struct.Histogram.html
fn metrics_histogram_demo() {
    let started_at = std::time::Instant::now();
    do_something(100);
    global::meter("helloworld")
        .f64_histogram("task_execute_duration_seconds")
        .with_description("Task execute duration in seconds")
        .build()
        .record(started_at.elapsed().as_secs_f64(), &[]);
}

// metrics_rpc_demo Metrics（指标）- 调用分析场景
// 基于该指标规范上报，可在 APM 服务使用「调用分析」功能，省去自行配置仪表盘、告警等工作。
// 本样例更多演示如何定义、上报调用分析指标，实际使用时，可在客户端调用前、服务端处理请求前后进行埋点，以得到真实的调用数据。
// Refer: https://opentelemetry.io/docs/specs/semconv/rpc/rpc-metrics/
fn metrics_rpc_demo(role: &str) {
    let started_at = std::time::Instant::now();
    do_something(100);
    let attributes = [
        // RPC 系统，支持自定义。
        KeyValue::new("rpc_system", "custom"),
        // 指标分组，server_metrics/client_metrics。
        KeyValue::new("scope_name", format!("{role}_metrics")),
        // 实例，部署 IP 地址。
        KeyValue::new("instance", "127.0.0.1"),
        // 环境类型，支持自定义，e.g. Production/Development/..。
        KeyValue::new("namespace", "Development"),
        // 环境名称，支持自定义。
        KeyValue::new("env_name", "dev"),
        // 主调服务。
        KeyValue::new("caller_server", "helloworld"),
        // 主调 Service，如果不区分服务/Service，可与 caller_server 保持一致。
        KeyValue::new("caller_service", "helloworld.timer"),
        // 主调接口。
        KeyValue::new("caller_method", "loop_query_hello_world"),
        // 被调服务。
        KeyValue::new("callee_server", "helloworld"),
        // 被调 Service，如果不区分服务/Service，可与 callee_server 保持一致。
        KeyValue::new("callee_service", "helloworld.http"),
        // 被调接口。
        KeyValue::new("callee_method", "/helloworld"),
        // 返回码，支持自定义。
        KeyValue::new("code", "200"),
        // 返回码类型，可选：success / timeout / exception。
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

// traces_custom_span_demo Traces（调用链）- 增加自定义 Span
// Refer: https://opentelemetry.io/docs/languages/rust/getting-started/#traces
fn traces_custom_span_demo() {
    let span = tracing::info_span!("CustomSpanDemo/doSomething");
    let _entered = span.enter();

    // 增加 Span 自定义属性
    // Refer: https://docs.rs/tracing-opentelemetry/0.33.0/tracing_opentelemetry/trait.OpenTelemetrySpanExt.html#method.set_attribute
    span.set_attribute("helloworld.kind", 1_i64);
    span.set_attribute("helloworld.step", "traces_custom_span_demo");

    do_something(50);
}

// traces_set_custom_span_attributes Traces（调用链）- 在当前 Span 上设置自定义属性
// Refer: https://docs.rs/tracing/0.1.44/tracing/struct.Span.html#method.current
fn traces_set_custom_span_attributes() {
    let span = tracing::Span::current();
    span.set_attribute("api_name", "ApiRequest");
    span.set_attribute("act_id", 12345_i64);
}

// traces_span_event_demo Traces（调用链）- Span 事件
// Refer: https://docs.rs/tracing-opentelemetry/0.33.0/tracing_opentelemetry/trait.OpenTelemetrySpanExt.html#method.add_event
fn traces_span_event_demo() {
    let span = tracing::info_span!("SpanEventDemo/doSomething");
    let _entered = span.enter();
    let attributes = vec![
        KeyValue::new("helloworld.kind", 2_i64),
        KeyValue::new("helloworld.step", "traces_span_event_demo"),
    ];

    span.add_event("Before doSomething", attributes.clone());
    do_something(50);
    span.add_event("After doSomething", attributes);
}

// traces_span_links_demo Traces（调用链）- Span Links
// Refer: https://opentelemetry.io/docs/specs/otel/trace/api/#specifying-links
fn traces_span_links_demo() {
    let caller = tracing::info_span!("SpanLinkDemo/asyncCaller", otel.kind = "producer");
    let caller_context = caller.context();

    let fanout_count = rand::rng().random_range(0_i64..3);
    for link_index in 1..=fanout_count {
        let callee = tracing::info_span!(
            "SpanLinkDemo/asyncCallee",
            otel.kind = "consumer",
            relation.index = link_index
        );
        if let Err(error) = callee.set_parent(Context::new()) {
            tracing::warn!(%error, "设置 Span Link Callee 为 Root Trace 失败");
            continue;
        }
        callee.add_link_with_attributes(
            caller_context.span().span_context().clone(),
            vec![
                KeyValue::new("relation.step", "SpanLinkDemo"),
                KeyValue::new("relation.index", link_index),
            ],
        );

        tokio::spawn(
            async {
                traces_custom_span_demo();
            }
            .instrument(callee),
        );
    }
}

// traces_random_error_demo Traces（调用链）- 异常事件、状态
// Refer: https://docs.rs/opentelemetry/0.32.0/opentelemetry/trace/trait.Span.html#method.record_error
fn traces_random_error_demo(rng: &mut impl Rng) -> Result<(), std::io::Error> {
    let Some(error) = random_error(rng, 0.1) else {
        return Ok(());
    };
    let span = tracing::Span::current();
    span.set_status(Status::error(error.to_string()));
    span.context().span().record_error(&error);
    tracing::error!(%error, "[traces_random_error_demo] got error");
    Err(error)
}
