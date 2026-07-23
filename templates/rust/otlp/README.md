# Rust（OpenTelemetry SDK）接入

{{OVERVIEW}}

## 1. 前置准备

### 1.1 术语介绍

{{TERM_INTRO}}

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：
* Git
* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/rust-examples/helloworld
docker build -t helloworld-rust:latest .
```

## 2. 快速接入

### 2.1 创建应用

{{APPLICATION_SET_UP}}

### 2.2 开箱即用 SDK 接入示例

OpenTelemetry 提供标准化的框架和工具包，用于创建和管理 Traces、Metrics、Logs 数据。

示例项目提供集成 OpenTelemetry Rust SDK 并将观测数据发送到 bk-collector 的方式，可以参考 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/rust-examples/helloworld/src/telemetry/setup.rs" target="_blank">src/telemetry/setup.rs</a> 进行接入。

### 2.3 关键配置

🌟 请仔细阅读本章节，以确保观测数据能准确上报到 APM。

#### 2.3.1 上报地址 & 应用 Token

{{MUST_CONFIG_EXPORTER}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/rust-examples/helloworld/src/telemetry/setup.rs" target="_blank">src/telemetry/setup.rs setup</a> 提供了创建样例：

```rust
use std::collections::HashMap;

use opentelemetry_otlp::{Protocol, WithExportConfig, WithHttpConfig};

// 三种信号使用相同的方式配置 OTLP HTTP/protobuf exporter。
let headers = HashMap::from([
    // ❗❗【非常重要】请传入应用 Token，不能在代码中写入真实 Token。
    ("x-bk-token".to_owned(), config.token.clone()),
]);

let exporter = opentelemetry_otlp::SpanExporter::builder()
    .with_http()
    .with_protocol(Protocol::HttpBinary)
    // ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写。
    // 示例程序会为无协议的 OTLP_ENDPOINT 补充 http://，并在此追加 /v1/traces。
    .with_endpoint(format!(
        "{}/v1/traces",
        config.otlp_endpoint.trim_end_matches('/')
    ))
    .with_headers(headers)
    .build()?;
```

指标、日志的配置方式和上述一致，请参考 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/rust-examples/helloworld/src/telemetry/setup.rs" target="_blank">src/telemetry/setup.rs</a> 中的 `MetricExporter`、`LogExporter` 初始化代码。

如果没有在 SDK builder 中显式调用 `with_headers`，`x-bk-token` 也可以通过「环境变量」的方式进行配置：

```shell
export OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo"
```

配置优先级：SDK > 环境变量，更多请参考 <a href="https://docs.rs/opentelemetry-otlp/0.32.0/opentelemetry_otlp/#environment-variables" target="_blank">Header Configuration</a>。

#### 2.3.2 服务信息

{{MUST_CONFIG_RESOURCES}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/rust-examples/helloworld/src/telemetry/setup.rs" target="_blank">src/telemetry/setup.rs setup</a> 提供了创建样例：

```rust
use opentelemetry_sdk::Resource;

// 三种信号共享同一 Resource，平台据此将数据归属到指定服务。
let resource = Resource::builder()
    // ❗❗【非常重要】应用服务唯一标识，必须与 APM 应用中的服务标识保持一致。
    .with_service_name(config.service_name.clone())
    .build();
```

## 3. 使用场景

示例项目整理常见的使用场景，集中在：

```rust
async fn hello_world(
    State(state): State<AppState>,
    headers: HeaderMap,
) -> Result<String, (StatusCode, String)> {
    let parent_context = global::get_text_map_propagator(|propagator| {
        propagator.extract(&HeaderExtractor(&headers))
    });
    let span = tracing::info_span!("Handle/HelloWorld");
    if let Err(error) = span.set_parent(parent_context) {
        tracing::warn!(%error, "设置服务端调用链父上下文失败");
    }
    let _entered = span.enter();

    // Logs（日志）
    logs::logs_demo();

    let mut rng = rand::rng();
    let country = COUNTRIES[rng.random_range(0..COUNTRIES.len())];
    tracing::info!(country = country.as_str(), "选择国家");

    // Metrics（指标）
    metrics::metrics_counter_demo(country.as_str());
    metrics::metrics_histogram_demo();

    // Traces（调用链）
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
```

可以结合代码和下方说明进行使用：<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/rust-examples/helloworld/src/http/server.rs" target="_blank">src/http/server.rs</a>。

### 3.1 Traces

#### 3.1.1 创建 Resource

Resource 代表观测数据所属的资源实体。

例如运行在 Kubernetes 上的容器所生成的观测数据，具有进程名称、Pod 名称等资源实体信息。

Rust SDK 可以通过 `Resource::builder()` 创建 Resource，并通过 `with_service_name` 设置服务标识：

```rust
use opentelemetry_sdk::Resource;

let resource = Resource::builder()
    // ❗❗【非常重要】应用服务唯一标识。
    .with_service_name(config.service_name.clone())
    .build();

let tracer_provider = SdkTracerProvider::builder()
    .with_batch_exporter(exporter)
    .with_resource(resource.clone())
    .build();
```

* <a href="https://docs.rs/opentelemetry_sdk/0.32.0/opentelemetry_sdk/struct.Resource.html" target="_blank">Resources</a>

#### 3.1.2 创建 Span

Span 是 Traces 的构建块，代表一个工作或操作单元。

示例通过 `tracing` 创建 Span，并由 `tracing-opentelemetry` 桥接到 OpenTelemetry。`tracing::info_span!` 用于创建和管理 Span。

**后续样例提及的 Span 创建方式如下：**

```rust
let span = tracing::info_span!("CustomSpanDemo/doSomething");
let _entered = span.enter();
tracing::info!("custom span work completed");
```

进入 Span 后，在当前作用域内产生的子 Span 和日志会自动继承上下文，示例代码如下：

```rust
/// 创建描述内部操作的子 Span，并写入业务属性。
pub fn traces_custom_span_demo() {
    let span = tracing::info_span!("CustomSpanDemo/doSomething");
    span.set_attribute("custom_key", "custom_value");
    let _entered = span.enter();
    tracing::info!("custom span work completed");
}
```

* <a href="https://docs.rs/tracing/0.1.41/tracing/macro.info_span.html" target="_blank">Creating Spans</a>

#### 3.1.3 设置属性

Attributes（属性）是 Span 元数据，以 Key-Value 形式存在。

在 Span 设置属性，对问题定位、过滤、聚合非常有帮助。

```rust
use tracing_opentelemetry::OpenTelemetrySpanExt;

// 增加 Span 自定义属性。
let span = tracing::info_span!("CustomSpanDemo/doSomething");
span.set_attribute("custom_key", "custom_value");
```

* <a href="https://docs.rs/tracing-opentelemetry/0.33.0/tracing_opentelemetry/trait.OpenTelemetrySpanExt.html" target="_blank">Span Attributes</a>

#### 3.1.4 设置事件

Event（事件）是一段人类可读信息，用于记录 Span 生命周期内发生的事情。

```rust
use tracing_opentelemetry::OpenTelemetrySpanExt;

/// Traces（调用链）- Span 事件。
pub fn traces_span_event_demo() {
    let span = tracing::info_span!("SpanEventDemo/doSomething");
    let _entered = span.enter();
    span.add_event("Before doSomething", vec![]);
    span.add_event("After doSomething", vec![]);
}
```

* <a href="https://docs.rs/tracing-opentelemetry/0.33.0/tracing_opentelemetry/trait.OpenTelemetrySpanExt.html#tymethod.add_event" target="_blank">Span Events</a>

#### 3.1.5 设置 Links

Links 用于在当前 Span 和其他 Span 之间建立关联，适合表达异步调用、批处理等不适合用父子关系承载的场景。

示例中 `SpanLinkDemo/asyncCaller` 表示异步操作，并通过 Link 与当前请求 Span 建立关联。

Link 只表达 Span 之间的关联，不会改变当前 Span 的父子关系。

```rust
use opentelemetry::{trace::TraceContextExt, KeyValue};
use tracing_opentelemetry::OpenTelemetrySpanExt;

/// 使用 Span Link 关联异步操作与当前请求。
pub fn traces_span_links_demo() {
    let async_span = tracing::info_span!("SpanLinkDemo/asyncCaller");
    let async_context = async_span.context();
    tracing::Span::current().add_link_with_attributes(
        async_context.span().span_context().clone(),
        vec![KeyValue::new("link_type", "async")],
    );
    tracing::info!("SpanLinkDemo async caller");
}
```

* <a href="https://docs.rs/opentelemetry/0.32.0/opentelemetry/trace/struct.SpanRef.html#method.add_link" target="_blank">Specifying links</a>

#### 3.1.6 记录错误

当一个 Span 出现错误，可以对其进行错误记录。

```rust
use opentelemetry::trace::TraceContextExt;
use tracing_opentelemetry::OpenTelemetrySpanExt;

let error = std::io::Error::other("request failed");
let context = tracing::Span::current().context();
context.span().record_error(&error);
```

* <a href="https://docs.rs/opentelemetry/0.32.0/opentelemetry/trace/struct.SpanRef.html#method.record_error" target="_blank">Record errors</a>

#### 3.1.7 设置状态

当一个 Span 未能成功，可以通过设置状态进行显式标记。

```rust
use opentelemetry::trace::{Status, TraceContextExt};
use tracing_opentelemetry::OpenTelemetrySpanExt;

let context = tracing::Span::current().context();
context
    .span()
    .set_status(Status::error("request failed"));
```
* <a href="https://docs.rs/opentelemetry/0.32.0/opentelemetry/trace/struct.SpanRef.html#method.set_status" target="_blank">Set span status</a>

#### 3.1.8 在当前 Span 上设置自定义属性

在部分场景下，Span 可能在框架入口、中间件等位置便被创建，如果你希望在当前的 Span 设置属性，而不是新创建一个 Span，可以通过以下方式进行：

```rust
use tracing_opentelemetry::OpenTelemetrySpanExt;

/// Traces（调用链）- 在当前 Span 上设置自定义属性。
pub fn traces_set_custom_span_attributes() {
    let span = tracing::Span::current();
    span.set_attribute("api_name", "ApiRequest");
    span.set_attribute("act_id", 12345_i64);
}
```

### 3.2 Metrics

#### 3.2.1 创建 Meter

<a href="https://docs.rs/opentelemetry/0.32.0/opentelemetry/metrics/struct.Meter.html" target="_blank">`Meter`</a> 是一个负责创建 Instruments 的对象。它提供了 API 接口，允许开发人员在代码中定义和记录 Metrics。

后续样例提及的 `meter` 创建方式如下：

```rust
use opentelemetry::global;

let meter = global::meter("helloworld");
```

#### 3.2.2 Counters

Counters（计数器）用于记录非负递增值。

例如，可以通过以下方式上报请求总数：

```rust
use opentelemetry::{global, KeyValue};

/// Metrics（指标）- 使用 Counter 类型指标。
pub fn metrics_counter_demo(country: &str) {
    global::meter("helloworld")
        .u64_counter("requests_total")
        .with_description("Total number of HTTP requests")
        .build()
        .add(1, &[KeyValue::new("country", country.to_owned())]);
}
```
* <a href="https://docs.rs/opentelemetry/0.32.0/opentelemetry/metrics/struct.Counter.html" target="_blank">Using Counters</a>

#### 3.2.3 Histograms

Histograms（直方图）用于记录数值分布情况。

例如，可以通过以下方式上报某段逻辑的处理耗时：

```rust
use opentelemetry::global;

pub fn metrics_histogram_demo() {
    let started_at = std::time::Instant::now();
    do_something();
    global::meter("helloworld")
        .f64_histogram("task_execute_duration_seconds")
        .with_description("Task execute duration in seconds")
        .build()
        .record(started_at.elapsed().as_secs_f64(), &[]);
}
```

* <a href="https://docs.rs/opentelemetry/0.32.0/opentelemetry/metrics/struct.Histogram.html" target="_blank">Using Histograms</a>

#### 3.2.4 Gauges

Gauges（仪表）用于记录瞬时值。

例如，可以通过以下方式，上报当前内存使用率：

```rust
use opentelemetry::global;
use rand::Rng;

/// Metrics（指标）- 使用 ObservableGauge 类型指标。
pub fn register_metrics_gauge_demo() {
    global::meter("helloworld")
        .f64_observable_gauge("memory_usage")
        .with_description("Memory usage")
        .with_callback(|observer| {
            observer.observe(0.1 + rand::rng().random_range(0.0..0.2), &[]);
        })
        .build();
}
```

* <a href="https://docs.rs/opentelemetry/0.32.0/opentelemetry/metrics/struct.ObservableGauge.html" target="_blank">Using Gauges</a>

### 3.3 Logs

#### 3.3.1 记录日志

```rust
/// Logs（日志）- 打印日志。
pub fn logs_demo() {
    // 上报日志。
    tracing::info!("收到请求：GET /helloworld");

    // 添加自定义属性。
    tracing::info!(
        method = "GET",
        k1 = "v1",
        k2 = 123,
        "上报带自定义属性的请求日志"
    );
}
```

## 4. 快速体验

### 4.1 运行样例

#### 4.1.1 运行

❗❗【非常重要】Rust SDK 的场景 `OTLP_ENDPOINT` 无需 `http://` 前缀，示例程序会统一补充，否则启动会失败。

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="{{access_config.token}}" \
-e SERVICE_NAME="{{service_name}}" \
-e OTLP_ENDPOINT="{{access_config.otlp.http_endpoint_without_schema}}" \
-e ENABLE_TRACES="{{access_config.otlp.enable_traces}}" \
-e ENABLE_METRICS="{{access_config.otlp.enable_metrics}}" \
-e ENABLE_LOGS="{{access_config.otlp.enable_logs}}" helloworld-rust:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

#### 4.1.2 运行参数说明

| 参数                   | 值（根据所填写接入信息生成）                                          | 说明                                                                                                                                                       |
|----------------------|:--------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| `TOKEN`              | `"{{access_config.token}}"`                             | 【必须】APM 应用 `Token`。                                                                                                                                      |
| `SERVICE_NAME`       | `"{{service_name}}"`                                    | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分。                                                                                                                          |
| `OTLP_ENDPOINT`      | `"{{access_config.otlp.http_endpoint_without_schema}}"` | 【必须】OT 数据上报地址。Rust demo 使用 `HTTP/protobuf` 协议，并为该地址追加对应信号路径。 |
| `PROFILING_ENDPOINT` | `"{{access_config.profiling.endpoint}}"`                | 当前 Rust demo 不读取该参数。                                                                                                                                    |
| `ENABLE_TRACES`      | `{{access_config.otlp.enable_traces}}`                  | 是否启用调用链上报。                                                                                                                                               |
| `ENABLE_METRICS`     | `{{access_config.otlp.enable_metrics}}`                 | 是否启用指标上报。                                                                                                                                                |
| `ENABLE_LOGS`        | `{{access_config.otlp.enable_logs}}`                    | 是否启用日志上报。                                                                                                                                                |
| `ENABLE_PROFILING`   | `{{access_config.profiling.enabled}}`                   | 当前 Rust demo 不读取该参数。                                                                                                                                              |

* *<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/" target="_blank">OTLP Exporter Configuration</a>*

### 4.2 查看数据

#### 4.2.1 Traces 检索

{{VIEW_TRACES_DATA}}

#### 4.2.2 指标检索

{{VIEW_CUSTOM_METRICS_DATA}}

#### 4.2.3 日志检索

{{VIEW_LOG_DATA}}

## 5. 了解更多

{{LEARN_MORE}}
