# Profiling-Python（Pyroscope SDK）接入

本指南将帮助您使用 Pyroscope SDK 接入蓝鲸应用性能监控，以入门项目-Profiling 为例，介绍性能分析数据接入及 SDK 使用场景。

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
cd bkmonitor-ecosystem/examples/python-examples/profiling
docker build -t profiling-python:latest .
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/profiling-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `PROFILING_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

### 2.2 环境依赖

安装 `pyroscope-io` 包：

```shell
pip install pyroscope-io==0.8.8
```

### 2.3 Pyroscope SDK

示例项目使用 <a href="https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/python/" target="_blank">pyroscope-io</a> 指定的配置方式，将性能数据发送到 bk-collector。

可以参考 `main.py` 文件进行接入:

```python
import pyroscope

pyroscope.configure(
    # 服务名，一个应用可以有多个服务，通过该属性区分。
    application_name=config.service_name,
    # ❗❗【非常重要】数据上报地址，请根据页面指引提供的 Profiling 接入地址进行填写
    server_address=config.endpoint,
    tags={
        "service.name": config.service_name,
        "service.version": "0.1",
        "service.environment": "dev",
        "net.host.ip": "127.0.0.1",
        "net.host.name": "localhost",
    },
    http_headers={
        # ❗❗【非常重要】`X-BK-TOKEN` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。
        "X-BK-TOKEN": config.token,
    },
)
```

## 3. 快速体验

### 3.1 运行样例

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
-e ENABLE_PROFILING="true" profiling-python:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

### 3.2 查看数据

等待片刻，便可在「服务详情-Profiling」看到应用数据。

![](image/img.png)

## 4. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>