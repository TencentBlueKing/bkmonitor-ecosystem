// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//! HTTP 服务的 HelloWorld 接口实现。

use std::sync::OnceLock;

use axum::{
    extract::State,
    http::{HeaderMap, StatusCode},
    routing::get,
    Router,
};
use opentelemetry::{
    global,
    metrics::ObservableGauge,
    trace::{Status, TraceContextExt},
    KeyValue,
};
use opentelemetry_http::HeaderExtractor;
use rand::Rng;
use tracing_opentelemetry::OpenTelemetrySpanExt;

static MEMORY_USAGE: OnceLock<ObservableGauge<f64>> = OnceLock::new();

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum Country {
    UnitedStates,
    Canada,
    UnitedKingdom,
    Germany,
    France,
    Japan,
    Australia,
    China,
    India,
    Brazil,
}

impl Country {
    pub const fn as_str(self) -> &'static str {
        match self {
            Self::UnitedStates => "United States",
            Self::Canada => "Canada",
            Self::UnitedKingdom => "United Kingdom",
            Self::Germany => "Germany",
            Self::France => "France",
            Self::Japan => "Japan",
            Self::Australia => "Australia",
            Self::China => "China",
            Self::India => "India",
            Self::Brazil => "Brazil",
        }
    }
}

pub const COUNTRIES: [Country; 10] = [
    Country::UnitedStates,
    Country::Canada,
    Country::UnitedKingdom,
    Country::Germany,
    Country::France,
    Country::Japan,
    Country::Australia,
    Country::China,
    Country::India,
    Country::Brazil,
];

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum DemoError {
    MySqlConnectTimeout,
    UserNotFound,
    NetworkUnreachable,
    FileNotFound,
}

impl DemoError {
    pub const fn as_str(self) -> &'static str {
        match self {
            Self::MySqlConnectTimeout => "mysql connect timeout",
            Self::UserNotFound => "user not found",
            Self::NetworkUnreachable => "network unreachable",
            Self::FileNotFound => "file not found",
        }
    }
}

impl std::fmt::Display for DemoError {
    fn fmt(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        formatter.write_str(self.as_str())
    }
}

impl std::error::Error for DemoError {}

#[derive(Clone, Copy, Debug)]
pub struct AppState {
    pub error_rate: f64,
}

pub fn app(state: AppState) -> Router {
    Router::new()
        .route("/helloworld", get(hello_world))
        .route("/healthz", get(healthz))
        .with_state(state)
}

async fn healthz() -> StatusCode {
    StatusCode::NO_CONTENT
}

// hello_world 处理 HTTP 请求并返回问候语
async fn hello_world(
    State(state): State<AppState>,
    headers: HeaderMap,
) -> Result<String, (StatusCode, String)> {
    let parent_context = global::get_text_map_propagator(|propagator| {
        propagator.extract(&HeaderExtractor(&headers))
    });
    let span = tracing::info_span!("Handle/HelloWorld", otel.kind = "server");
    if let Err(error) = span.set_parent(parent_context) {
        tracing::warn!(%error, "设置服务端调用链父上下文失败");
    }
    let _entered = span.enter();

    // Logs（日志）
    logs_demo();

    let mut rng = rand::rng();
    let country = COUNTRIES[rng.random_range(0..COUNTRIES.len())];
    tracing::info!(country = country.as_str(), "选择国家");

    // Metrics（指标） - Counter 类型
    metrics_counter_demo(country.as_str());
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
    if rng.random::<f64>() < state.error_rate {
        let error = traces_random_error_demo(&mut rng);
        return Err((StatusCode::INTERNAL_SERVER_ERROR, error.as_str().to_owned()));
    }

    Ok(format!("Hello World, {}!", country.as_str()))
}

fn do_something(max_ms: u64) {
    let sleep_ms = 10 + rand::rng().random_range(0..max_ms);
    std::thread::sleep(std::time::Duration::from_millis(sleep_ms));
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
    let async_span = tracing::info_span!("SpanLinkDemo/asyncCaller");
    let async_context = async_span.context();
    tracing::Span::current().add_link_with_attributes(
        async_context.span().span_context().clone(),
        vec![KeyValue::new("link_type", "async")],
    );
    tracing::info!("SpanLinkDemo async caller");
}

// traces_random_error_demo Traces（调用链）- 异常事件、状态
// Refer: https://docs.rs/opentelemetry/0.32.0/opentelemetry/trace/trait.Span.html#method.record_error
fn traces_random_error_demo(rng: &mut impl Rng) -> DemoError {
    let error = DEMO_ERRORS[rng.random_range(0..DEMO_ERRORS.len())];
    let span = tracing::Span::current();
    span.set_status(Status::error(error.to_string()));
    span.context().span().record_error(&error);
    tracing::error!(%error, "[traces_random_error_demo] got error");
    error
}

const DEMO_ERRORS: [DemoError; 4] = [
    DemoError::MySqlConnectTimeout,
    DemoError::UserNotFound,
    DemoError::NetworkUnreachable,
    DemoError::FileNotFound,
];
