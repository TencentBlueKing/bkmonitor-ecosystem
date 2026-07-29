# Java-日志（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

{{LOGS_HTTP_TERM_INTRO}}

### 1.2 开发环境要求

{{LOGS_HTTP_DEV_REQUIREMENTS}}

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

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.List;
import java.util.Map;
import java.util.Random;
import java.util.logging.Level;
import java.util.logging.Logger;


public class Main {

    private static final Logger LOGGER = Logger.getLogger(Main.class.getName());
    private static final Random RANDOM = new Random();
    private static final Gson GSON = new GsonBuilder().create();

    // ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
    private static final String DEFAULT_TOKEN = "fixme";
    /** ❗❗【非常重要】上报地址，国内站点默认是「{{access_config.otlp.http_endpoint}}/v1/logs」，
        * 其他环境、跨云场景请根据页面接入指引填写
        */
    private static final String DEFAULT_API_URL = "{{access_config.otlp.http_endpoint}}/v1/logs";
    private static final int HTTP_TIMEOUT_SECONDS = 10;

    // 日志级别映射
    private static final List<Map<String, Object>> LOG_LEVELS = List.of(
        Map.of("severityNumber", 5, "severityText", "DEBUG", "message", "debug log from java http"),
        Map.of("severityNumber", 9, "severityText", "INFO", "message", "info log from java http"),
        Map.of("severityNumber", 13, "severityText", "WARN", "message", "warn log from java http"),
        Map.of("severityNumber", 17, "severityText", "ERROR", "message", "error log from java http")
    );

    public static void main(String[] args) {
        initLogger();

        LOGGER.info("Starting log reporter (press Ctrl+C to stop)...");

        // 从环境变量获取配置，支持自定义
        String token = System.getenv().getOrDefault("TOKEN", DEFAULT_TOKEN);
        String apiUrl = System.getenv().getOrDefault("API_URL", DEFAULT_API_URL);

        HttpClient client = createHttpClient();

        runLogReporter(client, apiUrl, token);
    }

    /**
     * 初始化日志记录器
     */
    private static void initLogger() {
        LOGGER.setLevel(Level.INFO);
        System.setProperty("java.util.logging.SimpleFormatter.format",
                "%1$tF %1$tT,%1$tL - %4$s - %5$s%n");
    }

    /**
     * 创建 HTTP 客户端
     */
    private static HttpClient createHttpClient() {
        return HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(HTTP_TIMEOUT_SECONDS))
                .build();
    }

    /**
     * 运行日志上报循环
     */
    private static void runLogReporter(HttpClient client, String apiUrl, String token) {
        while (true) {
            try {
                Map<String, Object> level = getRandomLevel();
                Map<String, Object> payload = buildPayload(level);
                LOGGER.info("Sending log level: " + level.get("severityText") + " (" + level.get("severityNumber") + ")");
                doPost(client, apiUrl, token, payload);
                Thread.sleep(100);
            } catch (IOException e) {
                LOGGER.severe("failed to post request: " + e.getMessage());
            } catch (InterruptedException e) {
                LOGGER.info("Received keyboard interrupt, exiting...");
                Thread.currentThread().interrupt();
                break;
            }
        }
    }

    /**
     * 获取当前时间的纳秒时间戳
     */
    private static String getCurrentNanoTimestamp() {
        return String.valueOf(System.currentTimeMillis() * 1_000_000L +
                (System.nanoTime() % 1_000_000L));
    }

    /**
     * 随机获取一个日志级别
     */
    @SuppressWarnings("unchecked")
    private static Map<String, Object> getRandomLevel() {
        return LOG_LEVELS.get(RANDOM.nextInt(LOG_LEVELS.size()));
    }

    /**
     * 构建 OTLP 日志上报数据体
     * 结构参考：https://opentelemetry.io/docs/specs/otel/protocol/logs-data-model/
     *
     * @return Map 格式的 OTLP 日志数据，结构形如最终的 JSON 上报体
     */
    private static Map<String, Object> buildPayload(Map<String, Object> level) {
        String currentNano = getCurrentNanoTimestamp();

        return Map.of("resourceLogs", List.of(Map.of(
            "resource", Map.of("attributes", List.of(
                Map.of("key", "service.name", "value", Map.of("stringValue", "custom-log-demo")),
                Map.of("key", "deployment.environment.name", "value", Map.of("stringValue", "local"))
            )),
            "scopeLogs", List.of(Map.of(
                "scope", Map.of("name", "java-http-demo"),
                "logRecords", List.of(Map.of(
                    "timeUnixNano", currentNano,
                    "observedTimeUnixNano", currentNano,
                    "severityNumber", level.get("severityNumber"),
                    "severityText", level.get("severityText"),
                    "body", Map.of("stringValue", level.get("message")),
                    "attributes", List.of(
                        Map.of("key", "demo.source", "value", Map.of("stringValue", "java"))
                    )
                ))
            ))
        )));
    }

    /**
     * 发送 HTTP POST 请求上报日志
     *
     * @param client HTTP 客户端
     * @param apiUrl 上报地址
     * @param token 认证令牌
     * @param payload 日志数据体
     */
    private static void doPost(HttpClient client, String apiUrl, String token, Map<String, Object> payload)
            throws IOException, InterruptedException {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(apiUrl))
                .header("Content-Type", "application/json")
                .header("x-bk-token", token)
                .POST(HttpRequest.BodyPublishers.ofString(GSON.toJson(payload)))
                .timeout(Duration.ofSeconds(HTTP_TIMEOUT_SECONDS))
                .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        LOGGER.info("response.status_code=" + response.statusCode() + ", body=" + response.body());
    }
}
```

## 3. 了解更多

{{LOGS_HTTP_LEARN_MORE}}
