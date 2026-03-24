# C++（OpenTelemetry SDK）接入

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
git clone {{ECOSYSTEM_REPOSITORY_URL}}.git
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/cpp-examples/helloworld
docker build -t helloworld-cpp:latest .
```

## 2. 快速接入

### 2.1 创建应用

{{APPLICATION_SET_UP}}

### 2.2 开箱即用 SDK 接入示例

OpenTelemetry 提供标准化的框架和工具包，用于创建和管理 Traces、Metrics、Logs 数据。

#### 2.2.1 Traces、Metrics、Logs 接入示例

示例项目提供集成 OpenTelemetry Cpp SDK 并将观测数据发送到 bk-collector 的方式，可以参考下面的代码：
* Traces：<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/cpp-examples/helloworld/include/otlp/tracer_common.h" target="_blank">include/otlp/tracer_common.h</a>
* Metrics：<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/cpp-examples/helloworld/include/otlp/meter_common.h" target="_blank">include/otlp/meter_common.h</a>
* Logs：<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/cpp-examples/helloworld/include/otlp/logger_common.h" target="_blank">include/otlp/logger_common.h</a>

在 `main` 文件中启动注册：

```cpp
#include "otlp/resource_common.h"
#include "otlp/tracer_common.h"
#include "otlp/meter_common.h"
#include "otlp/logger_common.h"

int main() {
    const Config &config = Config::getInstance();
    auto resource = CreateResource(config);

    initTracer(config, resource);
    initMeter(config, resource);
    initLogger(config, resource);

    // .. 业务启动代码

    cleanupTracer(config);
    cleanupMeter(config);
    cleanupLogger(config);

    return 0;
}
```

#### 2.2.2 构建

引入 OpenTelemetry C++ SDK 需要重新编译项目，示例项目提供 Dockerfile 以供参考：<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/cpp-examples/helloworld/Dockerfile" target="_blank">Dockerfile</a>。

### 2.3 关键配置

🌟 请仔细阅读本章节，以确保观测数据能准确上报到 APM。

#### 2.3.1 上报地址 & 应用 Token

{{MUST_CONFIG_EXPORTER}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/cpp-examples/helloworld/include/otlp/tracer_common.h" target="_blank">include/otlp/tracer_common.h</a> 提供了创建样例：

```cpp
#include "opentelemetry/exporters/otlp/otlp_http.h"
#include "opentelemetry/exporters/otlp/otlp_http_exporter_factory.h"
#include "opentelemetry/exporters/otlp/otlp_http_exporter_options.h"

void initTracer(const Config &config, const resource_sdk::Resource &resource) {
    otel_exporter::OtlpHttpExporterOptions otlpOptions;
    //❗️❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
    otlpOptions.url = config.OtlpEndpoint + "/v1/traces";
    //❗️❗【非常重要】请传入应用 Token
    otlpOptions.http_headers.insert({"x-bk-token", config.Token});
    auto exporter = otel_exporter::OtlpHttpExporterFactory::Create(otlpOptions);
    ...
```
* Logs / Metrics Exporter 创建代码也类似，可以参考同目录下的 `meter_common.h`、`logger_common.h`。
* 使用 `HTTP` 上报需要在页面提供的接入地址基础上，指定<a href="https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#endpoint-urls-for-otlphttp" target="_blank">上报接口路径</a>：
  * Traces - `/v1/traces`
  * Metrics - `/v1/metrics`
  * Logs - `/v1/logs`

`x-bk-token` 也可以通过「环境变量」的方式进行配置：

```shell
# 多个 kv 以英文逗号分隔，例如：k1=v1,k2=v2。
export OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo"
```

配置优先级：SDK > 环境变量，更多请参考 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#header-configuration" target="_blank">Header Configuration</a>。

#### 2.3.2 服务信息

{{MUST_CONFIG_RESOURCES}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/cpp-examples/helloworld/include/otlp/resource_common.h" target="_blank">include/otlp/meter_common.h</a> 提供了创建样例：

```cpp
resource_sdk::Resource CreateResource(const Config &config) {
    // 使用 SDK 默认属性
    auto defaultResource = resource_sdk::Resource::GetDefault();
    auto resourceAttributes = resource_sdk::ResourceAttributes{
            //❗️❗【非常重要】应用服务唯一标识
            {resource_sdk::SemanticConventions::kServiceName, config.ServiceName},
            ...
    };
    return defaultResource.Merge(resource_sdk::Resource::Create(resourceAttributes));
}
```

## 3. 使用场景

### 3.1 Traces

示例项目整理常见的使用场景，集中在：

```cpp
std::shared_ptr<HttpRequestHandler::OutgoingResponse>
Handler::handleHelloWorld(const std::shared_ptr<HttpRequestHandler::IncomingRequest> &request) {
    const Config &config = Config::getInstance();
    auto logger = getLogger(config.ServiceName);

    auto span = get_tracer(config.ServiceName)->StartSpan("Handle/HelloWorld");
    auto scope = get_tracer(config.ServiceName)->WithActiveSpan(span);

    // Logs（日志）
    helloWorldHelper.logsDemo(request);

    auto country = helloWorldHelper.choiceCountry();
    logger->Info("get country -> " + country);

    // Metrics（指标） - Counter 类型
    helloWorldHelper.metricsCounterDemo(country);
    // Metrics（指标） - Histograms 类型
    helloWorldHelper.metricsHistogramDemo();

    // Traces（调用链）- 自定义 Span
    HelloWorldHelper::tracesCustomSpanDemo();
    // Traces（调用链）- Span 事件
    HelloWorldHelper::tracesSpanEventDemo();

    // Traces（调用链）- 模拟错误
    if (auto err = helloWorldHelper.tracesRandomErrorDemo()) {
        auto response = ResponseFactory::createResponse(Status::CODE_500, err->what());

        span->End();
        return response;
    }

    auto greeting = HelloWorldHelper::generateGreeting(country);
    auto response = ResponseFactory::createResponse(Status::CODE_200, greeting.c_str());

    span->End();
    return response;
}
```

可以结合代码和下方说明进行使用：<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/cpp-examples/helloworld/src/server.cpp" target="_blank">src/server.cpp</a>。

#### 3.1.1 创建 Resource

Resource 代表观测数据所属的资源实体。

例如运行在 Kubernetes 上的容器所生成的观测数据，具有进程名称、Pod 名称等资源实体信息。

我们提供一个样例，用于自动检测正在运行的进程、所在操作系统等资源属性信息：

```cpp
namespace {
    resource_sdk::Resource CreateResource(const Config &config) {
        auto defaultResource = resource_sdk::Resource::GetDefault();
        auto resourceAttributes = resource_sdk::ResourceAttributes{
                //❗️❗【非常重要】应用服务唯一标识
                {resource_sdk::SemanticConventions::kServiceName, config.ServiceName},
                {resource_sdk::SemanticConventions::kProcessPid,  GetProcessId()},
                {resource_sdk::SemanticConventions::kOsType,      GetOperatingSystem()},
                {resource_sdk::SemanticConventions::kHostName,    GetHostName()},
        };
        return defaultResource.Merge(resource_sdk::Resource::Create(resourceAttributes));
    }
}
```

* 参考代码：<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/cpp-examples/helloworld/include/otlp/resource_common.h" target="_blank">include/otlp/resource_common.h</a>

#### 3.1.2 创建 Span

Span 是 Traces 的构建块，代表一个工作或操作单元。

通过 `get_tracer` 获取 `Tracer` 对象，调用 `StartSpan` 进行创建 Span：

```cpp
const Config &config = Config::getInstance();
auto span = get_tracer(config.ServiceName)
        ->StartSpan("CustomSpanDemo/doSomething");
auto scope = get_tracer(config.ServiceName)->WithActiveSpan(span);
doSomething(50);
span->End();
```

* <a href="https://opentelemetry.io/docs/languages/cpp/instrumentation/#start-a-span" target="_blank">Creating Spans</a>

#### 3.1.3 设置属性

Attributes（属性）是 Span 元数据，以 Key-Value 形式存在。

在 Span 设置属性，对问题定位、过滤、聚合非常有帮助。

```cpp
// 增加 Span 自定义属性
span->SetAttribute("helloworld.kind", 1);
span->SetAttribute("helloworld.step", "tracesCustomSpanDemo");
```

#### 3.1.4 设置事件

Event（事件）是一段人类可读信息，用于记录 Span 生命周期内发生的事情。

```cpp
// tracesSpanEventDemo Traces（调用链）- Span 事件
void HelloWorldHelper::tracesSpanEventDemo() {
    const Config &config = Config::getInstance();
    auto span = get_tracer(config.ServiceName)->StartSpan("SpanEventDemo/doSomething");
    auto scope = get_tracer(config.ServiceName)->WithActiveSpan(span);

    span->AddEvent("Before doSomething");
    doSomething(50);
    span->AddEvent("After doSomething");
    span->End();
}
```
* <a href="https://opentelemetry-cpp.readthedocs.io/en/latest/otel_docs/classopentelemetry_1_1trace_1_1Span.html#exhale-class-classopentelemetry-1-1trace-1-1span" target="_blank">Span Events</a>

#### 3.1.5 记录错误

当一个 Span 出现错误，可以对其进行错误记录。

```cpp
std::shared_ptr<std::runtime_error> HelloWorldHelper::tracesRandomErrorDemo() {
    if (auto err = randErr(0.1)) {
        auto ctx = opentelemetry::context::RuntimeContext::GetCurrent();
        auto span = trace_api::GetSpan(ctx);

        auto exceptionMessage = err->what();
        auto exceptionType = typeid(err).name();
        span->AddEvent("exception", {
            {trace_api::SemanticConventions::kExceptionMessage, exceptionMessage},
            {trace_api::SemanticConventions::kExceptionType, exceptionType}
        });
        return err;
    }
    return nullptr;
}
```
* <a href="https://opentelemetry-cpp.readthedocs.io/en/latest/otel_docs/classopentelemetry_1_1trace_1_1Span.html#exhale-class-classopentelemetry-1-1trace-1-1span" target="_blank">Record errors</a>

#### 3.1.6 设置状态

当一个 Span 未能成功，可以通过设置状态进行显式标记。

```cpp
auto exceptionMessage = err->what();
span->SetStatus(trace_api::StatusCode::kError, exceptionMessage);
```
* <a href="https://opentelemetry-cpp.readthedocs.io/en/latest/otel_docs/classopentelemetry_1_1trace_1_1Span.html#exhale-class-classopentelemetry-1-1trace-1-1span" target="_blank">Set span status</a>

### 3.2 Metrics

#### 3.2.1 Counters

Counters（计数器）用于记录非负递增值。

例如，可以通过以下方式上报请求总数：

```cpp
//【建议】初始化指标再使用，而不是在业务逻辑里初始化
auto meter = get_meter(config.ServiceName);
requestsTotal = meter->CreateUInt64Counter("requests_total", "Total number of HTTP requests");

// metricsCounterDemo Metrics（指标）- 使用 Counter 类型指标
void HelloWorldHelper::metricsCounterDemo(const std::string &country) {
    requestsTotal->Add(1, {{"country", country}});
}
```
* <a href="https://opentelemetry.io/docs/languages/cpp/instrumentation/#create-a-counter" target="_blank">Create a counter</a>

#### 3.2.2 Histograms

Histograms（直方图）用于记录数值分布情况。

例如，可以通过以下方式上报某段逻辑的处理耗时：

```cpp
taskExecuteDurationSeconds = meter->CreateDoubleHistogram("task_execute_duration_seconds", "Task execute duration in seconds");

// metricsHistogramDemo Metrics（指标）- 使用 Histogram 类型指标
void HelloWorldHelper::metricsHistogramDemo() {
    auto begin = std::chrono::high_resolution_clock::now();
    doSomething(100);
    auto end = std::chrono::high_resolution_clock::now();

    std::chrono::duration<double> duration = end - begin;
    taskExecuteDurationSeconds->Record(duration.count(), {});
}
```
* <a href="https://opentelemetry.io/docs/languages/cpp/instrumentation/#create-a-histogram" target="_blank">Create a histogram</a>

### 3.3 Logs

#### 3.3.1 记录日志

```cpp
// logsDemo Logs（日志）打印日志
void HelloWorldHelper::logsDemo(const std::shared_ptr<oatpp::web::server::HttpRequestHandler::IncomingRequest> &request) {
    std::string url = request->getPathTail().getValue("");
    std::string method = request->getStartingLine().method.toString();
    logger->Info(std::string(__func__ ) + "received request: " + method + " " + url);
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
-e ENABLE_LOGS="{{access_config.otlp.enable_logs}}" helloworld-cpp:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

#### 4.1.2 运行参数说明

{{DEMO_RUN_PARAMETERS}}

### 4.2 查看数据

#### 4.2.1 Traces 检索

{{VIEW_TRACES_DATA}}

#### 4.2.2 指标检索

{{VIEW_CUSTOM_METRICS_DATA}}

#### 4.2.3 日志检索

{{VIEW_LOG_DATA}}

## 5. 了解更多

{{LEARN_MORE}}