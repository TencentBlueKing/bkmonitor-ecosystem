// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

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

    // ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
    private static final String TOKEN = "fixme";
    /** ❗❗【非常重要】上报地址，国内站点默认是「{{access_config.otlp.http_endpoint}}/v1/logs」，
        * 其他环境、跨云场景请根据页面接入指引填写
        */
    private static final String API_URL = "{{access_config.otlp.http_endpoint}}/v1/logs";
    private static final String SERVICE_NAME = "custom-log-demo";
    private static final String DEPLOYMENT_ENV = "local";
    private static final String SCOPE_NAME = "java-http-demo";
    private static final String LOG_SOURCE = "java";
    private static final long SLEEP_INTERVAL_MS = 100;
    private static final int HTTP_TIMEOUT_SECONDS = 10;

    // 日志级别映射
    private static final LogLevel[] LOG_LEVELS = {
        new LogLevel(5, "DEBUG", "debug log from java http"),
        new LogLevel(9, "INFO", "info log from java http"),
        new LogLevel(13, "WARN", "warn log from java http"),
        new LogLevel(17, "ERROR", "error log from java http")
    };

    /**
     * 日志级别封装类
     */
    private static class LogLevel {
        private final int severityNumber;
        private final String severityText;
        private final String message;

        LogLevel(int severityNumber, String severityText, String message) {
            this.severityNumber = severityNumber;
            this.severityText = severityText;
            this.message = message;
        }

        int getSeverityNumber() {
            return severityNumber;
        }

        String getSeverityText() {
            return severityText;
        }

        String getMessage() {
            return message;
        }
    }

    public static void main(String[] args) {
        initLogger();

        LOGGER.info("Starting log reporter (press Ctrl+C to stop)...");

        // 从环境变量获取配置，支持自定义
        String token = System.getenv().getOrDefault("TOKEN", TOKEN);
        String apiUrl = System.getenv().getOrDefault("API_URL", API_URL);

        HttpClient client = createHttpClient();

        try {
            runLogReporter(client, apiUrl, token);
        } catch (InterruptedException e) {
            LOGGER.info("Received keyboard interrupt, exiting...");
            Thread.currentThread().interrupt();
        }
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
    private static void runLogReporter(HttpClient client, String apiUrl, String token)
            throws InterruptedException {
        while (true) {
            JsonObject payload = buildPayload();
            doPost(client, apiUrl, token, payload);
            Thread.sleep(SLEEP_INTERVAL_MS);
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
    private static LogLevel getRandomLevel() {
        return LOG_LEVELS[RANDOM.nextInt(LOG_LEVELS.length)];
    }

    /**
     * 构建 OTLP 日志上报数据体
     * 结构参考：https://opentelemetry.io/docs/specs/otel/protocol/logs-data-model/
     *
     * @return JsonObject OTLP 格式的日志数据
     */
    private static JsonObject buildPayload() {
        String currentNano = getCurrentNanoTimestamp();
        LogLevel level = getRandomLevel();

        // 日志体
        JsonObject body = new JsonObject();
        body.addProperty("stringValue", level.getMessage());

        // 日志记录属性
        JsonArray logAttributes = new JsonArray();
        logAttributes.add(createStringAttribute("demo.source", LOG_SOURCE));

        // 单条日志记录
        JsonObject logRecord = new JsonObject();
        logRecord.addProperty("timeUnixNano", currentNano);
        logRecord.addProperty("observedTimeUnixNano", currentNano);
        logRecord.addProperty("severityNumber", level.getSeverityNumber());
        logRecord.addProperty("severityText", level.getSeverityText());
        logRecord.add("body", body);
        logRecord.add("attributes", logAttributes);

        // Scope 日志
        JsonObject scope = new JsonObject();
        scope.addProperty("name", SCOPE_NAME);

        JsonObject scopeLog = new JsonObject();
        scopeLog.add("scope", scope);
        JsonArray logRecords = new JsonArray();
        logRecords.add(logRecord);
        scopeLog.add("logRecords", logRecords);

        // 资源属性
        JsonArray resourceAttributes = new JsonArray();
        resourceAttributes.add(createStringAttribute("service.name", SERVICE_NAME));
        resourceAttributes.add(createStringAttribute("deployment.environment.name", DEPLOYMENT_ENV));

        JsonObject resource = new JsonObject();
        resource.add("attributes", resourceAttributes);

        // 资源日志
        JsonObject resourceLog = new JsonObject();
        resourceLog.add("resource", resource);
        JsonArray scopeLogs = new JsonArray();
        scopeLogs.add(scopeLog);
        resourceLog.add("scopeLogs", scopeLogs);

        // 根对象
        JsonObject root = new JsonObject();
        JsonArray resourceLogs = new JsonArray();
        resourceLogs.add(resourceLog);
        root.add("resourceLogs", resourceLogs);

        return root;
    }

    /**
     * 创建字符串类型属性键值对
     *
     * @param key 属性键
     * @param value 属性值
     * @return JsonObject 属性对象
     */
    private static JsonObject createStringAttribute(String key, String value) {
        JsonObject attribute = new JsonObject();
        attribute.addProperty("key", key);
        JsonObject valueObj = new JsonObject();
        valueObj.addProperty("stringValue", value);
        attribute.add("value", valueObj);
        return attribute;
    }

    /**
     * 发送 HTTP POST 请求上报告警日志
     *
     * @param client HTTP 客户端
     * @param apiUrl 上报地址
     * @param token 认证令牌
     * @param payload 日志数据体
     */
    private static void doPost(HttpClient client, String apiUrl, String token, JsonObject payload) {
        // 提取日志级别信息用于日志记录
        String severityText = extractSeverityText(payload);
        int severityNumber = extractSeverityNumber(payload);
        LOGGER.info("Sending log level: " + severityText + " (" + severityNumber + ")");

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(apiUrl))
                .header("Content-Type", "application/json")
                .header("x-bk-token", token)
                .POST(HttpRequest.BodyPublishers.ofString(GSON.toJson(payload)))
                .timeout(Duration.ofSeconds(HTTP_TIMEOUT_SECONDS))
                .build();

        try {
            HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
            LOGGER.info("response.status_code=" + response.statusCode() + ", body=" + response.body());
        } catch (IOException e) {
            LOGGER.severe("IO error while posting request: " + e.getMessage());
            Thread.currentThread().interrupt();
        } catch (InterruptedException e) {
            LOGGER.severe("Interrupted while posting request: " + e.getMessage());
            Thread.currentThread().interrupt();
        }
    }

    /**
     * 从 payload 中提取日志级别文本
     * 路径：resourceLogs[0] -> scopeLogs[0] -> logRecords[0] -> severityText
     */
    private static String extractSeverityText(JsonObject payload) {
        JsonObject logRecord = extractLogRecord(payload);
        return logRecord.get("severityText").getAsString();
    }

    /**
     * 从 payload 中提取日志级别数字
     * 路径：resourceLogs[0] -> scopeLogs[0] -> logRecords[0] -> severityNumber
     */
    private static int extractSeverityNumber(JsonObject payload) {
        JsonObject logRecord = extractLogRecord(payload);
        return logRecord.get("severityNumber").getAsInt();
    }

    /**
     * 提取第一条日志记录（公共辅助方法）
     */
    private static JsonObject extractLogRecord(JsonObject payload) {
        return payload.getAsJsonArray("resourceLogs")
                .get(0).getAsJsonObject()
                .getAsJsonArray("scopeLogs")
                .get(0).getAsJsonObject()
                .getAsJsonArray("logRecords")
                .get(0).getAsJsonObject();
    }
}
