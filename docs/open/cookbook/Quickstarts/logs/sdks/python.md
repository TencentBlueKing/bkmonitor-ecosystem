# Python-日志（SDK）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="docs.logs.learn.sdk_logs_report" target="_blank">自定义日志 OpenTelemetry SDK 上报</a>

* <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs（OTel 日志）</a>：OpenTelemetry 中用于描述离散日志事件的信号类型。

* <a href="https://opentelemetry.io/docs/specs/otel/logs/data-model/" target="_blank">Logs Data Model（OTel 日志数据模型）</a>：定义 `resourceLogs`、`scopeLogs`、`logRecords`、`body`、`attributes`、`severityNumber` 等字段含义。

* <a href="https://opentelemetry.io/docs/specs/otlp/#otlphttp" target="_blank">OTLP/HTTP（OpenTelemetry HTTP 上报协议）</a>：定义通过 HTTP 上报 OTel 数据的协议方式，Python SDK 通过 `/v1/logs` 上报日志。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/logs/sdks/python
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="docs.logs.learn.sdk_logs_report" target="_blank">自定义日志 OpenTelemetry SDK 上报</a> 创建一个上报协议为 `JSON` 的自定义日志，关注创建后提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。

* `API_URL`：国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，其他环境、跨云场景请根据页面接入指引填写。

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `API_URL` | 是 | ❗❗【非常重要】日志上报接口地址（`Access URL`），请根据页面接入指引填写；如果页面提供的是 OTLP HTTP Endpoint 根地址，请在末尾追加 `/v1/logs`。 |
| `TOKEN` | 是 | ❗❗【非常重要】日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。 |

#### 2.2.1 关键配置

使用 OpenTelemetry Python SDK 进行日志上报时，需正确配置以下核心参数：

```python
from opentelemetry.sdk._logs import LoggerProvider, LoggingHandler
from opentelemetry.sdk._logs.export import BatchLogRecordProcessor
from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter

# ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
token = os.environ.get("TOKEN", "fixme")
# ❗❗【非常重要】上报地址，国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，
# 其他环境、跨云场景请根据页面接入指引填写
api_url = os.environ.get("API_URL", "http://127.0.0.1:4318/v1/logs")

# 创建 OTLP HTTP Log Exporter，注入认证 Header
log_exporter = OTLPLogExporter(
    endpoint=api_url,
    headers={"x-bk-token": token},
)

# 创建 LoggerProvider，关联资源属性（服务名、环境等）
logger_provider = LoggerProvider(
    resource=Resource.create({
        "service.name": "custom-log-demo",
        "deployment.environment": "local",
    })
)
logger_provider.add_log_record_processor(BatchLogRecordProcessor(log_exporter))

# 将 Python 标准 logging 桥接到 OTel LoggerProvider
handler = LoggingHandler(level=logging.NOTSET, logger_provider=logger_provider)
logging.getLogger("otel").addHandler(handler)
```

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/master/examples/logs/sdks/python" target="_blank">bkmonitor-ecosystem/examples/logs/sdks/python</a> 中找到。

通过 docker build 构建名为 logs-sdk-python 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、API_URL 传递配置参数，实现周期上报日志：

```bash
docker build -t logs-sdk-python .

docker run -e TOKEN="fixme" \
 -e API_URL="http://127.0.0.1:4318/v1/logs" \
 logs-sdk-python
```

运行输出：

```bash
2026-09-18 11:50:30,515 - INFO - Starting log reporter (press Ctrl+C to stop)...
2026-09-18 11:50:30,523 - INFO - info log from python sdk
2026-09-18 11:50:30,632 - WARNING - warn log from python sdk
2026-09-18 11:50:30,741 - ERROR - error log from python sdk
2026-09-18 11:50:30,844 - DEBUG - debug log from python sdk
2026-09-18 11:50:30,952 - INFO - info log from python sdk
...
```

### 2.4 样例代码

该样例使用 OpenTelemetry Python SDK，通过 OTLP/HTTP 协议上报日志。代码集成 `opentelemetry-sdk` 和 `opentelemetry-exporter-otlp-proto-http`，将 Python 标准 `logging` 模块桥接到 OTel LoggerProvider，实现自动采集与上报。

```python
import logging
import os
import random
import time

from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter
from opentelemetry.sdk._logs import LoggerProvider, LoggingHandler
from opentelemetry.sdk._logs.export import BatchLogRecordProcessor
from opentelemetry.sdk.resources import Resource

# ---------- 基础配置 ----------
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger("otel")

# ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
TOKEN = os.environ.get("TOKEN", "fixme")
# ❗❗【非常重要】上报地址，国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，
# 其他环境、跨云场景请根据页面接入指引填写
API_URL = os.environ.get("API_URL", "http://127.0.0.1:4318/v1/logs")

# ---------- 初始化 OTel LoggerProvider ----------
log_exporter = OTLPLogExporter(
    endpoint=API_URL,
    headers={"x-bk-token": TOKEN},
)

logger_provider = LoggerProvider(
    resource=Resource.create({
        "service.name": "custom-log-sdk-demo",
        "deployment.environment": "local",
    })
)
logger_provider.add_log_record_processor(BatchLogRecordProcessor(log_exporter))

# 将 Python logging 桥接到 OpenTelemetry
handler = LoggingHandler(level=logging.NOTSET, logger_provider=logger_provider)
logger.addHandler(handler)

LOG_LEVELS = [
    (logging.DEBUG, "debug log from python sdk"),
    (logging.INFO, "info log from python sdk"),
    (logging.WARNING, "warn log from python sdk"),
    (logging.ERROR, "error log from python sdk"),
]


def main():
    logger.info("Starting log reporter (press Ctrl+C to stop)...")
    try:
        while True:
            level, message = random.choice(LOG_LEVELS)
            logger.log(level, message)
            time.sleep(0.1)  # 每 0.1 秒上报一条随机级别的日志
    except KeyboardInterrupt:
        logger.info("Received keyboard interrupt, exiting...")


if __name__ == "__main__":
    main()
```

## 3. 了解更多

进一步了解以下内容：

* 进行 <a href="#" target="_blank">日志检索</a>。

* 了解 <a href="#" target="_blank">容器日志自定义上报使用文档</a>。

* 了解 <a href="#" target="_blank">容器日志采集器安装</a>。