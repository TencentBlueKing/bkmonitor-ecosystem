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
import java.nio.charset.StandardCharsets;
import java.text.SimpleDateFormat;
import java.time.Duration;
import java.util.Date;
import java.util.List;
import java.util.Map;
import java.util.Random;
import java.util.concurrent.TimeUnit;
import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * 指标上报示例主类.
 */
public class Main {
    // ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
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
    private static final HttpClient HTTP_CLIENT = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(10)).build();

    /**
     * 发送指标报告.
     *
     * @param cpuLoad CPU 使用率
     * @param memUsage 内存使用率
     * @return HTTP 状态码
     */
    private static int sendReport(double cpuLoad, double memUsage) throws Exception {
        // 定义指定指标名及上报值
        Map<String, Object> metricsData = Map.of(
                "cpu_load", cpuLoad, "memory_usage", memUsage);
        // 定义上报维度
        Map<String, String> dimension = Map.of(
                "module", "server", "region", "guangdong");

        Map<String, Object> dataItem = Map.of(
                "metrics", metricsData,
                "target", "127.0.0.1",
                "dimension", dimension,
                // 设置上报时间
                "timestamp", System.currentTimeMillis());

        Map<String, Object> payload = Map.of(
                // ❗❗【非常重要】data_id，标识上报的数据类型，配置为应用数据 ID
                "data_id", Integer.parseInt(DATA_ID),
                // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
                "access_token", TOKEN,
                "data", List.of(dataItem));

        String requestBody = OBJECT_MAPPER.writeValueAsString(payload);

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(API_URL))
                .header("Content-Type", "application/json")
                .header("Accept", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(
                        requestBody, StandardCharsets.UTF_8))
                .timeout(Duration.ofSeconds(10))
                .build();

        return HTTP_CLIENT.send(request,
                HttpResponse.BodyHandlers.ofString()).statusCode();
    }

    /**
     * 主函数入口.
     *
     * @param args 命令行参数
     */
    public static void main(String[] args) throws Exception {
        System.out.println("🚀 启动指标上报服务");
        System.out.println("API地址: " + API_URL);
        System.out.println("数据ID: " + DATA_ID);
        System.out.println("上报间隔: " + INTERVAL + "秒");
        System.out.println("=================================");

        SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
        while (true) {
            double cpuLoad = RANDOM.nextDouble() * 100;
            double memUsage = RANDOM.nextDouble() * 100;
            int statusCode = sendReport(cpuLoad, memUsage);
            String timestamp = sdf.format(new Date());
            if (statusCode == 200) {
                System.out.printf(
                        "[%s] ✅ 上报成功 | CPU: %.2f%% 内存: %.2f%%\n",
                        timestamp, cpuLoad, memUsage);
            } else {
                System.out.printf("[%s] ❌ 上报失败 | 状态码: %d\n",
                        timestamp, statusCode);
            }

            TimeUnit.SECONDS.sleep(INTERVAL);
        }
    }
}
