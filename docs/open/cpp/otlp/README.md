# C++（OpenTelemetry SDK）接入

本指南将帮助您使用 OpenTelemetry SDK 接入蓝鲸应用性能监控，以 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/examples/helloworld.md" target="_blank">入门项目-HelloWorld</a> 为例，介绍调用链、指标、日志数据接入及 SDK 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速入门 OpenTelemetry，接入并体验蓝鲸应用性能监控相关功能。

## 1. 前置准备

### 1.1 术语介绍

* Traces：<a href="https://opentelemetry.io/docs/concepts/signals/traces/" target="_blank">调用链</a>，表示请求在应用程序的执行路径。
* Metrics：<a href="https://opentelemetry.io/docs/concepts/signals/metrics/" target="_blank">指标</a>，表示对运行服务的测量。
* Logs: <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">日志</a>，表示对事件的记录。
* Profiling: <a href="https://grafana.com/docs/pyroscope/latest/introduction/profiling/" target="_blank">性能分析</a>，表示对应用程序运行时资源的持续测量。
* Telemetry Data：观测数据，指代 Traces、Metrics、Logs、Profiling 等。
* APM：蓝鲸观测平台应用性能监控，提供四类观测数据开箱即用的观测能力。
* <a href="https://github.com/TencentBlueKing/bkmonitor-datalink/tree/main/pkg/collector" target="_blank">bk-collector</a>：腾讯蓝鲸的 APM 服务端组件，负责接收 Prometheus、OpenTelemetry、Jaeger、Skywalking 等主流开源组件的观测数据，并对数据进行清洗转发到观测平台链路。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：
* Git
* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem.git
cd bkmonitor-ecosystem/examples/cpp-examples/helloworld
docker build -t helloworld-cpp:latest .
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

### 2.2 开箱即用 SDK 接入示例

OpenTelemetry 提供标准化的框架和工具包，用于创建和管理 Traces、Metrics、Logs 数据。

#### 2.2.1 Traces、Metrics、Logs 接入示例

示例项目提供集成 OpenTelemetry Cpp SDK 并将观测数据发送到 bk-collector 的方式，可以参考下面的代码：
* Traces：<a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/cpp-examples/helloworld/include/otlp/tracer_common.h" target="_blank">include/otlp/tracer_common.h</a>
* Metrics：<a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/cpp-examples/helloworld/include/otlp/meter_common.h" target="_blank">include/otlp/meter_common.h</a>
* Logs：<a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/cpp-examples/helloworld/include/otlp/logger_common.h" target="_blank">include/otlp/logger_common.h</a>

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

引入 OpenTelemetry C++ SDK 需要重新编译项目，示例项目提供 Dockerfile 以供参考：<a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/cpp-examples/helloworld/Dockerfile" target="_blank">Dockerfile</a>。

### 2.3 关键配置

🌟 请仔细阅读本章节，以确保观测数据能准确上报到 APM。

#### 2.3.1 上报地址 & 应用 Token

请在创建 <a href="https://opentelemetry.io/docs/specs/otel/protocol/exporter/" target="_blank">Exporter</a> 时准确传入以下信息：

| 参数         | 说明                            |
|------------|-------------------------------|
| `endpoint` | 【必须】数据上报地址，请根据页面指引提供的接入地址进行填写。 |
| `x-bk-token`| 【必须】APM 应用 Token，作为 headers 传递。 |

示例项目在 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/cpp-examples/helloworld/include/otlp/tracer_common.h" target="_blank">include/otlp/tracer_common.h</a> 提供了创建样例：

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

请在 <a href="https://opentelemetry.io/docs/concepts/resources/" target="_blank">Resources</a> 添加以下属性，蓝鲸观测平台通过这些属性，将数据关联到具体的应用、资源实体：

| 属性                       | 说明                                          |
|--------------------------|---------------------------------------------|
| `service.name`           | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分              |
| `net.host.ip`            | 【可选】关联 CMDB 主机                              |
| `telemetry.sdk.language` | 【可选】标识应用对应的开发语言，SDK Default Resource 会提供该属性 |
| `telemetry.sdk.name`     | 【可选】OT SDK 名称，SDK Default Resource 会提供该属性   |
| `telemetry.sdk.version`  | 【可选】OT SDK 版本，SDK Default Resource 会提供该属性   |
| `k8s.bcs.cluster.id`     | 【可选】集群 ID，支持自动关联。                                        |
| `k8s.pod.name`           | 【可选】Pod 名称                                       |
| `k8s.namespace.name`     | 【可选】Pod 所在命名空间                                |

**a. 如何自动发现容器信息**

蓝鲸 APM 支持与 BCS 打通，你可以通过以下方式简单配置，将服务与容器信息进行关联，实现在 APM 查看服务所关联容器负载的监控、事件数据。

方案 1：🌟 通过集群内上报【推荐】

将上报域名切换为集群内域名，端口、上报路径与之前一致，即可自动获取关联。

方案 2：手动关联

手动补充上述的 `k8s.bcs.cluster.id`、`k8s.pod.name`、`k8s.namespace.name` 字段，也可以进行关联。

除了 `k8s.bcs.cluster.id` 外，可以在相应的 k8s 描述文件（Yaml）中，将 Pod 字段作为环境变量的值，然后在程序端读取，设置到 Resources：

```yaml
template:
  spec:
    containers:
      - name: xxx
        image: xxx
        env:
          - name: "K8S_POD_IP"
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
          - name: "K8S_POD_NAME"
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: "K8S_NAMESPACE"
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
```

示例项目在 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/cpp-examples/helloworld/include/otlp/resource_common.h" target="_blank">include/otlp/meter_common.h</a> 提供了创建样例：

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

可以结合代码和下方说明进行使用：<a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/cpp-examples/helloworld/src/server.cpp" target="_blank">src/server.cpp</a>。

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

* 参考代码：<a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/cpp-examples/helloworld/include/otlp/resource_common.h" target="_blank">include/otlp/resource_common.h</a>

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
    requestsTotal->Add(1, ('country', Undefined));
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
docker run -e TOKEN="xxx" \
-e SERVICE_NAME="helloworld" \
-e OTLP_ENDPOINT="http://127.0.0.1:4318" \
-e ENABLE_TRACES="true" \
-e ENABLE_METRICS="true" \
-e ENABLE_LOGS="true" helloworld-cpp:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

#### 4.1.2 运行参数说明

运行参数说明：

| 参数                   | 推荐值                                | 说明                                        |
|----------------------|--------------------------------------|-------------------------------------------|
| `TOKEN`              | `""`                                 | APM 应用 `Token`。                            |
| `SERVICE_NAME`       | `"helloworld"`                       | 服务唯一标识，一个应用可以有多个服务，通过该属性区分。                |
| `OTLP_ENDPOINT`      | `"http://127.0.0.1:4318"` | OT 数据上报地址，请根据页面指引提供的接入地址进行填写，支持以下协议：<br />`gRPC`：`http://127.0.0.1:4317`<br />`HTTP`：`http://127.0.0.1:4318`（demo 使用该协议演示上报）<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。        |
| `PROFILING_ENDPOINT` | `"http://127.0.0.1:4318/pyroscope"`  | Profiling 数据上报地址，请根据页面指引提供的 HTTP 接入地址进行填写。<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。 |
| `ENABLE_TRACES`      | `false`                              | 是否启用调用链上报。                                 |
| `ENABLE_METRICS`     | `false`                              | 是否启用指标上报。                                  |
| `ENABLE_LOGS`        | `false`                              | 是否启用日志上报。                                  |
| `ENABLE_PROFILING`   | `false`                            | 是否启用性能分析上报。                                |

### 4.2 查看数据

#### 4.2.1 Traces 检索

Tracing 检索功能主要用于对分布式系统中的请求链路进行跟踪和分析，请参考<a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/explore_traces.md" target="_blank">「应用性能监控 APM/调用链追踪」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/traces.png)

#### 4.2.2 指标检索

自定义指标功能旨在帮助用户针对特定应用及其服务进行深度性能指标监控，请参考<a href="#" target="_blank">「应用性能监控 APM/自定义指标」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/custom-metrics.png)

#### 4.2.3 日志检索

日志功能主要用于查看和分析对应服务（应用程序）运行过程中产生的各类日志信息，请参考<a href="#" target="_blank">「应用性能监控 APM/日志分析」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/logs.png)

## 5. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>