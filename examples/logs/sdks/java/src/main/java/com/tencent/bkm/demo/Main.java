// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo;

import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.common.AttributeKey;
import io.opentelemetry.api.logs.Severity;
import io.opentelemetry.exporter.otlp.http.logs.OtlpHttpLogRecordExporter;
import io.opentelemetry.sdk.logs.SdkLoggerProvider;
import io.opentelemetry.sdk.logs.export.BatchLogRecordProcessor;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.semconv.ResourceAttributes;

import java.time.Instant;
import java.util.Random;
import java.util.concurrent.TimeUnit;
import java.util.logging.Logger;


public class Main {

    private static final Logger consoleLogger = Logger.getLogger("otel");

    private static final LogLevel[] LOG_LEVELS = {
            new LogLevel(Severity.DEBUG, "DEBUG", "debug log from java sdk"),
            new LogLevel(Severity.INFO, "INFO", "info log from java sdk"),
            new LogLevel(Severity.WARN, "WARN", "warn log from java sdk"),
            new LogLevel(Severity.ERROR, "ERROR", "error log from java sdk")
    };

    private static final Random RANDOM = new Random();

    public static void main(String[] args) {
        // ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
        String token = System.getenv().getOrDefault("TOKEN", "fixme");
        /* ❗❗【非常重要】上报地址，国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，
         * 其他环境、跨云场景请根据页面接入指引填写
         */
        String apiUrl = System.getenv().getOrDefault("API_URL", "http://127.0.0.1:4318/v1/logs");

        OtlpHttpLogRecordExporter exporter = OtlpHttpLogRecordExporter.builder()
                .setEndpoint(apiUrl)
                .addHeader("x-bk-token", token)
                .build();

        Resource resource = Resource.getDefault().merge(
                Resource.create(Attributes.of(
                        ResourceAttributes.SERVICE_NAME, "custom-log-sdk-demo",
                        ResourceAttributes.DEPLOYMENT_ENVIRONMENT, "local"
                ))
        );

        SdkLoggerProvider loggerProvider = SdkLoggerProvider.builder()
                .setResource(resource)
                .addLogRecordProcessor(BatchLogRecordProcessor.builder(exporter).build())
                .build();

        Runtime.getRuntime().addShutdownHook(new Thread(loggerProvider::shutdown));

        io.opentelemetry.api.logs.Logger otelLogger = loggerProvider.get("otel");

        consoleLogger.info("Starting log reporter (press Ctrl+C to stop)...");

        try {
            while (true) {
                LogLevel level = LOG_LEVELS[RANDOM.nextInt(LOG_LEVELS.length)];

                long epochNanos = Instant.now().toEpochMilli() * 1_000_000L;
                otelLogger.logRecordBuilder()
                        .setTimestamp(epochNanos, TimeUnit.NANOSECONDS)
                        .setObservedTimestamp(epochNanos, TimeUnit.NANOSECONDS)
                        .setSeverity(level.severity)
                        .setSeverityText(level.severityText)
                        .setBody(level.message)
                        .setAllAttributes(Attributes.of(
                                AttributeKey.stringKey("demo.source"), "java"
                        ))
                        .emit();

                switch (level.severity) {
                    case DEBUG:
                        consoleLogger.fine(level.message);
                        break;
                    case INFO:
                        consoleLogger.info(level.message);
                        break;
                    case WARN:
                        consoleLogger.warning(level.message);
                        break;
                    case ERROR:
                        consoleLogger.severe(level.message);
                        break;
                    default:
                        consoleLogger.info(level.message);
                }

                TimeUnit.MILLISECONDS.sleep(100);
            }
        } catch (InterruptedException e) {
            consoleLogger.info("Received keyboard interrupt, exiting...");
            Thread.currentThread().interrupt();
        } finally {
            loggerProvider.shutdown();
        }
    }

    private static class LogLevel {
        final Severity severity;
        final String severityText;
        final String message;

        LogLevel(Severity severity, String severityText, String message) {
            this.severity = severity;
            this.severityText = severityText;
            this.message = message;
        }
    }
}
