# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

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
