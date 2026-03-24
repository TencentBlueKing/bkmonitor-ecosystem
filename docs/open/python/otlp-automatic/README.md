# Python（OpenTelemetry Zero-code Instrumentation）接入

本指南将帮助您通过 Zero-code Instrumentation 接入蓝鲸应用性能监控，以 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/helloworld-automatic/" target="_blank">入门项目-helloworld-automatic</a> 为例，介绍调用链、指标、日志数据接入及 SDK 使用场景。

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
cd bkmonitor-ecosystem/examples/python-examples/helloworld-automatic
docker build -t helloworld-automatic-python:latest .
```

## 2. 快速接入

本节将介绍如何在现有项目基础上，自动探测并附加埋点到应用程序上。我们也通过 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/helloworld-automatic/Dockerfile" target="_blank">Dockerfile</a> 演示了这一过程。

### 2.1 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

### 2.2 接入原理

Python Zero-code Instrument 通过可附加在应用程序上的 Python Agent 实现。Agent 使用 monkey patching 在运行时修改库函数，从而从许多流行的库和框架捕获观测数据。

要理解这个实现原理，必须理解 <a href="https://docs.python.org/zh-cn/3/reference/import.html#the-module-cache" target="_blank">Python 模块的导入原理</a>：

- 当一个模块首次被导入时，Python 会在内存中缓存该模块。

- 后续对同一模块的导入操作实际上是从缓存中获取模块对象。

- 对模块中函数或类的任何动态修改（monkey patching）会在整个应用程序的生命周期生效。

### 2.3 环境依赖

自动安装项目所需 OpenTelemetry instrument 库：

```shell
pip install opentelemetry-distro opentelemetry-exporter-otlp
opentelemetry-bootstrap -a install
```
- 安装 distro 版本才能使用 Zero-code Instrument。

- `opentelemetry-bootstrap -a install` 会读取 site-packages 文件夹中安装的包列表，如果存在包对应的检测库，会自动安装。

- 更多信息请前往 <a href="https://opentelemetry.io/docs/zero-code/python/#setup" target="_blank">Zero-code Instrumentation Setup</a> 查看。

### 2.4 Agent 启动示例

配置环境变量并启动 Agent 示例：

```shell
OTEL_SERVICE_NAME=your-service-name opentelemetry-instrument python myapp.py
```
- 配置 Agent 有两种方式：一种是环境变量（推荐），一种是 CLI 的配置选项。CLI 的配置选项优先级更高。

- 配置选项转换为环境变量的规则：`OTEL_` 前缀 + 配置选项大写形式，例如 `service_name` 转换为 `OTEL_SERVICE_NAME`。

- 查看完整的配置信息请前往 <a href="https://opentelemetry.io/docs/zero-code/python/configuration/" target="_blank">Agent Configuration</a> 查看。

### 2.5 关键配置

#### 2.5.1 环境变量配置

| 环境变量名称                                             | 推荐值                        | 说明                                                                                                                                                                                                                                                                                                                                                                                               |
|----------------------------------------------------|----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `OTEL_SERVICE_NAME`                                | `"${服务名称，请根据右侧说明填写}"`   | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分。优先级比资源属性的设置高，更多信息请参考<a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name" target="_blank">服务名配置</a>。                                                                                                                                                                                                                     |
| `OTEL_EXPORTER_OTLP_HEADERS`                         | `"x-bk-token=todo"` | 【必须】Exporter 导出数据时附加额外的 Headers，多个 key-value 以逗号分隔，例如：`key1=value1,key2=value2`。</br>【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。 |
| `OTEL_EXPORTER_OTLP_PROTOCOL`                      | `"http/protobuf"`          | 【必须】指定<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#protocol-configuration" target="_blank">上报协议</a>，上报协议改变时，上报地址也需要手动修改。</br>【推荐】`protobuf/http`：使用 HTTP 协议上报。</br>【可选】`grpc`：使用 gRPC 上报，如果使用该方式，请确保 `OTEL_EXPORTER_OTLP_ENDPOINT` 也同步改为 gRPC 上报地址。                                                                                                          |
| `OTEL_EXPORTER_OTLP_ENDPOINT`                      | `"http://127.0.0.1:4318"`                         | 【必须】数据<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_endpoint" target="_blank">上报地址</a>，请根据页面指引提供的接入地址进行填写。支持以下协议：<br />`gRPC`：`http://127.0.0.1:4317`<br />`HTTP`：`http://127.0.0.1:4318`（demo 使用该协议演示上报）<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。                                                                                                                                                                                                                   |
| `OTEL_TRACES_EXPORTER`                             | `"otlp"`                   | 【可选】指定用于 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_traces_exporter" target="_blank">Traces Exporter</a>，值为 `"console,otlp"` 时，可以同时在控制台输出。                                                                                                                                                                                                                  |
| `OTEL_METRICS_EXPORTER`                            | `"otlp"`                   | 【可选】指定用于 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_metrics_exporter" target="_blank">Metrics Exporter</a>，值为 `"console,otlp"` 时，可以同时在控制台输出。                                                                                                                                                                                                                |
| `OTEL_LOGS_EXPORTER`                               | `"otlp"`                   | 【可选】指定用于 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_logs_exporter" target="_blank">Logs Exporter</a>，值为 `"console,otlp"` 时，可以同时在控制台输出。 |
| `OTEL_RESOURCE_ATTRIBUTES`                         | `""` | 【可选】<a href="https://opentelemetry.io/docs/concepts/resources/" target="_blank">Resource</a> 代表观测数据所属的资源实体，并通过资源属性进行描述。<a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_resource_attributes" target="_blank">Resource Attributes</a> 设置，多个 key-value 以逗号分隔，例如：`key1=value1,key2=value2`。参考下一小节 `服务信息`。 |
| `OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED` | `"true"`                   | 【可选】值为 "true" 时启用日志自动检测。如果需要调整日志级别、格式等，请参考<a href="https://opentelemetry.io/docs/zero-code/python/configuration/#logging" target="_blank">这里</a>。                                                                                                                                                                                                                                                |

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

## 3. 使用场景

Zero-code Instrumentation 已自动设置 OpenTelemetry 所需的 Resource、Exporter 对象，并对常见库进行插桩，即在无需修改代码的情况下，能上报常用库的观测数据。

如果需要在程序中埋点上报更多数据，请参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/python/otlp/README.md#3-使用场景" target="_blank">Python（OpenTelemetry SDK）接入「3. 使用场景」</a>部分。

## 4. 快速体验

### 4.1 运行样例

运行前注意事项：

- 运行之前请记得执行 `docker build` 命令，参考本文 1.3 节。

- 如果是本地开发测试，请确保您已运行快速验证 demo 数据上报逻辑，快速开始 👉 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/common/ob-all-in-one" target="_blank">ob-all-in-one</a>。

复制以下命令参数在你的终端运行：

```shell
docker run \
-e OTEL_SERVICE_NAME="helloworld-automatic" \
-e OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo" \
-e OTEL_EXPORTER_OTLP_PROTOCOL="http/protobuf" \
-e OTEL_EXPORTER_OTLP_ENDPOINT="http://127.0.0.1:4318" helloworld-automatic-python:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

如果您运行命令是要接入蓝鲸 APM 平台，那么请务必注意以下事项：

- 【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。

- 【必须】`OTEL_EXPORTER_OTLP_ENDPOINT`：数据上报地址，请根据页面指引提供的接入地址进行填写。

- 【必须】`OTEL_EXPORTER_OTLP_PROTOCOL`：如果使用 gRPC 上报，请确保 `OTEL_EXPORTER_OTLP_ENDPOINT` 同步改为 gRPC 上报地址。

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