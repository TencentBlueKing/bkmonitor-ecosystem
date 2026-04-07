# java-事件（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/events/http/README.md" target="_blank">自定义事件 HTTP 上报</a>

### 1.2 上报速率限制

默认的 API 接收频率，单个 dataid 限制 1000 次／ min，单次上报 Body 最大为 500 KB。

如超过频率限制，请联系`蓝鲸助手`调整。

### 1.3 初始化 demo

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/events/java
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/events/http/README.md" target="_blank">自定义事件 HTTP 上报</a> 创建自定义事件后需关注提供的两个配置项：

* `TOKEN`：自定义事件数据源 Token，上报数据时使用。

* `数据 ID`: 数据 ID（Data ID），自定义事件数据源唯一标识，上报数据时使用。

同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image-1.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数     | 类型                | 描述                         |
| ------------ | ------------------- | ---------------------------- |
|`TOKEN`       |String      |❗❗【非常重要】自定义事件数据源 `Token`。  |
|`DATA_ID`       |Integer     |❗❗【非常重要】数据 ID（`Data ID`），自定义事件数据源唯一标识。|
|`API_URL`       |String         |❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写。|
|`INTERVAL` |Integer     |数据上报间隔，默认值为 60 秒。   |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/events/java" target="_blank">bkmonitor-ecosystem/examples/events/java</a> 中找到。

通过 docker build 构建名为 events-http-java 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、DATA_ID、API_URL 传递配置参数，实现周期上报事件：

```bash
docker build -t events-http-java .

docker run -e TOKEN="xxx" \
 -e DATA_ID=000000 \
 -e API_URL="http://127.0.0.1:10205/v2/push/" \
 -e INTERVAL=60 events-http-java
```

运行输出：

```bash
2026-03-23 15:35:31 | 事件上报服务启动 | 目标: 127.0.0.1 | 间隔: 60 秒
2026-03-23 15:35:31 | 生成事件数据:
[ {
  "target" : "127.0.0.1",
  "event" : {
    "content" : "CPU告警: 81%"
  },
  "timestamp" : 1774251331176,
  "event_name" : "cpu_alert",
  "dimension" : {
    "location" : "guangdong",
    "module" : "db"
  }
} ]
2026-03-23 15:35:31 | 上报结果: success 上报成功
```

### 2.4 样例代码

该样例通过模拟周期上报 CPU 及内存使用率（数值随机生成），演示如何进行自定义事件上报：

```java
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
```

## 3. 了解更多

* <a href="#" target="_blank">事件数据接入</a>。

* <a href="#" target="_blank">主机事件</a>。

* <a href="#" target="_blank">容器事件</a>。