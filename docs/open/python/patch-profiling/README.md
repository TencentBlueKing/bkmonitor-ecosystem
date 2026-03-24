# Profiling-Python（Datadog SDK）接入

本指南将帮助您使用 Datadog SDK 接入蓝鲸应用性能监控，以入门项目-Profiling 为例，介绍性能分析数据接入及 SDK 使用场景。

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
cd bkmonitor-ecosystem/examples/python-examples/patch-profiling
docker build -t patch-profiling-python:latest .
```

## 2. 快速体验

### 2.1 运行样例

运行参数说明：

| 参数                   | 推荐值                                | 说明                                        |
|----------------------|------------------------------------|-------------------------------------------|
| `TOKEN`              | `""`                               | APM 应用 `Token`。                            |
| `PROFILING_ENDPOINT` | `"http://127.0.0.1:4318/pyroscope"` | Profiling 数据上报地址，请根据页面指引提供的 HTTP 接入地址进行填写。<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。 |
| `SERVICE_NAME`       | `"helloworld"`                     | 服务唯一标识，一个应用可以有多个服务，通过该属性区分。                |
| `ENABLE_PROFILING`   | `false`                            | 是否启用性能分析上报。                                |

💡 为保证数据能上报到平台，`TOKEN`、`PROFILING_ENDPOINT` 请务必根据应用接入指引提供的实际值填写。

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="xxx" \
-e SERVICE_NAME="helloworld" \
-e PROFILING_ENDPOINT="http://127.0.0.1:4318/pyroscope" \
-e ENABLE_PROFILING="true" \
-e ENABLE_MEMORY_PROFILING="true" patch-profiling-python:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

### 2.2 查看数据

等待片刻，便可在「服务详情-Profiling」看到应用数据。

![](image/img.png)

## 3. 快速接入

### 3.1 Datadog SDK

示例项目提供集成 Datadog Python SDK 并通过自定义导出器，将性能数据发送到 bk-collector 的方式。
这是通过 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/patch-profiling/src/patch.py" target="_blank">patch.py</a> 模块中的 patch_ddtrace_to_pyroscope 补丁实现的。其他项目如有同样需求，可以简单复制该模块来使用补丁。
使用方法可以参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/python-examples/patch-profiling/src/main.py" target="_blank">main.py</a>:

```python
from ddtrace.profiling.profiler import Profiler

from patch import patch_ddtrace_to_pyroscope


patch_ddtrace_to_pyroscope(
    # ❗❗【非常重要】应用服务唯一标识
    service_name=config.service_name,
    # ❗❗【非常重要】请传入应用 Token
    token=config.token,
    # ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
    endpoint=config.profiling_endpoint,
    enable_memory_profiling=config.enable_memory_profiling,
)
prof = Profiler()
prof.start()
```

参考官方文档以获得更多信息：<a href="https://docs.datadoghq.com/profiler/enabling/python/" target="_blank">Enabling the Python Profiler</a>

## 4. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>