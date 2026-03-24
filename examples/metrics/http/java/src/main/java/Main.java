// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.text.SimpleDateFormat;
import java.time.Duration;
import java.util.Arrays;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;
import java.util.Random;
import java.util.concurrent.TimeUnit;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * 指标上报示例主类.
 */
public class Main {
    // ❗❗【非常重要】数据上报接口地址（Access URL）
    // 国内站点请填写「 {{access_config.custom.http}} 」
    // 其他环境、跨云场景请根据页面接入指引填写
    private static final String API_URL = System.getenv("API_URL");
    // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
    private static final String TOKEN = System.getenv("TOKEN");
    // ❗❗【非常重要】data_id，标识上报的数据类型，配置为应用数据 ID
    private static final String DATA_ID = System.getenv("DATA_ID");
    // 上报间隔，默认为60秒
    private static final long INTERVAL = System.getenv("INTERVAL") != null
            ? Long.parseLong(System.getenv("INTERVAL")) : 60;

    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();
    private static final Random RANDOM = new Random();

    private static double[] collectMetrics() {
        return new double[]{getCpuUsage(), getMemoryUsage()};
    }

    private static double getCpuUsage() {
        return RANDOM.nextDouble() * 100;
    }

    private static double getMemoryUsage() {
        return RANDOM.nextDouble() * 100;
    }

    /**
     * 发送指标报告.
     *
     * @param apiUrl API 地址
     * @param token 认证令牌
     * @param dataId 数据 ID
     * @param metrics 指标数据
     * @return HTTP 状态码
     */
    private static int sendReport(String apiUrl, String token,
                                  String dataId, double[] metrics) {
        try {
            Map<String, Object> metricsData = new HashMap<>();
            // 定义指定指标名及上报值。
            metricsData.put("cpu_load", metrics[0]);
            metricsData.put("memory_usage", metrics[1]);

            Map<String, String> dimension = new HashMap<>();
            // 定义上报维度。
            dimension.put("module", "server");
            dimension.put("region", "guangdong");
            dimension.put("language", "java");

            Map<String, Object> dataItem = new HashMap<>();
            dataItem.put("metrics", metricsData);
            dataItem.put("target", "127.0.0.1");
            dataItem.put("dimension", dimension);
            // 设置上报时间。
            dataItem.put("timestamp", System.currentTimeMillis());

            Map<String, Object> payload = new HashMap<>();
            // data_id，标识上报的数据类型，配置为应用数据 ID
            payload.put("data_id", Integer.parseInt(dataId));
            // 认证令牌，用于接口鉴定，配置为应用 TOKEN
            payload.put("access_token", token);
            payload.put("data", Arrays.asList(dataItem));

            String requestBody = OBJECT_MAPPER.writeValueAsString(payload);
            System.out.println("[Data upload] " + requestBody);

            HttpClient client = HttpClient.newBuilder()
                    .connectTimeout(Duration.ofSeconds(10))
                    .build();

            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(apiUrl))
                    .header("Content-Type", "application/json")
                    .header("Accept", "application/json")
                    .POST(HttpRequest.BodyPublishers.ofString(requestBody,
                            StandardCharsets.UTF_8))
                    .timeout(Duration.ofSeconds(10))
                    .build();

            HttpResponse<String> response = client.send(request,
                    HttpResponse.BodyHandlers.ofString());

            return response.statusCode();

        } catch (IOException e) {
            System.err.println("Network IO exception: " + e.getMessage());
            return 500;
        } catch (InterruptedException e) {
            System.err.println("Request was interrupted: " + e.getMessage());
            Thread.currentThread().interrupt();
            return 500;
        } catch (Exception e) {
            System.err.println("Request exception: " + e.getMessage());
            return 500;
        }
    }

    /**
     * 主函数入口.
     *
     * @param args 命令行参数
     */
    public static void main(String[] args) {
        if (API_URL == null || API_URL.isEmpty()
                || TOKEN == null || TOKEN.isEmpty()
                || DATA_ID == null || DATA_ID.isEmpty()) {
            System.err.println(
                    "Error: Environment variables API_URL, TOKEN and DATA_ID must be set");
            System.exit(1);
        }

        SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");

        while (true) {
            try {
                double[] metrics = collectMetrics();
                int statusCode = sendReport(API_URL, TOKEN, DATA_ID, metrics);

                String timestamp = sdf.format(new Date());
                if (statusCode == 200) {
                    System.out.printf(
                            "[%s] Report successful | CPU: %.2f%% Memory: %.2f%%\n",
                            timestamp, metrics[0], metrics[1]);
                } else {
                    System.out.printf("[%s] Report failed | Status code: %d\n",
                            timestamp, statusCode);
                }
            } catch (Exception e) {
                System.err.println("Exception in main loop: " + e.getMessage());
                e.printStackTrace();
            }

            try {
                TimeUnit.SECONDS.sleep(INTERVAL);
            } catch (InterruptedException e) {
                System.out.println("Monitoring interrupted");
                break;
            }
        }
    }
}
