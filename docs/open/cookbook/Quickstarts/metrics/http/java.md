# java-指标（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Term/metrics/what.md" target="_blank">什么是指标</a>

* <a href="{{COOKBOOK_METRICS_TYPES}}" target="_blank">指标类型</a>

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/metrics/http/java
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Quickstarts/metrics/http/README.md" target="_blank">自定义指标 HTTP 上报</a> 创建一个上报协议为 `JSON` 的自定义指标，关注创建后提供的两个配置项：

* `TOKEN`：自定义指标数据源 Token，上报数据时使用。

* `数据 ID`: 数据 ID（Data ID），自定义指标数据源唯一标识，上报数据时使用。
同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数         | 类型      | 描述                                                                                                 |
|------------|---------|----------------------------------------------------------------------------------------------------|
| `TOKEN`    | String  | ❗❗【非常重要】 自定义指标数据源 `Token`。                                                                               |
| `DATA_ID`  | Integer | ❗❗【非常重要】 数据 ID（`Data ID`），自定义指标数据源唯一标识。                                                                         |
| `API_URL`  | String  | ❗❗【非常重要】 数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写。 |
| `INTERVAL` | Integer | 数据上报间隔，默认值为 60 秒。       ​​                                                             |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/metrics/http/java" target="_blank">bkmonitor-ecosystem/examples/metrics/http/java</a> 中找到。

复制以下命令参数在你的终端运行：

```bash
docker build -t metrics-http-java .

docker run -e TOKEN="xxx" \
 -e DATA_ID=00000 \
 -e API_URL="http://127.0.0.1:10205/v2/push/" \
 -e INTERVAL=60 metrics-http-java
```

运行输出：

```bash
[Data upload] {"access_token":"fixme","data":[{"metrics":{"memory_usage":3.095511188657265,"cpu_load":0.0},"dimension":{"module":"server","region":"guangdong"},"target":"127.0.0.1","timestamp":1761290098246}],"data_id":fixme}
[2025-10-24 15:14:59] Report successful | CPU: 0.00% Memory: 3.10%
[Data upload] {"access_token":"fixme","data":[{"metrics":{"memory_usage":5.040454113577295,"cpu_load":11.398525323340989},"dimension":{"module":"server","region":"guangdong"},"target":"127.0.0.1","timestamp":1761290129490}],"data_id":fixme}
[2025-10-24 15:15:29] Report successful | CPU: 11.40% Memory: 5.04%
[Data upload] {"access_token":"fixme","data":[{"metrics":{"memory_usage":5.153520839420829,"cpu_load":11.864691316573635},"dimension":{"module":"server","region":"guangdong"},"target":"127.0.0.1","timestamp":1761290159529}],"data_id":fixme}
[2025-10-24 15:15:59] Report successful | CPU: 11.86% Memory: 5.15%
...
```

### 2.4 样例代码

该样例通过模拟周期上报 CPU 及内存使用率（数值随机生成），演示如何进行自定义指标上报：

```java
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.text.SimpleDateFormat;
import java.util.*;
import java.util.concurrent.TimeUnit;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.URI;
import java.time.Duration;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.core.JsonProcessingException;

public class Main {
    // ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写
    private static final String API_URL = System.getenv("API_URL");
    private static final String TOKEN = System.getenv("TOKEN");        // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
    private static final String DATA_ID = System.getenv("DATA_ID");    // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
    private static final long INTERVAL = System.getenv("INTERVAL") != null ?
        Long.parseLong(System.getenv("INTERVAL")) : 60;      // 上报间隔，默认为60秒

    private static final ObjectMapper objectMapper = new ObjectMapper();
    private static final Random random = new Random();

    private static double[] collectMetrics() {
        return new double[]{getCpuUsage(), getMemoryUsage()};
    }

    private static double getCpuUsage() {
        return random.nextDouble() * 100;
    }

    private static double getMemoryUsage() {
        return random.nextDouble() * 100;
    }

    private static int sendReport(String apiUrl, String token, String dataId, double[] metrics) {
        try {
            Map<String, Object> metricsData = new HashMap<>();
            // 定义指定指标名及上报值。
            metricsData.put("cpu_load", metrics[0]);
            metricsData.put("memory_usage", metrics[1]);

            Map<String, String> dimension = new HashMap<>();
            // 定义上报维度。
            dimension.put("module", "server");
            dimension.put("region", "guangdong");

            Map<String, Object> dataItem = new HashMap<>();
            dataItem.put("metrics", metricsData);
            dataItem.put("target", "127.0.0.1");
            dataItem.put("dimension", dimension);
            // 设置上报时间。
            dataItem.put("timestamp", System.currentTimeMillis());

            Map<String, Object> payload = new HashMap<>();
            payload.put("data_id", Integer.parseInt(dataId));     // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
            payload.put("access_token", token);                    // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
            payload.put("data", Arrays.asList(dataItem));

            String requestBody = objectMapper.writeValueAsString(payload);
            System.out.println("[Data upload] " + requestBody);

            HttpClient client = HttpClient.newBuilder()
                    .connectTimeout(Duration.ofSeconds(10))
                    .build();

            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(apiUrl))
                    .header("Content-Type", "application/json")
                    .header("Accept", "application/json")
                    .POST(HttpRequest.BodyPublishers.ofString(requestBody, StandardCharsets.UTF_8))
                    .timeout(Duration.ofSeconds(10))
                    .build();

            HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());

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

    public static void main(String[] args) {
        if (API_URL == null || API_URL.isEmpty() || TOKEN == null || TOKEN.isEmpty() ||
            DATA_ID == null || DATA_ID.isEmpty()) {
            System.err.println("Error: Environment variables API_URL, TOKEN and DATA_ID must be set");
            System.exit(1);
        }

        SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");

        while (true) {
            try {
                double[] metrics = collectMetrics();
                int statusCode = sendReport(API_URL, TOKEN, DATA_ID, metrics);

                String timestamp = sdf.format(new Date());
                if (statusCode == 200) {
                    System.out.printf("[%s] Report successful | CPU: %.2f%% Memory: %.2f%%\n", timestamp, metrics[0], metrics[1]);
                } else {
                    System.out.printf("[%s] Report failed | Status code: %d\n", timestamp, statusCode);
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
```

## 3. 了解更多

* 进行 <a href="#" target="_blank">指标检索</a>。

* 了解 <a href="#" target="_blank">怎么使用监控指标</a>。

* 了解如何 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/dashboard.md" target="_blank">配置仪表盘</a>。

* 了解如何使用 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/alarm-configurations/rules.md" target="_blank">监控告警</a>。