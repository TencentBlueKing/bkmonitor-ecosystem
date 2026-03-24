# Python-指标（Prometheus）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="{{docs.metrics.What}}" target="_blank">什么是指标</a>

* <a href="{{docs.metrics.Types}}" target="_blank">指标类型</a>

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/metrics/sdks/python
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="{{docs.metrics.learn.metrics_sdks_readme}}" target="_blank">自定义指标 Prometheus SDK 上报</a> 创建一个上报协议为 `Prometheus` 的自定义指标，关注创建后提供的配置项：

* `TOKEN`：数据源 Token，后续需要在上报数据时使用。

同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image-1.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数     | 类型                | 描述                         |
| ------------ | ------------------- | ---------------------------- |
|`TOKEN`        |String      | ❗❗【非常重要】 自定义指标数据源 `Token`。 |
|`API_URL`       |String      | ❗❗【非常重要】 数据上报接口地址（`Access URL`），国内站点请填写「 {{access_config.custom.sdk}} 」，其他环境、跨云场景请根据页面接入指引填写。|
|`INTERVAL`      |Integer  　 |数据上报间隔，默认值为 60 秒。  ​​ |
|`METRICS_PORT`  |Integer  　 |指标暴露端口，默认 2323。|

#### 2.2.1 关键配置

蓝鲸监控支持原生 Prometheus 协议，如果业务已接入 Prometheus SDK，只需在 `push_to_gateway` 方法，修改上报地址为 `API_URL`，增加注入 `X-BK-TOKEN` 的 handler。

```python
from prometheus_client import push_to_gateway, default_handler

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
            grouping_key={"instance": INSTANCE},
             # ❗️❗️【非常重要】将申请到的自定义指标认证令牌（Token）到 HTTP 请求头（X-BK-TOKEN），该步骤将决定数据能否准确被蓝鲸监控平台接收。
            handler=bk_handler
        )
        logger.info(f"成功推送指标到 {API_URL}")
    except Exception as e:
        logger.error(f"推送失败: {str(e)}")
```

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/metrics/sdks/python" target="_blank">bkmonitor-ecosystem/examples/metrics/sdks/python</a> 中找到。

PUSH 上报（metric 服务主动上报到端点）：

```bash
docker build -t metrics-sdk-python .

docker run \
-e JOB="default_monitor_job" \
-e INSTANCE="127.0.0.1" \
-e API_URL="{{access_config.custom.sdk}}" \
-e TOKEN="{{access_config.token}}" \
-e INTERVAL=60 metrics-sdk-python
```

### 2.4 使用示例

#### 2.4.1 Counter

用于记录累计值（如 API 调用总量、错误次数），只能递增。 可用于统计接口请求量、错误率（结合 rate / increase 等函数计算）。

例如，可以通过以下方式上报请求总数：

```python
from prometheus_client import Counter

api_counter = Counter(
    "api_called_total",
    "API调用总次数",
    # 定义维度，可用于聚合、过滤。
    ["api_name", "status"],
    registry=registry
)

# 使用 Counter 类型指标
# Refer：https://prometheus.github.io/client_python/instrumenting/counter/
def counter_demo():
    status = "200" if random.random() > 0.1 else "500"
    api_counter.labels(api_name="/user/login", status=status).inc(random.randint(1, 3))
```

<a href="https://prometheus.github.io/client_python/instrumenting/counter/" target="_blank">Prometheus Python SDK - Counter</a>。

#### 2.4.2 Gauge

用于记录瞬时值（可任意增减），如实时资源状态、队列长度、活跃连接数等。

```python
from prometheus_client import Gauge
import random

cpu_gauge = Gauge(
    "cpu_usage_percent",
    "CPU 使用率百分比",
    ["host", "device"],  # 支持多维度标签
    registry=registry
)

# 使用 Gauge 类型指标
# Refer：https://prometheus.github.io/client_python/instrumenting/gauge/
def gauge_demo():
    # 设置主机 host1 的 CPU 使用率（随机值）
    gauge.labels(host="host1", device="cpu0").set(round(random.uniform(0.1, 99.9), 2))

    # 减少 host2 的 CPU 值（模拟负载下降）
    gauge.labels(host="host2", device="cpu0").dec(10)  # 减少 10%
```

<a href="https://prometheus.github.io/client_python/instrumenting/gauge/" target="_blank">Prometheus Python SDK - Gauge</a>。

#### 2.4.3 Histogram

用于记录数值分布情况（如任务耗时、响应大小），通过预定的桶（bucket）统计观测值落入各区间的频率，并自动生成 _sum （总和）、_count （总数）等衍生指标。适用于分析耗时分布、计算分位数（P90/P95）等场景。

```python
from prometheus_client import Histogram
import time
import random

task_histogram = Histogram(
    "task_duration_seconds",
    "任务耗时分布（秒）",
    ["task_type"],  # 标签维度
    buckets=[0.1, 0.5, 1, 2, 5, 10],  # 预定义分桶区间
    registry=registry
)

# 使用 Histogram 类型指标
# Refer：https://prometheus.github.io/client_python/instrumenting/histogram/
def histogram_demo():
    task_type = random.choice(["import", "export", "process"])
    duration = random.uniform(0.05, 12)  # 模拟任务耗时

    # 上下文管理器自动计时
    with task_histogram.labels(task_type=task_type).time():
        time.sleep(duration)  # 模拟任务执行
```

<a href="https://prometheus.github.io/client_python/instrumenting/histogram/" target="_blank">Prometheus Python SDK - Histogram</a>。

#### 2.4.4 Summary

用于在客户端直接计算分位数（如 P95/P99 请求耗时），适用于需高精度分位数且无需跨实例聚合场景。

```python
from prometheus_client import Summary
import random
import time

process_summary = Summary(
    "task_processing_seconds",
    "任务处理时间摘要（秒）",
    ["stage"],  # 标签维度
    registry=registry,
    objectives={
        0.5: 0.05,
        0.95: 0.01,
        0.99: 0.001
    }  # 预定义分位数及误差范围
)

# 使用 Summary 类型指标
# Refer：https://prometheus.github.io/client_python/instrumenting/summary/
def summary_demo():
    stage = random.choice(["validation", "execution", "cleanup"])
    duration = random.uniform(0.01, 15)  # 模拟耗时

    # 记录观测值
    process_summary.labels(stage=stage).observe(duration)
```

<a href="https://prometheus.github.io/client_python/instrumenting/summary/" target="_blank">Prometheus Python SDK - Summary</a>。

### 2.5 样例代码

该样例使用 Prometheus_client 库实现四种指标类型（`Counter`、`Gauge`、`Histogram`、`Summary`）上报：

```python
# -*- coding: utf-8 -*-
import os
import time
import random
import logging
from prometheus_client import CollectorRegistry, Counter, Gauge, Summary, Histogram, push_to_gateway, start_http_server
from prometheus_client.exposition import default_handler

# 基础配置
registry = CollectorRegistry()
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("prom_demo")

# 环境变量配置
TOKEN = os.getenv("TOKEN", "")   # ❗️❗️【非常重要】请填写为申请到的自定义指标认证令牌（`Token`）。
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
            grouping_key={"instance": INSTANCE},
             # ❗️❗️【非常重要】将申请到的自定义指标认证令牌（Token）到 HTTP 请求头（X-BK-TOKEN），该步骤将决定数据能否准确被蓝鲸监控平台接收。
            handler=bk_handler
        )
        logger.info(f"成功推送指标到 {API_URL}")
    except Exception as e:
        logger.error(f"推送失败: {str(e)}")

# ===== 指标类型定义与演示函数 =====

# Counter类型 - API调用统计
# Refer：https://prometheus.github.io/client_python/instrumenting/counter/
api_counter = Counter(
    "api_called_total",
    "API调用总次数",
    ["api_name", "status"],
    registry=registry
)

def counter_demo():
    """模拟API调用计数"""
    status = "200" if random.random() > 0.9 else "500"  # 10%错误率
    api_name = random.choice(["/user/login", "/data/query", "/order/create"])
    api_counter.labels(api_name=api_name, status=status).inc()
    logger.debug(f"记录API调用: {api_name} | 状态: {status}")

# Gauge类型 - CPU使用率监控
# Refer：https://prometheus.github.io/client_python/instrumenting/gauge/
cpu_gauge = Gauge(
    "cpu_usage_percent",
    "CPU使用率百分比",
    ["host"],
    registry=registry
)

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
    buckets=[0.1, 0.5, 1, 2, 5]  # 关键耗时阈值
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
process_summary = Summary(
    "task_processing_seconds",
    "任务处理时间摘要",
    ["stage"],
    registry=registry
)

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
```

### 2.5 PULL 模式

上文主要介绍将指标数据，**主动推送**到蓝鲸监控平台，也可以通过 HTTP 暴露指标，通过 ServiceMonitor（BCS）或采集插件的方式拉取。

样例代码同时兼容 PULL 和 PUSH，通过 `start_http_server` 在给定端口上的守护进程线程中启动 HTTP 服务器，暴露指标：

```python
from prometheus_client import start_http_server

METRICS_PORT = int(os.getenv("METRICS_PORT", "2323"))  # 默认2323端口暴露/metrics端点

def main():
    start_http_server(METRICS_PORT)
    logger.info(f"已启用Pull模式 | 指标端点: http://127.0.0.1:{METRICS_PORT}/metrics")
    ...
```

运行样例：

```bash
docker build -t metrics-sdk-python .
docker run -p 2323:2323 --name sdk-pull-python metrics-sdk-python
```

获取指标：

```bash
curl http://127.0.0.1:2323/metrics
```

得到类似输出说明启动成功：

```python
# HELP python_gc_objects_collected_total Objects collected during gc
# TYPE python_gc_objects_collected_total counter
python_gc_objects_collected_total{generation="0"} 362.0
python_gc_objects_collected_total{generation="1"} 0.0
python_gc_objects_collected_total{generation="2"} 0.0
```

## 3. 了解更多

* 进行 <a href="{{docs.metrics.learn.Index_search}}" target="_blank">指标检索</a>。

* 了解 <a href="{{docs.metrics.learn.Use_indicators}}" target="_blank">怎么使用监控指标</a>。

* 了解如何 <a href="{{docs.metrics.learn.configure_dashboard}}" target="_blank">配置仪表盘</a>。

* 了解如何使用 <a href="{{docs.metrics.learn.alarms}}" target="_blank">监控告警</a>。

* 了解 <a href="https://prometheus.github.io/client_python/" target="_blank"> Prometheus Python SDK</a>。

* 了解 <a href="https://prometheus.github.io/client_python/exporting/" target="_blank">Promethues Python SDK 指标导出方式 </a>。
