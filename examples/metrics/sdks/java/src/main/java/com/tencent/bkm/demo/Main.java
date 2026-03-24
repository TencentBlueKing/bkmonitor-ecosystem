// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

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

/**
 * Prometheus 指标上报示例主类.
 */
public class Main {

    private static final Logger logger = Logger.getLogger(Main.class.getName());

    // 环境变量配置
    // ❗️❗️【非常重要】请填写为申请到的自定义指标认证令牌（Token）
    private static final String TOKEN = System.getenv("TOKEN");
    // ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
    private static final String API_URL = System.getenv("API_URL")
            .replaceFirst("^https?://", "");
    private static final String JOB = System.getenv("JOB") != null
            ? System.getenv("JOB") : "default_monitor_job";
    private static final String INSTANCE = System.getenv("INSTANCE") != null
            ? System.getenv("INSTANCE") : "127.0.0.1";
    private static final int INTERVAL = System.getenv("INTERVAL") != null
            ? Integer.parseInt(System.getenv("INTERVAL")) : 60;
    private static final int METRICS_PORT = System.getenv("METRICS_PORT") != null
            ? Integer.parseInt(System.getenv("METRICS_PORT")) : 2323;

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
                .groupingKey("instance", INSTANCE)
                .groupingKey("language", "java");

        // ❗️❗️【非常重要】注入 TOKEN
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

    /**
     * 主函数入口.
     *
     * @param args 命令行参数
     * @throws Exception 异常
     */
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
        logger.info("已启用 Pull 模式 | 指标端点: http://127.0.0.1:"
                + METRICS_PORT + "/metrics");
        logger.info("启动指标上报服务 | 实例: " + INSTANCE + " | 任务: " + JOB);
        logger.info("目标地址: " + API_URL + " | 认证令牌: "
                + (TOKEN != null && !TOKEN.isEmpty() ? "已配置" : "未配置"));
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
            // 至少间隔1秒
            long sleepTime = Math.max(INTERVAL * 1000 - elapsed, 1000);
            Thread.sleep(sleepTime);
        }
    }
}
