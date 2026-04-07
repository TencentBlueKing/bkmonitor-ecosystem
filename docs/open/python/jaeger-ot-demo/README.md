# Python（Jaeger OpenTracing Shim）接入

本指南将帮助使用 Jaeger client 上报数据的用户，平滑过渡到使用 OpenTelemetry SDK 接入蓝鲸应用性能监控，以 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/jaeger-ot-demo/" target="_blank">入门项目-jaeger-ot-demo</a> 为例，介绍调用链接入及 SDK 使用场景。

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
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/python-examples/jaeger-ot-demo
docker build -t jaeger-ot-demo-py:latest .
```

## 2. 快速接入

Jaeger Client 是 OpenTracing API 规范的具体实现，目前业界标准已从 OpenTracing 演进为 OpenTelemetry（融合 OpenTracing 和 OpenCensus）。本文详解 Jaeger Client 到 OpenTelemetry SDK 的最小化迁移方案，我们也通过 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/jaeger-ot-demo/Dockerfile" target="_blank">Dockerfile</a> 演示了这一过程。

### 2.1 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

### 2.2 环境依赖

安装 OpenTelemetry API、OpenTelemetry SDK 和项目所需 OpenTelemetry instrument 库：

```shell
pip install opentelemetry-distro opentelemetry-exporter-otlp
opentelemetry-bootstrap -a install
```

安装 Jaeger Client 和 OpenTracing shim 库：

```shell
pip install jaeger-client
pip install opentelemetry-opentracing-shim
```
- 迁移完成后可移除 jaeger-client 依赖项，但应在过渡期保持其作为备用采集通道，通过流量灰度控制策略（如按服务版本/用户群分流）实现数据双向验证，保障可观测数据的完整性和一致性。
- <a href="https://github.com/open-telemetry/opentelemetry-python/tree/main/shim/opentelemetry-opentracing-shim" target="_blank">OpenTracing Shim</a> 提供双向适配层，维持 OpenTracing API 兼容性的同时桥接 OpenTelemetry SDK。

### 2.3 OpenTelemetry SDK 配置

OpenTelemetry 提供标准化的框架和工具包，用于创建和管理 Traces、Metrics、Logs 数据。示例项目集成 OpenTelemetry SDK 并将观测数据发送到 bk-collector。

样例代码 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/jaeger-ot-demo/src/services/otlp.py" target="_blank">jaeger-ot-demo services/otlp.py</a> 只演示上报 Traces 数据的配置，完整的配置可以参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/helloworld/src/services/otlp.py" target="_blank">helloworld services/otlp.py</a> 进行接入。

### 2.4 项目 tracer 修改

将 Jaeger tracer 切换到 OpenTracing-Shim 提供的 OpenTelemetry tracer 实现，确保向后兼容的前提下接入现代可观测性生态。

```python
from opentelemetry import trace
from opentelemetry.shim.opentracing_shim import create_tracer, TracerShim
from opentracing import set_global_tracer

def init_tracer() -> TracerShim:
    # 注释代码是使用 jaeger_client 上报的样例，用来同迁移到 OTel SDK 的代码做对比
    # from jaeger_client import Config
    # from config import config as custom_config
    # config = Config(
    #     config={
    #         'sampler': {
    #             'type': 'const',
    #             'param': 1,
    #         },
    #         'logging': True,
    #         "reporter_queue_size":10,
    #     },
    #     service_name=custom_config.service_name,
    #     validate=True,
    # )
    # return config.initialize_tracer()
    # 获取全局 Tracer Provider
    global_tracer_provider = trace.get_tracer_provider()

    # Create an OpenTracing shim.
    shim_tracer = create_tracer(global_tracer_provider)
    set_global_tracer(shim_tracer)
    return shim_tracer
```

### 2.5 关键配置

🌟 请仔细阅读本章节，以确保观测数据能准确上报到 APM。

#### 2.5.1 上报地址 & 应用 Token

请在创建 <a href="https://opentelemetry.io/docs/specs/otel/protocol/exporter/" target="_blank">Exporter</a> 时准确传入以下信息：

| 参数         | 说明                            |
|------------|-------------------------------|
| `endpoint` | 【必须】数据上报地址，请根据页面指引提供的接入地址进行填写。 |
| `x-bk-token`| 【必须】APM 应用 Token，作为 headers 传递。 |

示例项目在 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/jaeger-ot-demo/src/services/otlp.py" target="_blank">services/otlp.py _setup_traces</a> 提供了创建样例：

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

#### 2.5.2 服务信息

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

示例项目在 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/jaeger-ot-demo/src/services/otlp.py" target="_blank">services/otlp.py _create_resource</a> 提供了创建样例：

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

当前示例项目聚焦于 OpenTelemetry 和 Jaeger Client 的 Traces 应用场景，如需探索 OpenTelemetry SDK 的完整可观测性能力（包括指标、日志等），请参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/helloworld/" target="_blank">helloworld 项目</a> 进行实现。

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
    detectors = [ProcessResourceDetector()]
    if OsResourceDetector is not None:
        detectors.append(OsResourceDetector())

    # create 提供了部分 SDK 默认属性
    initial_resource = Resource.create(
        {
            # ❗❗【非常重要】应用服务唯一标识
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

- 可通过了解 <a href="https://opentelemetry.io/docs/specs/otel/trace/api/" target="_blank">OpenTelemetry Tracing API 的概念</a> 和 <a href="https://opentracing.io/docs/overview/tracers/" target="_blank">OpenTracing tracers</a> 进一步学习。

必须完成 OpenTelemetry SDK 的初始化配置后，才能获取可用的 OpenTelemetry tracer。
```python
# 获取 OpenTracing 的 tracer
from jaeger_tracer import init_tracer
tracer = init_tracer()
# 获取 OpenTelemetry 的 tracer
from opentelemetry import trace
tracer = trace.get_tracer(__name__)
```

#### 3.1.3 创建 Span

Span 是 Traces 的构建块，代表一个工作或操作单元。

OpenTracing Span 通过 `tracer.start_active_span()` 进行创建：

```python
def traces_custom_span_demo(self):
    with self.tracer.start_active_span("custom_span_demo/do_something") as scope:
        span = scope.span
        self.do_something(50)
```

​​<a href="https://opentelemetry.io/docs/languages/python/instrumentation/#creating-spans" target="_blank">创建 OpenTelemetry Span</a> 使用 `tracer.start_as_current_span()` 方法实现：

```python
def traces_custom_span_demo(self):
    with self.tracer.start_as_current_span("custom_span_demo/do_something") as span:
        self.do_something(50)
```

#### 3.1.4 设置属性

Attributes（属性）是 Span 元数据，以 Key-Value 形式存在。

在 Span 设置属性，对问题定位、过滤、聚合非常有帮助。

OpenTracing 属性通过 `span.set_tag()` 进行创建：

```python
span.set_tag("helloworld.kind", 1)
span.set_tag("helloworld.step", "traces_custom_span_demo")
```

<a href="https://opentelemetry.io/docs/languages/python/instrumentation/#add-attributes-to-a-span" target="_blank">增加 Opentelemetry 属性</a> 使用 `span.set_attribute()` 方法实现：
```python
# 增加 Span 自定义属性
span.set_attribute("helloworld.kind", 1)
span.set_attribute("helloworld.step", "traces_custom_span_demo")
```

#### 3.1.5 设置事件

Event（事件）是一段人类可读信息，用于记录 Span 生命周期内发生的事情。

OpenTracing 事件属性通过 `span.log_kv()` 进行创建：
```python
span.log_kv({
    "helloworld.kind": 2,
    "helloworld.step": "traces_span_event_demo",
})

```

<a href="https://opentelemetry.io/docs/languages/python/instrumentation/#adding-events" target="_blank">增加 Opentelemetry 事件属性</a> 使用 `span.add_event()` 方法实现：
```python
attributes = {
    "helloworld.kind": 2,
    "helloworld.step": "traces_span_event_demo",
}
span.add_event("Before do_something", attributes)
self.do_something(50)
span.add_event("After do_something", attributes)
```

## 4. 快速体验

### 4.1 运行样例

#### 4.1.1 运行

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="xxx" \
-e SERVICE_NAME="helloworld" \
-e OTLP_ENDPOINT="http://127.0.0.1:4318" \
-e ENABLE_TRACES="true" jaeger-ot-demo-py:latest
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
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/traces.png)

## 5. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>