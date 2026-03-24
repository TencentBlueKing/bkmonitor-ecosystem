// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.metric;

import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.AttributeKey;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.metrics.DoubleHistogram;
import io.opentelemetry.api.metrics.LongCounter;
import io.opentelemetry.api.metrics.Meter;

import java.util.List;
import java.util.Random;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

/**
 * 指标上报示例主类.
 */
public class Main {

    /**
     * 执行任务.
     */
    private static void doSomething() {
        Random random = new Random();
        int sleepTime = 10 + random.nextInt(100);
        try {
            Thread.sleep(sleepTime);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    /**
     * 主函数入口.
     *
     * @param args 命令行参数
     */
    public static void main(String[] args) {
        OpenTelemetry openTelemetry = OtlpConfiguration.getInstance().getOpenTelemetry();

        // 当 OpenTelemetry SDK 初始化完成后，也可以通过 GlobalOpenTelemetry 获取 Meter
        // import io.opentelemetry.api.GlobalOpenTelemetry;
        // GlobalOpenTelemetry.getMeter("com.tencent.bkm.demo.metric");
        Meter meter = openTelemetry.getMeter("com.tencent.bkm.demo.metric");

        // Metrics（指标）- 使用 Counter 类型指标
        // Refer: https://opentelemetry.io/docs/languages/java/instrumentation/#using-counters
        LongCounter requestsTotal = meter.counterBuilder("requests_total")
                .setDescription("Total number of HTTP requests")
                .setUnit("requests")
                .build();

        // Metrics（指标）- 使用 Histogram 类型指标
        // Refer:
        // https://opentelemetry.io/docs/languages/java/instrumentation/#using-histograms
        DoubleHistogram taskExecuteDurationSeconds = meter
                .histogramBuilder("task_execute_duration_seconds")
                .setDescription("Task execute duration in seconds")
                .setExplicitBucketBoundariesAdvice(
                        List.of(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0))
                .setUnit("seconds")
                .build();

        ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(1);
        scheduler.scheduleAtFixedRate(
                () -> {
                    // 打印日志（可读时间 + 运行日志）
                    System.out.println("[" + System.currentTimeMillis() + "] Executing task...");

                    // 模拟请求计数
                    requestsTotal.add(1, Attributes.of(
                            AttributeKey.stringKey("k1"), "v1",
                            AttributeKey.stringKey("k2"), "v2"
                    ));

                    // 模拟任务执行时间
                    long begin = System.nanoTime();
                    doSomething();
                    long end = System.nanoTime();
                    double durationInSeconds = (end - begin) / 1_000_000_000.0;
                    taskExecuteDurationSeconds.record(durationInSeconds,
                            Attributes.of(
                                    AttributeKey.stringKey("task"), "exampleTask",
                                    AttributeKey.stringKey("status"), "success",
                                    AttributeKey.stringKey("k1"), "v1",
                                    AttributeKey.stringKey("k2"), "v2"

                            ));
                }, 0, 5, TimeUnit.SECONDS);

        Runtime.getRuntime().addShutdownHook(new Thread(scheduler::shutdown));
    }
}