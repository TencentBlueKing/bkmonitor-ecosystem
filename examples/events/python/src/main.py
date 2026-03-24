# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

import json
import os
import random
import time
from datetime import datetime

import requests

# ===== 配置层 =====
# ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
API_URL = os.getenv("API_URL", "fixme")
# ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
DATA_ID = int(os.getenv("DATA_ID", 0))
# ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
TOKEN = os.getenv("TOKEN", "fixme")
# 目标设备IP
TARGET_IP = os.getenv("TARGET_IP", "127.0.0.1")
# 上报间隔（秒）
INTERVAL = int(os.getenv("INTERVAL", 60))


# ===== 日志功能 =====
def log(msg):
    """输出带时间戳的日志"""
    print(f"\033[1m{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\033[0m | {msg}")


# ===== 核心功能 =====
def send_events():
    """构造事件数据并发送到上报接口，返回 (status, message)"""
    # 步骤1：构造事件数据
    event_data = [
        {
            "event_name": "cpu_alert",
            "event": {"content": f"CPU告警: {random.randint(80, 99)}%"},
            "target": TARGET_IP,
            "dimension": {"module": "db", "location": "guangdong"},
            "timestamp": int(time.time() * 1000),
        }
    ]
    log(f"生成事件数据:\n{json.dumps(event_data, indent=2, ensure_ascii=False)}")

    # 步骤2：构造请求负载
    # ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
    # ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
    payload = {"data_id": DATA_ID, "access_token": TOKEN, "data": event_data}

    # 步骤3：发送请求
    # ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
    resp = requests.post(API_URL, json=payload, timeout=5, headers={"Content-Type": "application/json"})

    if resp.status_code == 200:
        return "success", "上报成功"
    return "error", f"HTTP {resp.status_code}"


def main():
    log(f"事件上报服务启动 | 目标: {TARGET_IP} | 间隔: {INTERVAL}秒")
    # 持续上报，每次上报后等待 INTERVAL 秒
    while True:
        status, message = send_events()
        info_color = "\033[32m" if status == "success" else "\033[31m"
        log(f"上报结果: {info_color}{status} {message}\033[0m")
        time.sleep(INTERVAL)


# ===== 执行入口 =====
if __name__ == "__main__":
    main()
