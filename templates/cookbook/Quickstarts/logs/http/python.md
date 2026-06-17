# Python-日志（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://opentelemetry.io/" target="_blank"> opentelemetry 官方文档</a>

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/logs/http/python
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="{{docs.logs.http.readme.http_logs_report}}" target="_blank">自定义日志 HTTP 上报</a> 创建一个上报协议为 `JSON` 的自定义日志，关注创建后提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。

* `API_URL`：国内站点默认是「{{access_config.otlp.http_endpoint}}/v1/logs」，其他环境、跨云场景请根据页面接入指引填写。

### 2.2 样例运行参数

运行参数说明：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `API_URL` | 是 | ❗❗【非常重要】日志上报接口地址（`Access URL`），请根据页面接入指引填写；如果页面提供的是 OTLP HTTP Endpoint 根地址，请在末尾追加 `/v1/logs`。 |
| `TOKEN` | 是 | ❗❗【非常重要】日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。 |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/logs/http/python" target="_blank">bkmonitor-ecosystem/examples/logs/http/python</a> 中找到。

通过 docker build 构建名为 logs-http-python 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、API_URL 传递配置参数，实现周期上报日志：

```bash
docker build -t logs-http-python .

docker run -e TOKEN="fixme" \
 -e API_URL="{{access_config.otlp.http_endpoint}}/v1/logs" \
 logs-http-python
```

运行输出：

```bash
2026-06-16 20:07:13,126 - INFO - response.status_code=200, body={}
2026-06-16 20:07:16,224 - INFO - response.status_code=200, body={}
2026-06-16 20:07:19,263 - INFO - response.status_code=200, body={}
...
```

### 2.4 样例代码

上报代码示例：

```python
import json
import logging
import os
import signal
import sys
import time
from datetime import datetime, timezone

import requests

# ❗❗【非常重要】上报地址，国内站点默认是「{{access_config.otlp.http_endpoint}}/v1/logs」，
# 其他环境、跨云场景请根据页面接入指引填写
API_URL = os.environ.get("API_URL", "{{access_config.otlp.http_endpoint}}/v1/logs")
# ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
TOKEN = os.environ.get("TOKEN", "fixme")


# 日志格式
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)

if not API_URL.startswith("http"):
    logging.warning("API_URL 环境变量未设置或格式不正确，使用默认值: %s", API_URL)
if not TOKEN:
    logging.warning("TOKEN 环境变量未设置，请求头中将不携带 X-BK-TOKEN")


def get_current_nano_timestamp() -> str:
    """返回当前UTC时间的纳秒级Unix时间戳字符串"""
    now = datetime.now(timezone.utc)
    nano = int(now.timestamp() * 1_000_000_000)
    return str(nano)


def build_payload() -> bytes:
    """
    读取 logs.json 模板，更新其中的 timeUnixNano 和 observedTimeUnixNano，
    返回 JSON 字节串
    """
    try:
        with open("logs.json", "r", encoding="utf-8") as f:
            payload = json.load(f)
    except FileNotFoundError:
        logging.error("logs.json not found")
        sys.exit(1)

    current_nano = get_current_nano_timestamp()

    for rl in payload.get("resourceLogs", []):
        for sl in rl.get("scopeLogs", []):
            for lr in sl.get("logRecords", []):
                lr["timeUnixNano"] = current_nano
                lr["observedTimeUnixNano"] = current_nano

    # 去除多余空格
    return json.dumps(payload, separators=(",", ":")).encode("utf-8")


def do_post(data: bytes):
    """发送 POST 请求"""
    headers = {
        "Content-Type": "application/json",
    }
    if TOKEN:
        headers["X-BK-TOKEN"] = TOKEN

    try:
        resp = requests.post(API_URL, data=data, headers=headers, timeout=10)
        logging.info(f"response.status_code={resp.status_code}, body={resp.text}")
    except requests.RequestException as e:
        logging.error(f"failed to post request: {e}")


def main():
    stop_flag = False

    def signal_handler(sig, frame):
        nonlocal stop_flag
        logging.info("Received interrupt signal, exiting...")
        stop_flag = True

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    while not stop_flag:
        payload_bytes = build_payload()
        do_post(payload_bytes)

        # 每3秒发送一次，分段睡眠以快速响应信号
        for _ in range(30):
            if stop_flag:
                break
            time.sleep(0.1)


if __name__ == "__main__":
    main()
```

logs.json 格式示例:

```json
{
    "resourceLogs": [
      {
        "resource": {
          "attributes": [
            {
              "key": "process.name",
              "value": {
                "stringValue": "logger-example"
              }
            }
          ]
        },
        "scopeLogs": [
          {
            "scope": {
              "name": "otlp-logger"
            },
            "logRecords": [
              {
                "timeUnixNano": "0",
                "observedTimeUnixNano": "0",
                "severityNumber": "SEVERITY_NUMBER_INFO",
                "body": {
                  "stringValue": "logs from python http"
                },
                "attributes": [
                  {
                    "key": "count",
                    "value": {
                      "intValue": "6"
                    }
                  }
                ],
                "traceId": "",
                "spanId": ""
              }
            ]
          }
        ]
      }
    ]
  }
```

## 3. 了解更多

进一步了解以下内容：

* 进行 <a href="{{docs.logs.learn_search}}" target="_blank">日志检索</a>。

* 了解 <a href="{{docs.logs.container_custom_report}}" target="_blank">容器日志自定义上报使用文档</a>。

* 了解 <a href="{{docs.logs.container_collector_install}}" target="_blank">容器日志采集器安装</a>。
