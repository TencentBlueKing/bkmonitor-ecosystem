// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.Map;
import java.util.Random;
import com.fasterxml.jackson.databind.ObjectMapper;

public class Main {
    // ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
    private static final String API_URL = System.getenv().getOrDefault("API_URL", "fixme");
    // ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
    private static final int DATA_ID = Integer.parseInt(System.getenv().getOrDefault("DATA_ID", "0"));
    // ❗❗ 【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
    private static final String TOKEN = System.getenv().getOrDefault("TOKEN", "fixme");
    // 目标设备IP
    private static final String TARGET_IP = System.getenv().getOrDefault("TARGET_IP", "127.0.0.1");
    // 上报间隔（秒）
    private static final int INTERVAL = Integer.parseInt(System.getenv().getOrDefault("INTERVAL", "60"));

    private static final HttpClient httpClient = HttpClient.newBuilder()
        .connectTimeout(Duration.ofSeconds(5))
        .build();
    private static final Random random = new Random();
    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();
    private static final DateTimeFormatter TIME_FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private static void log(String msg) {
        System.out.printf("\033[1m%s\033[0m | %s%n",
            LocalDateTime.now().format(TIME_FMT), msg);
    }

    // 发送事件并返回 [status, message]
    private static String[] sendEvent() {
        try {
            // 生成80-99的随机整数
            int cpuUsage = 80 + random.nextInt(20);
            Map<String, Object> eventData = Map.of(
                "event_name", "cpu_alert",
                "event", Map.of("content", "CPU告警: " + cpuUsage + "%"),
                "target", TARGET_IP,
                "dimension", Map.of("module", "db", "location", "guangdong"),
                "timestamp", System.currentTimeMillis()
            );
            Map<String, Object> payload = Map.of(
                // ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
                "data_id", DATA_ID,
                // ❗❗ 【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
                "access_token", TOKEN,
                "data", List.of(eventData)
            );
            String prettyEvent = OBJECT_MAPPER.writerWithDefaultPrettyPrinter()
                .writeValueAsString(List.of(eventData));
            log("生成事件数据:\n" + prettyEvent);

            // ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
            HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(API_URL))
                .header("Content-Type", "application/json")
                .timeout(Duration.ofSeconds(5))
                .POST(HttpRequest.BodyPublishers.ofString(OBJECT_MAPPER.writeValueAsString(payload)))
                .build();

            HttpResponse<String> response = httpClient.send(
                request, HttpResponse.BodyHandlers.ofString()
            );

            if (response.statusCode() == 200) {
                return new String[]{"success", "上报成功"};
            }
            return new String[]{"error", "HTTP " + response.statusCode()};
        } catch (Exception e) {
            return new String[]{"error", "上报失败: " + e.getMessage()};
        }
    }

    public static void main(String[] args) {
        log("事件上报服务启动 | 目标: " + TARGET_IP + " | 间隔: " + INTERVAL + "秒");

        // 持续上报，每次上报后等待 INTERVAL 秒
        while (true) {
            String[] result = sendEvent();
            String color = "success".equals(result[0]) ? "\033[32m" : "\033[31m";
            log("上报结果: " + color + result[0] + " " + result[1] + "\033[0m");
            try {
                Thread.sleep(INTERVAL * 1000L);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }
    }
}
