//! HTTP 请求处理与 HelloWorld 示例数据。

use axum::{
    extract::State,
    http::{HeaderMap, StatusCode},
    routing::get,
    Router,
};
use opentelemetry::global;
use opentelemetry_http::HeaderExtractor;
use rand::Rng;
use tracing_opentelemetry::OpenTelemetrySpanExt;

use crate::telemetry::{logs, metrics, traces};

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

async fn hello_world(
    State(state): State<AppState>,
    headers: HeaderMap,
) -> Result<String, (StatusCode, String)> {
    // 从请求头提取 Client 注入的上下文，使服务端 Span 归入同一条调用链。
    let parent_context = global::get_text_map_propagator(|propagator| {
        propagator.extract(&HeaderExtractor(&headers))
    });
    let span = tracing::info_span!("Handle/HelloWorld", otel.kind = "server");
    if let Err(error) = span.set_parent(parent_context) {
        tracing::warn!(%error, "设置服务端调用链父上下文失败");
    }
    let _entered = span.enter();

    // 输出两条关联到当前调用链的日志。
    logs::logs_demo();

    let mut rng = rand::rng();
    let country = COUNTRIES[rng.random_range(0..COUNTRIES.len())];
    tracing::info!(country = country.as_str(), "选择国家");

    // 记录请求数、执行耗时和调用分析指标。
    metrics::metrics_counter_demo(country.as_str());
    metrics::metrics_histogram_demo();
    metrics::metrics_rpc_demo("server");
    metrics::metrics_rpc_demo("client");

    // 演示自定义 Span、属性、事件和关联关系。
    traces::traces_custom_span_demo();
    traces::traces_set_custom_span_attributes();
    traces::traces_span_event_demo();
    traces::traces_span_links_demo();
    if rng.random::<f64>() < state.error_rate {
        let error = traces_random_error_demo(&mut rng);
        return Err((StatusCode::INTERNAL_SERVER_ERROR, error.as_str().to_owned()));
    }

    Ok(format!("Hello World, {}!", country.as_str()))
}

// 随机选择一类预置异常，由入口 Span 返回 500。
fn traces_random_error_demo(rng: &mut impl Rng) -> DemoError {
    DEMO_ERRORS[rng.random_range(0..DEMO_ERRORS.len())]
}

const DEMO_ERRORS: [DemoError; 4] = [
    DemoError::MySqlConnectTimeout,
    DemoError::UserNotFound,
    DemoError::NetworkUnreachable,
    DemoError::FileNotFound,
];
