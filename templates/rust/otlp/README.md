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

### 2.4 OpenTelemetry 组件埋点工具

为 Web 框架、HTTP & 数据库 & 消息队列客户端等应用依赖接入 OpenTelemetry 时，可以先在 <a href="https://github.com/open-telemetry/opentelemetry-rust-contrib" target="_blank">OpenTelemetry Rust Contrib</a> 优先查找适用的插桩库（Instrumentation Library）。

已有成熟插桩库时，优先使用其自动插桩能力；对于插桩库无法覆盖的业务操作，参考下方文档按需手动创建 Span。

#### 2.4.1 选择埋点组件

<a href="https://github.com/open-telemetry/opentelemetry-rust-contrib" target="_blank">OpenTelemetry Rust Contrib</a> 中常用的插桩库有 👇：

| 库或框架 | 埋点组件 |
| --- | --- |
| Actix Web | <a href="https://github.com/open-telemetry/opentelemetry-rust-contrib/tree/main/opentelemetry-instrumentation-actix-web" target="_blank">opentelemetry-instrumentation-actix-web</a> |
| Tower、Axum、Hyper | <a href="https://github.com/open-telemetry/opentelemetry-rust-contrib/tree/main/opentelemetry-instrumentation-tower" target="_blank">opentelemetry-instrumentation-tower</a> |
| Tonic（gRPC） | <a href="https://github.com/open-telemetry/opentelemetry-rust-contrib/tree/main/opentelemetry-instrumentation-tower" target="_blank">opentelemetry-instrumentation-tower</a> |
| SQLx | <a href="https://docs.rs/sqlx-tracing/latest/sqlx_tracing/" target="_blank">sqlx-tracing</a> |

如果 Contrib 中没有对应组件，再检查目标库的官方文档和社区中间件。只有在缺少成熟方案时，才自行实现协议层 Span、语义属性和上下文传播。

#### 2.4.2 Actix Web 服务端示例

Actix Web 可以使用 OpenTelemetry Rust Contrib 中的 `opentelemetry-instrumentation-actix-web`。在 `Cargo.toml` 中引入依赖：

```toml
[dependencies]
actix-web = "4.12"
opentelemetry-instrumentation-actix-web = "0.24"
```

通过 middleware 自动创建服务端 HTTP Span 和指标：

```rust
use actix_web::{web, App, HttpServer};
use opentelemetry_instrumentation_actix_web::{
    RequestMetrics,
    RequestTracing,
};

HttpServer::new(|| {
    App::new()
        .wrap(RequestTracing::new())
        .wrap(RequestMetrics::default())
        .route("/helloworld", web::get().to(hello_world))
})
.bind(("0.0.0.0", 8080))?
.run()
.await?;
```

* <a href="https://github.com/open-telemetry/opentelemetry-rust-contrib/tree/main/opentelemetry-instrumentation-actix-web" target="_blank">Actix Web instrumentation</a>

## 3. 使用场景

示例项目整理常见的使用场景，集中在：

```rust
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
use opentelemetry::{trace::TraceContextExt, Context, KeyValue};
use rand::Rng;
use tracing::Instrument;
use tracing_opentelemetry::OpenTelemetrySpanExt;

/// 使用 Span Link 关联异步操作与当前请求。
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

fn metrics_histogram_demo() {
    let started_at = std::time::Instant::now();
    do_something(100);
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
use std::sync::OnceLock;

use opentelemetry::{global, metrics::ObservableGauge};
use rand::Rng;

static MEMORY_USAGE: OnceLock<ObservableGauge<f64>> = OnceLock::new();

/// Metrics（指标）- 使用 ObservableGauge 类型指标。
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

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="{{access_config.token}}" \
-e SERVICE_NAME="{{service_name}}" \
-e OTLP_ENDPOINT="{{access_config.otlp.http_endpoint}}" \
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
| `OTLP_ENDPOINT`      | `"{{access_config.otlp.http_endpoint}}"`                | 【必须】OT 数据上报地址，请根据页面指引提供的接入地址进行填写。Rust demo 使用 `HTTP/protobuf` 协议，并为该地址追加对应信号路径。 |
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
