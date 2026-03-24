# Python（OpenTelemetry SDK）接入

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
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/python-examples/helloworld
docker build -t helloworld-python:latest .
```

## 2. 快速接入

### 2.1 创建应用

{{APPLICATION_SET_UP}}

### 2.2 开箱即用 SDK 接入示例

OpenTelemetry 提供标准化的框架和工具包，用于创建和管理 Traces、Metrics、Logs 数据。

示例项目提供集成 OpenTelemetry Python SDK 并将观测数据发送到 bk-collector 的方式，可以参考 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/python-examples/helloworld/src/services/otlp.py" target="_blank">services/otlp.py</a> 进行接入。

### 2.3 关键配置

🌟 请仔细阅读本章节，以确保观测数据能准确上报到 APM。

#### 2.3.1 上报地址 & 应用 Token

{{MUST_CONFIG_EXPORTER}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/python-examples/helloworld/src/services/otlp.py" target="_blank">services/otlp.py _setup_traces</a> 提供了创建样例：

```python
def _setup_traces(self, resource: Resource):
    otlp_exporter = self._setup_trace_exporter()
    span_processor = BatchSpanProcessor(otlp_exporter)
    self.tracer_provider = TracerProvider(resource=resource)
    self.tracer_provider.add_span_processor(span_processor)
    trace.set_tracer_provider(self.tracer_provider)

def _setup_trace_exporter(self):
    return HTTPSpanExporter(
        # ❗️❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
        endpoint=f"{self.config.endpoint}/v1/traces",
        # ❗️❗【非常重要】配置为应用 Token
        headers={"x-bk-token": self.config.token}
    )
```

`x-bk-token` 也可以通过「环境变量」的方式进行配置：

```shell
export OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo"
```

配置优先级：SDK > 环境变量，更多请参考 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#header-configuration" target="_blank">Header Configuration</a>。

#### 2.3.2 服务信息

{{MUST_CONFIG_RESOURCES}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/python-examples/helloworld/src/services/otlp.py" target="_blank">services/otlp.py _create_resource</a> 提供了创建样例：

```python
from opentelemetry.sdk.resources import Resource, ResourceDetector, get_aggregated_resources, OsResourceDetector

def _create_resource(self) -> Resource:
    # ...
    # Detect os resources based on `Operating System conventions <https://opentelemetry.io/docs/specs/semconv/resource/os/>`_.
    detectors = [OsResourceDetector()]

    # create 提供了部分 SDK 默认属性
    initial_resource = Resource.create(
        {
            #❗❗【非常重要】应用服务唯一标识
            ResourceAttributes.SERVICE_NAME: self.config.service_name,
            # ...
        }
    )

    return get_aggregated_resources(detectors, initial_resource)
```

## 3. 使用场景

示例项目整理常见的使用场景，集中在：

```python
class HelloWorldHandler:
    ...
    def handle(self) -> str:
        # 不自动设置异常状态和记录异常，以展示手动设置方法 (traces_random_error_demo)
        with self.tracer.start_as_current_span(
                "handle/hello_world", record_exception=False, set_status_on_exception=False
        ):
            country = self.choice_country()
            otel_logger.info("get country -> %s", country)

            # Logs（日志）
            self.logs_demo(request)

            # Metrics（指标） - Counter 类型
            self.metrics_counter_demo(country)
            # Metrics（指标） - Histograms 类型
            self.metrics_histogram_demo()

            # Traces（调用链）- 自定义 Span
            self.traces_custom_span_demo()
            # Traces（调用链）- Span 事件
            self.traces_span_event_demo()
            # Traces（调用链）- 模拟错误
            self.traces_random_error_demo()

            return self.generate_greeting(country)
```

可以结合代码和下方说明进行使用：<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/python-examples/helloworld/src/services/server.py" target="_blank">services/server.py</a>。

### 3.1 Traces

#### 3.1.1 创建 Resource

Resource 代表观测数据所属的资源实体。

例如运行在 Kubernetes 上的容器所生成的观测数据，具有进程名称、Pod 名称等资源实体信息。

可以通过 <a href="https://opentelemetry-python.readthedocs.io/en/latest/sdk/resources.html#opentelemetry.sdk.resources.ResourceDetector" target="_blank">opentelemetry-resource-detectors</a> 自动检测正在运行的进程、所在操作系统等资源属性信息。

初始化 `Resources`

```Python
from opentelemetry.sdk.resources import (
    OsResourceDetector,
    ProcessResourceDetector,
    Resource,
    get_aggregated_resources,
)

def _create_resource(self) -> Resource:
    detectors = [ProcessResourceDetector(), OsResourceDetector()]

    # create 提供了部分 SDK 默认属性
    initial_resource = Resource.create(
        {
            #❗❗【非常重要】应用服务唯一标识
            ResourceAttributes.SERVICE_NAME: self.config.service_name,
            ResourceAttributes.OS_TYPE: platform.system().lower(),
            ResourceAttributes.HOST_NAME: socket.gethostname(),
        }
    )

    return get_aggregated_resources(detectors, initial_resource)
```

#### 3.1.2 获取 Tracer

Traces 是什么？

- 是请求通过您的应用程序的路径

- 一个完整的 Traces 通常由多个 Span 组成。每个 Span 代表一个在系统中发生的操作，例如一个服务处理请求或对数据库的查询。

- 可通过了解 <a href="https://opentelemetry.io/docs/concepts/signals/traces/" target="_blank">Traces 的概念</a> 进一步学习。

Tracer 是什么？

- Tracer 是一个用于创建和管理 Span 的对象。它提供了 API 接口，开发人员可以用它来在应用程序代码中生成和记录 Span。

- 可通过了解 <a href="https://opentelemetry.io/docs/specs/otel/trace/api/" target="_blank">Tracing API 的概念</a> 进一步学习。

获取 tracer：
```python
from opentelemetry import trace
tracer = trace.get_tracer(__name__)
```

#### 3.1.3 创建 Span

Span 是 Traces 的构建块，代表一个工作或操作单元。

Span 通过 `tracer.start_as_current_span()` 进行创建：

```python
def traces_custom_span_demo(self):
    with self.tracer.start_as_current_span("custom_span_demo/do_something"):
        self.do_something(50)
```

* <a href="https://opentelemetry.io/docs/languages/python/instrumentation/#creating-spans" target="_blank">Creating spans</a>

#### 3.1.4 设置属性

Attributes（属性）是 Span 元数据，以 Key-Value 形式存在。

在 Span 设置属性，对问题定位、过滤、聚合非常有帮助。

```python
# 增加 Span 自定义属性
span.set_attribute("helloworld.kind", 1)
span.set_attribute("helloworld.step", "traces_custom_span_demo")
```

* <a href="https://opentelemetry.io/docs/languages/python/instrumentation/#add-attributes-to-a-span" target="_blank">Add attributes to a span</a>

#### 3.1.5 设置事件

Event（事件）是一段人类可读信息，用于记录 Span 生命周期内发生的事情。

```python
def traces_span_event_demo(self):
    with self.tracer.start_as_current_span("span_event_demo/do_something") as span:
        attributes = {
            "helloworld.kind": 2,
            "helloworld.step": "traces_span_event_demo",
        }
        span.add_event("Before do_something", attributes)
        self.do_something(50)
        span.add_event("After do_something", attributes)
```

* <a href="https://opentelemetry.io/docs/languages/python/instrumentation/#adding-events" target="_blank">Adding events</a>

#### 3.1.6 记录异常

当一个 Span 出现异常，可以对其进行异常记录。

```python
def traces_random_error_demo(self):
    try:
        if random.random() < self.ERROR_RATE:
            error_message = random.choice(self.CUSTOM_ERROR_MESSAGES)
            raise APIException(error_message)
    except APIException as e:
        current_span: Span = trace.get_current_span()
        current_span.set_status(Status(StatusCode.ERROR, str(e)))
        current_span.record_exception(e)
        raise
```

* <a href="https://opentelemetry.io/docs/languages/python/instrumentation/#record-exceptions-in-spans" target="_blank">Record exceptions in spans</a>

#### 3.1.7 设置状态

当一个 Span 未能成功，可以通过设置状态进行显式标记。

```python
span.set_status(Status(StatusCode.ERROR, str(e)))
```

* <a href="https://opentelemetry.io/docs/languages/python/instrumentation/#set-span-status" target="_blank">Set span status</a>

### 3.2 Metrics

#### 3.2.1 获取 Meter

Metrics 是什么？

- Metrics 是在运行时捕获的服务测量指标。

- Metric Instruments 可以捕获测试结果。

- Instruments 有多种类型，如 Counter 和 Histogram 等。

- 可通过了解 <a href="https://opentelemetry.io/docs/concepts/signals/metrics/" target="_blank">Metrics 的概念</a> 进一步学习。

Meter 是什么？

- Meter 是一个负责创建 Instruments 的对象。它提供了 API 接口，允许开发人员在代码中定义和记录 Metrics。

- 可通过了解 <a href="https://opentelemetry.io/docs/specs/otel/metrics/api/" target="_blank">Metrics API 的概念</a> 进一步学习。

获取 meter：
```python
from opentelemetry import metrics
meter = metrics.get_meter(__name__)
```

#### 3.2.2 Counters

Counters（计数器）用于记录非负递增值。

例如，可以通过以下方式上报请求总数：

```python
#【建议】初始化指标再使用，而不是在业务逻辑里初始化
from opentelemetry import metrics

meter = metrics.get_meter(__name__)
requests_total = meter.create_counter(
    "requests_total",
    description="Total number of HTTP requests",
)

def metrics_counter_demo(self, country: str):
    requests_total.add(1, {"country": country})
```

* <a href="https://opentelemetry.io/docs/languages/python/instrumentation/#creating-and-using-synchronous-instruments" target="_blank">Creating and using synchronous instruments</a>

#### 3.2.3 Histograms

Histograms（直方图）用于记录数值分布情况。

例如，可以通过以下方式上报某段逻辑的处理耗时：

```python
task_execute_duration_seconds = meter.create_histogram(
    "task_execute_duration_seconds",
    unit="s",
    description="Task execute duration in seconds",
)

def metrics_histogram_demo(self):
    start_time = time.time()
    self.do_something(100)
    duration = time.time() - start_time
    task_execute_duration_seconds.record(duration)
```

#### 3.2.4 Gauges

Gauges（仪表）用于记录瞬时值。

例如，可以通过以下方式，上报当前内存使用率（这里使用随机值）：

```python
def generate_random_usage(options: CallbackOptions) -> Iterable[Observation]:
    usage = round(random.random(), 4)
    yield Observation(usage, {})

meter.create_observable_gauge(
    "memory_usage",
    callbacks=[generate_random_usage],
    unit="%",
    description="Memory usage",
)
```

* <a href="https://opentelemetry.io/docs/languages/python/instrumentation/#creating-and-using-asynchronous-instruments" target="_blank">Creating and using asynchronous instruments</a>

### 3.3 Logs

Logs 设计原理可参考：<a href="https://opentelemetry.io/docs/specs/otel/logs/" target="_blank">OpenTelemetry Logging</a>。

Python Logs API & SDK 当前处于实验性状态：<a href="https://opentelemetry.io/docs/languages/python/instrumentation/#logs" target="_blank">Instrumentation/Logs</a>。

示例参考：<a href="https://opentelemetry.io/blog/2023/logs-collection/" target="_blank">Collecting Logs with OpenTelemetry Python</a> 和 <a href="https://github.com/open-telemetry/opentelemetry-python/blob/main/docs/examples/logs/README.rst" target="_blank">OpenTelemetry Logs SDK</a>

#### 3.3.1 Logging 接入

获取 logger：

```python
import logging
otel_logger = logging.getLogger("otel")
```

示例项目采取直接将 OpenTelemetry Protocol (OTLP) 格式的日志发送到目标 collector 的方式。

```python
def _setup_logs(self, resource: Resource):
    otlp_exporter = self._setup_log_exporter()
    self.logger_provider = LoggerProvider(resource=resource)
    self.logger_provider.add_log_record_processor(BatchLogRecordProcessor(otlp_exporter))
    handler = LoggingHandler(level=logging.NOTSET, logger_provider=self.logger_provider)
    logging.getLogger("otel").addHandler(handler)
```

#### 3.3.2 记录日志

```python
import logging

otel_logger = logging.getLogger("otel")


def logs_demo(req: Request):
    otel_logger.info("received request: %s %s", req.method, req.path)
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
-e ENABLE_LOGS="{{access_config.otlp.enable_logs}}" helloworld-python:latest
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