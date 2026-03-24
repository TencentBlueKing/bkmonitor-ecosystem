# Java-指标（Prometheus）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Term/metrics/what.md" target="_blank">什么是指标</a>

* <a href="{{COOKBOOK_METRICS_TYPES}}" target="_blank">指标类型</a>

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/metrics/sdks/java
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Quickstarts/metrics/sdks/README.md" target="_blank">自定义指标 Prometheus SDK 上报</a> 创建一个上报协议为 `Prometheus` 的自定义指标，关注创建后提供的配置项：

* `TOKEN`：数据源 Token，后续需要在上报数据时使用。

同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image-1.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数             | 类型         | 描述                                                                                                 |
|----------------|------------|----------------------------------------------------------------------------------------------------|
| `TOKEN`       | String     | ❗❗【非常重要】 自定义指标数据源 `Token`。                                                                               |
| `API_URL`      | String     | ❗❗【非常重要】 数据上报接口地址（`Access URL`），国内站点请填写「 127.0.0.1:4318 」，其他环境、跨云场景请根据页面接入指引填写。 |
| `INTERVAL`     | Integer    | 数据上报间隔，默认值为 60 秒。                                                             |
| `METRICS_PORT` | Integer  　| 指标暴露端口，默认 2323。                                                                                    |

#### 2.2.1 增加依赖

需要使用到下面这些依赖，其中 prometheus-metrics-instrumentation-jvm 用于暴露 JVM 相关指标，prometheus-metrics-exporter-httpserver 用于 pull 模式上报指标，prometheus-metrics-exporter-pushgateway 用于 push 模式上报指标。

```groovy
implementation 'io.prometheus:prometheus-metrics-core:1.3.9'
implementation 'io.prometheus:prometheus-metrics-instrumentation-jvm:1.3.9'
implementation 'io.prometheus:prometheus-metrics-exporter-httpserver:1.3.9'
implementation 'io.prometheus:prometheus-metrics-exporter-pushgateway:1.3.9'
```

#### 2.2.2 关键配置

蓝鲸监控支持原生 Prometheus 协议，如果业务已接入 Prometheus SDK，只需在 `push_to_gateway` 方法，修改上报地址为 `API_URL`，注入 `TOKEN` 即可。

由于 prometheus 的 client_java 不支持加入自定义 headers，所以这里使用 basicAuth 进行验证, user:bkmonitor, password:$TOKEN。

```java

import io.prometheus.metrics.exporter.pushgateway.PushGateway;
// ===== 推送指标相关 =====
// 通过 PushGateway 推送指标
// Refer：https://prometheus.github.io/client_java/exporters/pushgateway/
private static void safePushMetrics() {

    // 创建PushGateway实例
    PushGateway.Builder builder = PushGateway.builder()
            .address(API_URL)
            .job(JOB)
            .groupingKey("instance", INSTANCE);

    // ❗️❗️【非常重要】注入 `TOKEN`。
    if (TOKEN != null && !TOKEN.isEmpty()) {
        builder.basicAuth("bkmonitor", TOKEN);
    }
    PushGateway pushGateway = builder.build();
    try {
        // 推送指标
        pushGateway.push();
        logger.info("成功推送指标到 " + API_URL);
    } catch (Exception e) {
        logger.severe("推送失败: " + e.getMessage());
    }
}
```

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/metrics/sdks/java" target="_blank">bkmonitor-ecosystem/examples/metrics/sdks/java</a> 中找到。

PUSH 上报（metric 服务主动上报到端点）：

```bash
docker build -t metrics-sdk-java .

docker run \
-e JOB="default_monitor_job" \
-e INSTANCE="127.0.0.1" \
-e API_URL="127.0.0.1:4318" \
-e TOKEN="xxx" \
-e INTERVAL=60  metrics-sdk-java
```

PULL 上报（通过接口获取自定义的 metric 信息）：

```bash
docker build -t metrics-sdk-java .

docker run -p 2323:2323 --name sdk-pull-java metrics-sdk-java
```

### 2.4 使用示例

#### 2.4.1 Counter

用于记录累计值（如 API 调用总量、错误次数），只能递增。 可用于统计接口请求量、错误率（结合 rate / increase 等函数计算）。

例如，可以通过以下方式上报请求总数：

```java
import io.prometheus.metrics.core.metrics.Counter;

// ===== 计数器相关 =====
// Metrics（指标）- 使用 Counter 类型指标
// Refer: https://prometheus.github.io/client_java/getting-started/metric-types/#counter
private static final Counter requestsTotal = Counter.builder()
        .name("requests_total")
        .help("Total number of HTTP requests")
        .labelNames("k1", "k2")
        register();

private static void simulateRequestCount() {
    requestsTotal.labelValues("v1", "v2").inc();
}
```

<a href="https://prometheus.github.io/client_java/getting-started/metric-types/#counter" target="_blank">Prometheus client_java - Counter</a>。

#### 2.4.2 Gauge

用于记录瞬时值（可任意增减），如实时资源状态、队列长度、活跃连接数等。

```java
// ===== 仪表盘相关 =====
// Metrics（指标）- 使用 Gauge 类型指标
// Refer: https://prometheus.github.io/client_java/getting-started/metric-types/#gauge
private static final Gauge activeRequests = Gauge.builder()
        .name("active_requests")
        .help("Current number of active HTTP requests")
        .labelNames("api_endpoint")
        .register();

private static void simulateActiveRequests() {
    activeRequests.labelValues("/api/v1/users").inc();
    // 模拟程序执行耗时
    randomDelaySomeTime();
    activeRequests.labelValues("/api/v1/users").dec();

}
```

<a href="https://prometheus.github.io/client_java/getting-started/metric-types/#gauge" target="_blank">Prometheus client_java - Gauge</a>。

#### 2.4.3 Histogram

用于记录数值分布情况（如任务耗时、响应大小），通过预定的桶（bucket）统计观测值落入各区间的频率，并自动生成 _sum （总和）、_count （总数）等衍生指标。适用于分析耗时分布、计算分位数（P90/P95）等场景。

```java
// ===== 直方图相关 =====
// Metrics（指标）- 使用 Histogram 类型指标
// Refer: https://prometheus.github.io/client_java/getting-started/metric-types/#histogram
private static final Histogram taskDuration = Histogram.builder()
        .name("task_execute_duration_seconds")
        .help("Task execute duration in seconds")
        .classicUpperBounds(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0)
        .labelNames("task", "status", "k1", "k2")
        .register();

private static void simulateTaskDuration() {
    long start = System.nanoTime();
    // 模拟程序执行耗时
    randomDelaySomeTime();
    taskDuration.labelValues("GET", "/", "200", "v1")
            .observe(Unit.nanosToSeconds(System.nanoTime() - start));

}
```

<a href="https://prometheus.github.io/client_java/getting-started/metric-types/#histogram" target="_blank">Prometheus client_java - Histogram</a>。

#### 2.4.3 Summary

用于在客户端直接计算分位数（如 P95/P99 请求耗时），适用于需高精度分位数且无需跨实例聚合场景。

```java
// ===== 摘要相关 =====
// Metrics（指标）- 使用 Summary 类型指标
// Refer: https://prometheus.github.io/client_java/getting-started/metric-types/#summary
private static final Summary requestLatency = Summary.builder()
                .name("http_request_latency_seconds")
                .help("HTTP request latency distribution")
                .quantile(0.5, 0.01)
                .quantile(0.9, 0.005)
                .labelNames("method", "path")
                .register();

private static void simulateRequestLatency() {
    long start = System.nanoTime();
    // 模拟程序执行耗时
    randomDelaySomeTime();
    requestLatency.labelValues("GET", "/users")
            .observe(Unit.nanosToSeconds(System.nanoTime() - start));

}
```

<a href="https://prometheus.github.io/client_java/getting-started/metric-types/#summary" target="_blank">Prometheus client_java - Summary</a>。

#### 2.4.5 JvmMetrics

用于自动收集 JVM 相关的指标数据，包括内存使用、GC 情况、线程状态等。这是 Prometheus Java 客户端提供的标准 JVM 监控指标。

```java
import io.prometheus.metrics.instrumentation.jvm.JvmMetrics;

private static void javaMetrics() {
    // 初始化JVM内置指标
    JvmMetrics.builder().register();
}
```

### 2.5 样例代码

该样例使用 Prometheus_client 库实现四种指标类型（`Counter`、`Gauge`、`Histogram`、`Summary`）上报：

```java
package com.tencent.bkm.demo;

import io.prometheus.metrics.core.metrics.Counter;
import io.prometheus.metrics.core.metrics.Histogram;
import io.prometheus.metrics.core.metrics.Gauge;
import io.prometheus.metrics.core.metrics.Summary;
import io.prometheus.metrics.exporter.httpserver.HTTPServer;
import io.prometheus.metrics.exporter.pushgateway.PushGateway;


import io.prometheus.metrics.instrumentation.jvm.JvmMetrics;
import io.prometheus.metrics.model.snapshots.Unit;
import java.util.logging.Logger;


public class Main {

    private static final Logger logger = Logger.getLogger(Main.class.getName());

    // 环境变量配置
    private static final String TOKEN = System.getenv("TOKEN");        // ❗️❗️【非常重要】请填写为申请到的自定义指标认证令牌（`Token`）。
    private static final String API_URL = System.getenv("API_URL");        // ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写。
    private static final String JOB = System.getenv("JOB") != null ? System.getenv("JOB") : "default_monitor_job";
    private static final String INSTANCE = System.getenv("INSTANCE") != null ? System.getenv("INSTANCE") : "127.0.0.1";
    private static final int INTERVAL = System.getenv("INTERVAL") != null ? Integer.parseInt(System.getenv("INTERVAL")) : 60;
    private static final int METRICS_PORT = System.getenv("METRICS_PORT") != null ? Integer.parseInt(System.getenv("METRICS_PORT")) : 9400;

    private static void randomDelaySomeTime() {
        try {
            Thread.sleep(10 + (long) (Math.random() * 100));
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    // ===== 计数器相关 =====
    // Metrics（指标）- 使用 Counter 类型指标
    // Refer: https://prometheus.github.io/client_java/getting-started/metric-types/#counter
    private static final Counter requestsTotal = Counter.builder()
            .name("requests_total")
            .help("Total number of HTTP requests")
            .labelNames("k1", "k2")
            .register();

    private static void simulateRequestCount() {
        requestsTotal.labelValues("v1", "v2").inc();
    }

    // ===== 仪表盘相关 =====
    // Metrics（指标）- 使用 Gauge 类型指标
    // Refer: https://prometheus.github.io/client_java/getting-started/metric-types/#gauge
    private static final Gauge activeRequests = Gauge.builder()
            .name("active_requests")
            .help("Current number of active HTTP requests")
            .labelNames("api_endpoint")
            .register();

    private static void simulateActiveRequests() {
        activeRequests.labelValues("/api/v1/users").inc();
        // 模拟程序执行耗时
        randomDelaySomeTime();
        activeRequests.labelValues("/api/v1/users").dec();

    }

    // ===== 直方图相关 =====
    // Metrics（指标）- 使用 Histogram 类型指标
    // Refer: https://prometheus.github.io/client_java/getting-started/metric-types/#histogram
    private static final Histogram taskDuration = Histogram.builder()
            .name("task_execute_duration_seconds")
            .help("Task execute duration in seconds")
            .classicUpperBounds(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0)
            .labelNames("task", "status", "k1", "k2")
            .register();

    private static void simulateTaskDuration() {
        long start = System.nanoTime();
        // 模拟程序执行耗时
        randomDelaySomeTime();
        taskDuration.labelValues("GET", "/", "200", "v1")
                .observe(Unit.nanosToSeconds(System.nanoTime() - start));

    }

    // ===== 摘要相关 =====
    // Metrics（指标）- 使用 Summary 类型指标
    // Refer: https://prometheus.github.io/client_java/getting-started/metric-types/#summary
    private static final Summary requestLatency = Summary.builder()
            .name("http_request_latency_seconds")
            .help("HTTP request latency distribution")
            .quantile(0.5, 0.01)
            .quantile(0.9, 0.005)
            .labelNames("method", "path")
            .register();

    private static void simulateRequestLatency() {
        long start = System.nanoTime();
        // 模拟程序执行耗时
        randomDelaySomeTime();
        requestLatency.labelValues("GET", "/users")
                .observe(Unit.nanosToSeconds(System.nanoTime() - start));

    }

    // ===== 推送指标相关 =====
    // 通过 PushGateway 推送指标
    // Refer：https://prometheus.github.io/client_java/exporters/pushgateway/
    private static void safePushMetrics() {

        // 创建PushGateway实例
        PushGateway.Builder builder = PushGateway.builder()
                .address(API_URL)
                .job(JOB)
                .groupingKey("instance", INSTANCE);

        // ❗️❗️【非常重要】注入 `TOKEN`。
        if (TOKEN != null && !TOKEN.isEmpty()) {
            builder.basicAuth("bkmonitor", TOKEN);
        }
        PushGateway pushGateway = builder.build();
        try {
            // 推送指标
            pushGateway.push();
            logger.info("成功推送指标到 " + API_URL);
        } catch (Exception e) {
            logger.severe("推送失败: " + e.getMessage());
        }
    }

    public static void main(String[] args) throws Exception {
        // 检查必要环境变量
        if (API_URL == null || API_URL.isEmpty()) {
            throw new IllegalArgumentException("API_URL 环境变量必须设置");
        }

        // 初始化JVM内置指标
        JvmMetrics.builder().register();

        // 同时启动HTTP服务器（可选）
        HTTPServer server = HTTPServer.builder()
                .port(METRICS_PORT)
                .buildAndStart();

        // 主执行函数 - 同时支持Pull模式与Push模式
        logger.info("主执行函数 - 同时支持Pull模式与Push模式");
        logger.info("已启用Pull模式 | 指标端点: http://127.0.0.1:" + METRICS_PORT + "/metrics");
        logger.info("启动指标上报服务 | 实例: " + INSTANCE + " | 任务: " + JOB);
        logger.info("目标地址: " + API_URL + " | 认证令牌: " + (TOKEN != null && !TOKEN.isEmpty() ? "已配置" : "未配置"));
        logger.info("上报间隔: " + INTERVAL + "秒");

        // 模拟指标更新并推送
        while (true) {
            long startTime = System.currentTimeMillis();

            // 更新指标
            simulateRequestCount();
            simulateActiveRequests();
            simulateTaskDuration();
            simulateRequestLatency();

            // 推送指标
            safePushMetrics();

            // 计算并等待下次推送
            long elapsed = System.currentTimeMillis() - startTime;
            long sleepTime = Math.max(INTERVAL * 1000 - elapsed, 1000); // 至少间隔1秒
            Thread.sleep(sleepTime);
        }
    }
}
```

### 2.6 Pull 模式

上文主要介绍将指标数据，**主动推送**到蓝鲸监控平台，也可以通过 HTTP 暴露指标，通过 ServiceMonitor（BCS）或采集插件的方式拉取。

样例代码同时兼容 PULL 和 PUSH，通过 `HTTPServer.builder` 在给定端口上的守护进程线程中启动 HTTP 服务器，暴露指标：

```java
import io.prometheus.metrics.exporter.httpserver.HTTPServer;

private static final int METRICS_PORT = System.getenv("METRICS_PORT") != null ? Integer.parseInt(System.getenv("METRICS_PORT")) : 9400;  // 默认 9400 端口暴露 /metrics 端点

public static void main(String[] args) throws Exception {

    // 初始化JVM内置指标，可选，也可以上报自定义指标
    JvmMetrics.builder().register();

    // 启动HTTP服务器
    HTTPServer server = HTTPServer.builder()
            .port(METRICS_PORT)
            .buildAndStart();

    logger.info("已启用Pull模式 | 指标端点: http://127.0.0.1:" + METRICS_PORT + "/metrics");
}
```

运行样例：

```bash
docker build -t metrics-sdk-java .

docker run -d -p 2323:2323 --name sdk-pull-java metrics-sdk-java
```

获取指标：

```bash
curl http://127.0.0.1:2323/metrics
```

得到类似输出说明启动成功：

```bash
# HELP jvm_buffer_pool_capacity_bytes Bytes capacity of a given JVM buffer pool.
# TYPE jvm_buffer_pool_capacity_bytes gauge
jvm_buffer_pool_capacity_bytes{pool="direct"} 16384.0
jvm_buffer_pool_capacity_bytes{pool="mapped"} 0.0
# HELP jvm_buffer_pool_used_buffers Used buffers of a given JVM buffer pool.
# TYPE jvm_buffer_pool_used_buffers gauge
jvm_buffer_pool_used_buffers{pool="direct"} 2.0
jvm_buffer_pool_used_buffers{pool="mapped"} 0.0
# HELP jvm_buffer_pool_used_bytes Used bytes of a given JVM buffer pool.
# TYPE jvm_buffer_pool_used_bytes gauge
jvm_buffer_pool_used_bytes{pool="direct"} 16384.0
jvm_buffer_pool_used_bytes{pool="mapped"} 0.0
```

## 3. 了解更多

* 进行 <a href="#" target="_blank">指标检索</a>。

* 了解 <a href="#" target="_blank">怎么使用监控指标</a>。

* 了解如何 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/dashboard.md" target="_blank">配置仪表盘</a>。

* 了解如何使用 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/alarm-configurations/rules.md" target="_blank">监控告警</a>。