# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

import os
import random
import time

import requests


def collect_metrics():
    """采集模拟周期上报 CPU 及内存使用率（数值随机生成）"""
    cpu_load = random.uniform(0, 100)
    mem_usage = random.uniform(0, 100)
    return cpu_load, mem_usage


def send_report(api_url, token, data_id, metrics):
    """发送数据到HTTP接口"""
    payload = {
        "data_id": int(data_id),  # ❗❗【非常重要】必须是整数类型，否则上报会失败
        "access_token": token,
        "data": [
            {
                "metrics": {"cpu_load": metrics[0], "memory_usage": metrics[1]},
                "target": "127.0.0.1",
                "dimension": {"module": "server", "region": "guangdong", "language": "python"},
            }
        ],
    }
    try:
        response = requests.post(api_url, json=payload, timeout=10, headers={"Content-Type": "application/json"})
        return response.status_code
    except requests.exceptions.RequestException as e:
        print(f"⚠️ 请求异常: {e}")
        return 500


def main():
    # ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 {{access_config.custom.http}} 」
    # 其他环境、跨云场景请根据页面接入指引填写
    api_url = os.getenv("API_URL", "")
    token = os.getenv("TOKEN", "")  # ❗❗【非常重要】access_token:认证令牌，用于接口鉴定，配置为应用 TOKEN
    data_id = os.getenv("DATA_ID", "")  # ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
    interval = int(os.getenv("INTERVAL", "60"))  # 默认间隔60秒
    while True:
        metrics = collect_metrics()
        status = send_report(api_url, token, data_id, metrics)

        timestamp = time.strftime("%Y-%m-%d %H:%M:%S")
        if status == 200:
            print(f"[{timestamp}] ✅ 上报成功 | CPU: {metrics[0]:.2f}% 内存: {metrics[1]:.2f}%")
        else:
            print(f"[{timestamp}] ❌ 上报失败 | 状态码: {status}")

        time.sleep(interval)


if __name__ == "__main__":
    main()
