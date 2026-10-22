# Java-日志（SDK）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/sdks/README.md" target="_blank">自定义日志 OpenTelemetry SDK 上报</a>

* <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs（OTel 日志）</a>：OpenTelemetry 中用于描述离散日志事件的信号类型。

* <a href="https://opentelemetry.io/docs/specs/otel/logs/data-model/" target="_blank">Logs Data Model（OTel 日志数据模型）</a>：定义 `resourceLogs`、`scopeLogs`、`logRecords`、`body`、`attributes`、`severityNumber` 等字段含义。

* <a href="https://opentelemetry.io/docs/specs/otlp/#otlphttp" target="_blank">OTLP/HTTP（OpenTelemetry HTTP 上报协议）</a>：定义通过 HTTP 上报 OTel 数据的协议方式，Java SDK 通过 `/v1/logs` 上报日志。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/logs/sdks/java
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/sdks/README.md" target="_blank">自定义日志 OpenTelemetry SDK 上报</a> 创建一个上报协议为 `JSON` 的自定义日志，关注创建后提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。

* `API_URL`：国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，其他环境、跨云场景请根据页面接入指引填写。

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `API_URL` | 是 | ❗❗【非常重要】日志上报接口地址（`Access URL`），请根据页面接入指引填写；如果页面提供的是 OTLP HTTP Endpoint 根地址，请在末尾追加 `/v1/logs`。 |
| `TOKEN` | 是 | ❗❗【非常重要】日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。 |

#### 2.2.1 关键配置

使用 OpenTelemetry Java SDK 进行日志上报时，需正确配置以下核心参数：

```java
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.exporter.otlp.http.logs.OtlpHttpLogRecordExporter;
import io.opentelemetry.sdk.logs.SdkLoggerProvider;
import io.opentelemetry.sdk.logs.export.BatchLogRecordProcessor;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.semconv.ResourceAttributes;

// ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
String token = System.getenv().getOrDefault("TOKEN", "fixme");
// ❗❗【非常重要】上报地址，国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，
// 其他环境、跨云场景请根据页面接入指引填写
String apiUrl = System.getenv().getOrDefault("API_URL", "http://127.0.0.1:4318/v1/logs");

// 创建 OTLP HTTP Log Exporter，注入认证 Header
OtlpHttpLogRecordExporter exporter = OtlpHttpLogRecordExporter.builder()
        .setEndpoint(apiUrl)
        .addHeader("x-bk-token", token)
        .build();

// 创建 LoggerProvider，关联资源属性（服务名、环境等）
SdkLoggerProvider loggerProvider = SdkLoggerProvider.builder()
        .setResource(Resource.getDefault().merge(
                Resource.create(Attributes.of(
                        ResourceAttributes.SERVICE_NAME, "custom-log-sdk-demo",
                        ResourceAttributes.DEPLOYMENT_ENVIRONMENT, "local"
                ))
        ))
        .addLogRecordProcessor(BatchLogRecordProcessor.builder(exporter).build())
        .build();

// 获取 OpenTelemetry Logger
io.opentelemetry.api.logs.Logger otelLogger = loggerProvider.get("otel");
```

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/master/examples/logs/sdks/java" target="_blank">bkmonitor-ecosystem/examples/logs/sdks/java</a> 中找到。

通过 docker build 构建名为 logs-sdk-java 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、API_URL 传递配置参数，实现周期上报日志：

```bash
docker build -t logs-sdk-java .

docker run -e TOKEN="fixme" \
 -e API_URL="http://127.0.0.1:4318/v1/logs" \
 logs-sdk-java
```

运行输出：

```bash
Picked up JAVA_TOOL_OPTIONS: -Dfile.encoding=UTF-8 -Dsun.stdout.encoding=UTF-8 -Dsun.stderr.encoding=UTF-8
Oct 21, 2026 7:39:05 PM com.tencent.bkm.demo.Main main
INFO: Starting log reporter (press Ctrl+C to stop)...
Oct 21, 2026 7:39:05 PM com.tencent.bkm.demo.Main main
SEVERE: error log from java sdk
Oct 21, 2026 7:39:06 PM com.tencent.bkm.demo.Main main
INFO: info log from java sdk
Oct 21, 2026 7:39:06 PM com.tencent.bkm.demo.Main main
WARNING: warn log from java sdk
Oct 21, 2026 7:39:06 PM com.tencent.bkm.demo.Main main
INFO: info log from java sdk
...
```

### 2.4 样例代码

该样例使用 OpenTelemetry Java SDK，通过 OTLP/HTTP 协议上报日志。代码集成 `opentelemetry-sdk-logs` 和 `opentelemetry-exporter-otlp`，通过 `SdkLoggerProvider` 对 OTel LogRecord 进行构造与批量上报。

```java
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
```

## 3. 了解更多

进一步了解以下内容：

* 进行 <a href="#" target="_blank">日志检索</a>。

* 了解 <a href="#" target="_blank">容器日志自定义上报使用文档</a>。

* 了解 <a href="#" target="_blank">容器日志采集器安装</a>。