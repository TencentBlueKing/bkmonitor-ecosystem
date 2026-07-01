# Java-日志（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="{{docs.logs.http.readme.http_logs_report}}" target="_blank">自定义日志 HTTP 上报</a>

* <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs（OTel 日志）</a>：OpenTelemetry 中用于描述离散日志事件的信号类型。

* <a href="https://opentelemetry.io/docs/specs/otel/logs/data-model/" target="_blank">Logs Data Model（OTel 日志数据模型）</a>：定义 `resourceLogs`、`scopeLogs`、`logRecords`、`body`、`attributes`、`severityNumber` 等字段含义。

* <a href="https://opentelemetry.io/docs/specs/otlp/#otlphttp" target="_blank">OTLP/HTTP（OpenTelemetry HTTP 上报协议）</a>：定义通过 HTTP 上报 OTel 数据的协议方式，本示例使用 `/v1/logs` 上报日志。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/logs/http/java
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="{{docs.logs.http.readme.http_logs_report}}" target="_blank">自定义日志 HTTP 上报</a> 创建一个上报协议为 `JSON` 的自定义日志，关注创建后提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。

* `API_URL`：国内站点默认是「 {{access_config.otlp.http_endpoint}}/v1/logs 」，其他环境、跨云场景请根据页面接入指引填写。

### 2.2 样例运行参数

运行参数说明：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `API_URL` | 是 | ❗❗【非常重要】日志上报接口地址（`Access URL`），请根据页面接入指引填写；如果页面提供的是 OTLP HTTP Endpoint 根地址，请在末尾追加 `/v1/logs`。 |
| `TOKEN` | 是 | ❗❗【非常重要】日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。 |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/logs/http/java" target="_blank">bkmonitor-ecosystem/examples/logs/http/java</a> 中找到。

通过 docker build 构建名为 logs-http-java 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、API_URL 传递配置参数，实现周期上报日志：

```bash
docker build -t logs-http-java .

docker run -e TOKEN="fixme" \
 -e API_URL="{{access_config.otlp.http_endpoint}}/v1/logs" \
 logs-http-java
```

运行输出：

```bash
> Task :run
Picked up JAVA_TOOL_OPTIONS: -Dfile.encoding=UTF-8 -Dsun.stdout.encoding=UTF-8 -Dsun.stderr.encoding=UTF-8
2026-06-30 19:54:56,712 - INFO - Starting log reporter (press Ctrl+C to stop)...
2026-06-30 19:54:56,917 - INFO - Sending log level: ERROR (17)
2026-06-30 19:54:57,034 - INFO - response.status_code=200, body={}
2026-06-30 19:54:57,135 - INFO - Sending log level: WARN (13)
2026-06-30 19:54:57,309 - INFO - response.status_code=200, body={}
2026-06-30 19:54:57,410 - INFO - Sending log level: DEBUG (5)
2026-06-30 19:54:57,440 - INFO - response.status_code=200, body={}
...
```

### 2.4 样例代码

上报代码示例：

```java
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonArray;
import com.google.gson.JsonObject;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.Random;
import java.util.logging.Level;
import java.util.logging.Logger;

public class Main {

    private static final Logger LOGGER = Logger.getLogger(Main.class.getName());
    private static final Random RANDOM = new Random();
    private static final Gson GSON = new GsonBuilder().create();

    // 日志级别映射
    private static final LogLevel[] LOG_LEVELS = {
        new LogLevel(5, "DEBUG", "debug log from python http"),
        new LogLevel(9, "INFO", "info log from python http"),
        new LogLevel(13, "WARN", "warn log from python http"),
        new LogLevel(17, "ERROR", "error log from python http")
    };

    private static class LogLevel {
        int severityNumber;
        String severityText;
        String message;

        LogLevel(int severityNumber, String severityText, String message) {
            this.severityNumber = severityNumber;
            this.severityText = severityText;
            this.message = message;
        }
    }

    public static void main(String[] args) throws InterruptedException {
        LOGGER.setLevel(Level.INFO);
        System.setProperty("java.util.logging.SimpleFormatter.format",
                "%1$tF %1$tT,%1$tL - %4$s - %5$s%n");

        LOGGER.info("Starting log reporter (press Ctrl+C to stop)...");

        // ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
        String token = System.getenv().getOrDefault("TOKEN", "fixme");
        /** ❗❗【非常重要】上报地址，国内站点默认是「{{access_config.otlp.http_endpoint}}/v1/logs」，
        * 其他环境、跨云场景请根据页面接入指引填写
        */
        String apiUrl = System.getenv().getOrDefault("API_URL", "{{access_config.otlp.http_endpoint}}/v1/logs");

        HttpClient client = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();

        try {
            while (true) {
                JsonObject payload = buildPayload();
                doPost(client, apiUrl, token, payload);
                Thread.sleep(100); // 0.1秒
            }
        } catch (InterruptedException e) {
            LOGGER.info("Received keyboard interrupt, exiting...");
            Thread.currentThread().interrupt();
        }
    }

    private static String getCurrentNanoTimestamp() {
        return String.valueOf(System.currentTimeMillis() * 1_000_000L +
                (System.nanoTime() % 1_000_000L));
    }

    private static LogLevel getRandomLevel() {
        return LOG_LEVELS[RANDOM.nextInt(LOG_LEVELS.length)];
    }

    private static JsonObject buildPayload() {
        String currentNano = getCurrentNanoTimestamp();
        LogLevel level = getRandomLevel();

        JsonObject resourceAttributes = new JsonObject();
        resourceAttributes.addProperty("key", "service.name");
        JsonObject serviceNameValue = new JsonObject();
        serviceNameValue.addProperty("stringValue", "custom-log-demo");
        resourceAttributes.add("value", serviceNameValue);

        JsonObject envAttribute = new JsonObject();
        envAttribute.addProperty("key", "deployment.environment.name");
        JsonObject envValue = new JsonObject();
        envValue.addProperty("stringValue", "local");
        envAttribute.add("value", envValue);

        JsonArray attributes = new JsonArray();
        attributes.add(resourceAttributes);
        attributes.add(envAttribute);

        JsonObject resource = new JsonObject();
        resource.add("attributes", attributes);

        JsonObject scope = new JsonObject();
        scope.addProperty("name", "python-http-demo");

        JsonObject logRecord = new JsonObject();
        logRecord.addProperty("timeUnixNano", currentNano);
        logRecord.addProperty("observedTimeUnixNano", currentNano);
        logRecord.addProperty("severityNumber", level.severityNumber);
        logRecord.addProperty("severityText", level.severityText);

        JsonObject body = new JsonObject();
        body.addProperty("stringValue", level.message);
        logRecord.add("body", body);

        JsonArray recordAttrs = new JsonArray();
        JsonObject sourceAttr = new JsonObject();
        sourceAttr.addProperty("key", "demo.source");
        JsonObject sourceValue = new JsonObject();
        sourceValue.addProperty("stringValue", "python");
        sourceAttr.add("value", sourceValue);
        recordAttrs.add(sourceAttr);
        logRecord.add("attributes", recordAttrs);

        JsonArray logRecords = new JsonArray();
        logRecords.add(logRecord);

        JsonObject scopeLog = new JsonObject();
        scopeLog.add("scope", scope);
        scopeLog.add("logRecords", logRecords);

        JsonArray scopeLogs = new JsonArray();
        scopeLogs.add(scopeLog);

        JsonObject resourceLog = new JsonObject();
        resourceLog.add("resource", resource);
        resourceLog.add("scopeLogs", scopeLogs);

        JsonArray resourceLogs = new JsonArray();
        resourceLogs.add(resourceLog);

        JsonObject root = new JsonObject();
        root.add("resourceLogs", resourceLogs);

        return root;
    }

    private static void doPost(HttpClient client, String apiUrl, String token, JsonObject payload) {
        JsonObject logRecord = payload.getAsJsonArray("resourceLogs")
                .get(0).getAsJsonObject()
                .getAsJsonArray("scopeLogs")
                .get(0).getAsJsonObject()
                .getAsJsonArray("logRecords")
                .get(0).getAsJsonObject();
        String severityText = logRecord.get("severityText").getAsString();
        int severityNumber = logRecord.get("severityNumber").getAsInt();
        LOGGER.info("Sending log level: " + severityText + " (" + severityNumber + ")");

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(apiUrl))
                .header("Content-Type", "application/json")
                .header("x-bk-token", token)
                .POST(HttpRequest.BodyPublishers.ofString(GSON.toJson(payload)))
                .timeout(Duration.ofSeconds(10))
                .build();

        try {
            HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
            LOGGER.info("response.status_code=" + response.statusCode() + ", body=" + response.body());
        } catch (IOException | InterruptedException e) {
            LOGGER.severe("failed to post request: " + e.getMessage());
        }
    }
}
```

## 3. 了解更多

进一步了解以下内容：

* 进行 <a href="{{docs.logs.learn_search}}" target="_blank">日志检索</a>。

* 了解 <a href="{{docs.logs.container_custom_report}}" target="_blank">容器日志自定义上报使用文档</a>。

* 了解 <a href="{{docs.logs.container_collector_install}}" target="_blank">容器日志采集器安装</a>。
