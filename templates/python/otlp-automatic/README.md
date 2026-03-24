# Python（OpenTelemetry Zero-code Instrumentation）接入

本指南将帮助您通过 Zero-code Instrumentation 接入蓝鲸应用性能监控，以 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/python-examples/helloworld-automatic/" target="_blank">入门项目-helloworld-automatic</a> 为例，介绍调用链、指标、日志数据接入及 SDK 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速入门 OpenTelemetry，接入并体验蓝鲸应用性能监控相关功能。

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
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/python-examples/helloworld-automatic
docker build -t helloworld-automatic-python:latest .
```

## 2. 快速接入

本节将介绍如何在现有项目基础上，自动探测并附加埋点到应用程序上。我们也通过 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/python-examples/helloworld-automatic/Dockerfile" target="_blank">Dockerfile</a> 演示了这一过程。

### 2.1 创建应用

{{APPLICATION_SET_UP}}

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

{{AUTOMATIC_RUN_PARAMETERS}}
| `OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED` | `"true"`                   | 【可选】值为 "true" 时启用日志自动检测。如果需要调整日志级别、格式等，请参考<a href="https://opentelemetry.io/docs/zero-code/python/configuration/#logging" target="_blank">这里</a>。                                                                                                                                                                                                                                                |

#### 2.5.2 服务信息

{{MUST_CONFIG_RESOURCES}}

## 3. 使用场景

Zero-code Instrumentation 已自动设置 OpenTelemetry 所需的 Resource、Exporter 对象，并对常见库进行插桩，即在无需修改代码的情况下，能上报常用库的观测数据。

如果需要在程序中埋点上报更多数据，请参考 <a href="{{REFER_PYTHON_OTLP_URL}}#3-使用场景" target="_blank">Python（OpenTelemetry SDK）接入「3. 使用场景」</a>部分。

## 4. 快速体验

### 4.1 运行样例

运行前注意事项：

- 运行之前请记得执行 `docker build` 命令，参考本文 1.3 节。

- 如果是本地开发测试，请确保您已运行快速验证 demo 数据上报逻辑，快速开始 👉 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/common/ob-all-in-one" target="_blank">ob-all-in-one</a>。

复制以下命令参数在你的终端运行：

```shell
docker run \
-e OTEL_SERVICE_NAME="helloworld-automatic" \
-e OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo" \
-e OTEL_EXPORTER_OTLP_PROTOCOL="http/protobuf" \
-e OTEL_EXPORTER_OTLP_ENDPOINT="{{access_config.otlp.http_endpoint}}" helloworld-automatic-python:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

如果您运行命令是要接入蓝鲸 APM 平台，那么请务必注意以下事项：

- 【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。

- 【必须】`OTEL_EXPORTER_OTLP_ENDPOINT`：数据上报地址，请根据页面指引提供的接入地址进行填写。

- 【必须】`OTEL_EXPORTER_OTLP_PROTOCOL`：如果使用 gRPC 上报，请确保 `OTEL_EXPORTER_OTLP_ENDPOINT` 同步改为 gRPC 上报地址。

### 4.2 查看数据

#### 4.2.1 Traces 检索

{{VIEW_TRACES_DATA}}

#### 4.2.2 指标检索

{{VIEW_CUSTOM_METRICS_DATA}}

#### 4.2.3 日志检索

{{VIEW_LOG_DATA}}

## 5. 了解更多

{{LEARN_MORE}}