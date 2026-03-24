# -*- coding: utf-8 -*-
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

from prometheus_client import CollectorRegistry, Counter, Gauge, Histogram, Summary, push_to_gateway, start_http_server
from prometheus_client.exposition import default_handler

# 基础配置
registry = CollectorRegistry()
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("prom_demo")

# 环境变量配置
TOKEN = os.getenv("TOKEN", "")  # ❗️❗️【非常重要】请填写为申请到的自定义指标认证令牌（`Token`）。
API_URL = os.getenv("API_URL", "")  # ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写。
JOB = os.getenv("JOB", "default_monitor_job")  # 任务名称
INSTANCE = os.getenv("INSTANCE", "127.0.0.1")  # 实例名称
INTERVAL = int(os.getenv("INTERVAL", "60"))  # 默认60秒

METRICS_PORT = int(os.getenv("METRICS_PORT", "2323"))  # 默认2323端口暴露/metrics端点


# 认证Handler
def bk_handler(url, method, timeout, headers, data):
    """添加认证头部信息"""
    if TOKEN:
        # ❗️❗️【非常重要】请填写为申请到的自定义指标认证令牌（`Token`）。
        headers.append(("X-BK-TOKEN", TOKEN))
    headers.append(("Content-Type", "text/plain"))
    return default_handler(url, method, timeout, headers, data)


# 带认证的安全推送
def safe_push():
    """安全推送指标到Pushgateway"""
    try:
        push_to_gateway(
            # ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写。
            gateway=API_URL,
            job=JOB,
            registry=registry,
            grouping_key={"instance": INSTANCE, "language": "python"},
            # ❗️❗️【非常重要】将申请到的自定义指标认证令牌（Token）设置为 HTTP 请求头（X-BK-TOKEN），
            # 该步骤将决定数据能否准确被蓝鲸监控平台接收。
            handler=bk_handler,
        )
        logger.info(f"成功推送指标到 {API_URL}")
    except Exception as e:  # pylint: disable=broad-except
        logger.error(f"推送失败: {str(e)}")


# ===== 指标类型定义与演示函数 =====

# Counter类型 - API调用统计
# Refer：https://prometheus.github.io/client_python/instrumenting/counter/
api_counter = Counter("api_called_total", "API调用总次数", ["api_name", "status"], registry=registry)


def counter_demo():
    """模拟API调用计数"""
    status = "200" if random.random() > 0.9 else "500"  # 10%错误率
    api_name = random.choice(["/user/login", "/data/query", "/order/create"])
    api_counter.labels(api_name=api_name, status=status).inc()
    logger.debug(f"记录API调用: {api_name} | 状态: {status}")


# Gauge类型 - CPU使用率监控
# Refer：https://prometheus.github.io/client_python/instrumenting/gauge/
cpu_gauge = Gauge("cpu_usage_percent", "CPU使用率百分比", ["host"], registry=registry)


def gauge_demo():
    """模拟CPU使用率波动"""
    host = f"host{random.randint(1, 3)}"
    usage = round(random.uniform(5.0, 95.0), 1)
    cpu_gauge.labels(host=host).set(usage)
    logger.debug(f"记录CPU使用率: {host} | 使用率: {usage}%")


# Histogram类型 - 任务耗时分布
# Refer：https://prometheus.github.io/client_python/instrumenting/histogram/
task_histogram = Histogram(
    "task_duration_seconds",
    "任务耗时分布",
    ["task_type"],
    registry=registry,
    buckets=[0.1, 0.5, 1, 2, 5],  # 关键耗时阈值
)


def histogram_demo():
    """记录任务耗时分布"""
    task_type = random.choice(["import", "export", "process"])
    duration = random.uniform(0.05, 6.0)  # 模拟任务执行

    # 自动计时并分桶统计
    with task_histogram.labels(task_type=task_type).time():
        time.sleep(duration)
    logger.debug(f"记录任务耗时: {task_type} | 耗时: {duration:.2f}s")


# Summary类型 - 处理时间摘要
# Refer：https://prometheus.github.io/client_python/instrumenting/summary/
process_summary = Summary("task_processing_seconds", "任务处理时间摘要", ["stage"], registry=registry)


def summary_demo():
    """记录处理阶段耗时摘要"""
    stage = random.choice(["validation", "execution", "cleanup"])
    duration = random.uniform(0.1, 3.0)
    process_summary.labels(stage=stage).observe(duration)
    logger.debug(f"记录处理阶段: {stage} | 耗时: {duration:.2f}s")


# ===== 主执行逻辑 =====


def main():
    """主执行函数 -  同时支持Pull模式与Push模式"""
    start_http_server(METRICS_PORT)
    logger.info(f"已启用Pull模式 | 指标端点: http://127.0.0.1:{METRICS_PORT}/metrics")

    logger.info(f"启动指标上报服务 | 实例: {INSTANCE} | 任务: {JOB}")
    logger.info(f"目标地址: {API_URL} | 认证令牌: {'已配置' if TOKEN else '未配置'}")
    logger.info(f"上报间隔: {INTERVAL}秒")

    while True:
        start_time = time.time()

        # 执行各指标上报函数
        counter_demo()
        gauge_demo()
        histogram_demo()
        summary_demo()

        # 推送指标
        safe_push()

        # 使用自定义间隔计算等待时间
        elapsed = time.time() - start_time
        sleep_time = max(INTERVAL - elapsed, 1)  # 确保至少间隔1秒
        logger.info(f"本轮上报完成 | 耗时: {elapsed:.2f}s | 下次上报: {sleep_time:.0f}s后")
        time.sleep(sleep_time)


if __name__ == "__main__":
    main()
