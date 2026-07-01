# Python-日志（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/http/README.md" target="_blank">自定义日志 HTTP 上报</a>

* <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs（OTel 日志）</a>：OpenTelemetry 中用于描述离散日志事件的信号类型。

* <a href="https://opentelemetry.io/docs/specs/otel/logs/data-model/" target="_blank">Logs Data Model（OTel 日志数据模型）</a>：定义 `resourceLogs`、`scopeLogs`、`logRecords`、`body`、`attributes`、`severityNumber` 等字段含义。

* <a href="https://opentelemetry.io/docs/specs/otlp/#otlphttp" target="_blank">OTLP/HTTP（OpenTelemetry HTTP 上报协议）</a>：定义通过 HTTP 上报 OTel 数据的协议方式，本示例使用 `/v1/logs` 上报日志。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/logs/http/python
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/http/README.md" target="_blank">自定义日志 HTTP 上报</a> 创建一个上报协议为 `JSON` 的自定义日志，关注创建后提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。

* `API_URL`：国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，其他环境、跨云场景请根据页面接入指引填写。

### 2.2 样例运行参数

运行参数说明：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `API_URL` | 是 | ❗❗【非常重要】日志上报接口地址（`Access URL`），请根据页面接入指引填写；如果页面提供的是 OTLP HTTP Endpoint 根地址，请在末尾追加 `/v1/logs`。 |
| `TOKEN` | 是 | ❗❗【非常重要】日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。 |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/logs/http/python" target="_blank">bkmonitor-ecosystem/examples/logs/http/python</a> 中找到。

通过 docker build 构建名为 logs-http-python 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、API_URL 传递配置参数，实现周期上报日志：

```bash
docker build -t logs-http-python .

docker run -e TOKEN="fixme" \
 -e API_URL="http://127.0.0.1:4318/v1/logs" \
 logs-http-python
```

运行输出：

```bash
2026-06-23 11:50:30,515 - INFO - Starting log reporter (press Ctrl+C to stop)...
2026-06-23 11:50:30,515 - INFO - Sending log level: ERROR (17)
2026-06-23 11:50:30,569 - INFO - response.status_code=200, body={}
2026-06-23 11:50:30,669 - INFO - Sending log level: INFO (9)
2026-06-23 11:50:30,713 - INFO - response.status_code=200, body={}
2026-06-23 11:50:30,813 - INFO - Sending log level: DEBUG (5)
2026-06-23 11:50:30,847 - INFO - response.status_code=200, body={}
2026-06-23 11:50:30,947 - INFO - Sending log level: WARN (13)
2026-06-23 11:50:30,987 - INFO - response.status_code=200, body={}
...
```

### 2.4 样例代码

上报代码示例：

```python
import logging
import os
import random
import time

import requests

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)

# ---------- 日志级别映射 ----------
LOG_LEVELS = [
    {"severityNumber": 5, "severityText": "DEBUG", "message": "debug log from python http"},
    {"severityNumber": 9, "severityText": "INFO", "message": "info log from python http"},
    {"severityNumber": 13, "severityText": "WARN", "message": "warn log from python http"},
    {"severityNumber": 17, "severityText": "ERROR", "message": "error log from python http"},
]


def get_current_nano_timestamp() -> str:
    """返回当前UTC时间的纳秒级Unix时间戳字符串"""
    return str(int(time.time() * 1_000_000_000))


def get_random_level() -> dict:
    """随机返回一个日志级别，用于演示 DEBUG、INFO、WARN、ERROR 日志上报。"""
    return random.choice(LOG_LEVELS)


def build_payload() -> dict:
    """构造 OTLP LogRecord 请求体"""
    current_nano = get_current_nano_timestamp()
    level = get_random_level()

    return {
        "resourceLogs": [
            {
                "resource": {
                    "attributes": [
                        {"key": "service.name", "value": {"stringValue": "custom-log-demo"}},
                        {"key": "deployment.environment.name", "value": {"stringValue": "local"}},
                    ]
                },
                "scopeLogs": [
                    {
                        "scope": {"name": "python-http-demo"},
                        "logRecords": [
                            {
                                "timeUnixNano": current_nano,
                                "observedTimeUnixNano": current_nano,
                                "severityNumber": level["severityNumber"],
                                "severityText": level["severityText"],
                                "body": {"stringValue": level["message"]},
                                "attributes": [
                                    {"key": "demo.source", "value": {"stringValue": "python"}},
                                ],
                            }
                        ],
                    }
                ],
            }
        ]
    }


def do_post(payload: dict) -> None:
    log_record = payload["resourceLogs"][0]["scopeLogs"][0]["logRecords"][0]
    logging.info("Sending log level: %s (%s)", log_record["severityText"], log_record["severityNumber"])

    # ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
    token = os.environ.get("TOKEN", "fixme")
    # ❗❗【非常重要】上报地址，国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，
    # 其他环境、跨云场景请根据页面接入指引填写
    api_url = os.environ.get("API_URL", "http://127.0.0.1:4318/v1/logs")

    headers = {
        "Content-Type": "application/json",
        "x-bk-token": token,
    }

    try:
        resp = requests.post(api_url, json=payload, headers=headers, timeout=10)
        logging.info("response.status_code=%s, body=%s", resp.status_code, resp.text)
    except requests.RequestException as e:
        logging.error("failed to post request: %s", e)


def main():
    logging.info("Starting log reporter (press Ctrl+C to stop)...")
    try:
        while True:
            payload = build_payload()
            do_post(payload)
            time.sleep(0.1)  # 每 0.1 秒上报一条随机级别的日志
    except KeyboardInterrupt:
        logging.info("Received keyboard interrupt, exiting...")


if __name__ == "__main__":
    main()
```

## 3. 了解更多

进一步了解以下内容：

* 进行 <a href="#" target="_blank">日志检索</a>。

* 了解 <a href="#" target="_blank">容器日志自定义上报使用文档</a>。

* 了解 <a href="#" target="_blank">容器日志采集器安装</a>。