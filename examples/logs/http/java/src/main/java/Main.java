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
