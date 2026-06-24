# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

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
    # ❗❗【非常重要】上报地址，国内站点默认是「 {{access_config.otlp.http_endpoint}}/v1/logs 」，
    # 其他环境、跨云场景请根据页面接入指引填写
    api_url = os.environ.get("API_URL", "{{access_config.otlp.http_endpoint}}/v1/logs")

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
