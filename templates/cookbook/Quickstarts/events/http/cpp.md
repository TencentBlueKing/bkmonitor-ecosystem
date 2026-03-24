# C++-事件（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="{{docs.events.http.Http_Preadme}}" target="_blank">自定义事件 HTTP 上报</a>

### 1.2 上报速率限制

默认的 API 接收频率，单个 dataid 限制 1000 次／ min，单次上报 Body 最大为 500 KB。

如超过频率限制，请联系`蓝鲸助手`调整。

### 1.3 初始化 demo

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/events/cpp
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="{{docs.events.http.Http_Preadme}}" target="_blank">自定义事件 HTTP 上报</a> 创建自定义事件后需关注提供的两个配置项：

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
|`API_URL`       |String         |❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 {{access_config.custom.http}} 」，其他环境、跨云场景请根据页面接入指引填写。|
|`INTERVAL` |Integer     |上报间隔（单位为秒），默认 60 秒上报一次。​ |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/events/cpp" target="_blank">bkmonitor-ecosystem/examples/events/cpp</a> 中找到。

通过 docker build 构建名为 events-http-cpp 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、DATA_ID、API_URL 传递配置参数，实现周期上报事件：

```bash
docker build -t events-http-cpp .

docker run -e TOKEN="{{access_config.token}}" \
 -e DATA_ID=000000 \
 -e API_URL="{{access_config.custom.http}}" \
 -e INTERVAL=60 events-http-cpp
```

运行输出：

```shell
2026-03-23 16:54:34 | ====== 事件上报服务启动 ======
2026-03-23 16:54:34 | 目标设备: 127.0.0.1 | 上报间隔: 60秒
2026-03-23 16:54:34 | 生成事件数据:
[
  {
    "dimension": {
      "location": "guangdong",
      "module": "db"
    },
    "event": {
      "content": "CPU 告警: 85%"
    },
    "event_name": "cpu_alert",
    "target": "127.0.0.1",
    "timestamp": 1774256074719
  }
]
2026-03-23 16:54:34 | 上报结果: success 上报成功
```

### 2.4 样例代码

该样例通过模拟周期上报 CPU 及内存使用率（数值随机生成），演示如何进行自定义事件上报：

```cpp
{% raw %}
#include <chrono>
#include <cstdlib>
#include <iostream>
#include <random>
#include <string>
#include <thread>
#include "httplib.h"
#include "nlohmann/json.hpp"

using json = nlohmann::json;

class EventReporter {
    static std::string getenv_str(const char* name, const std::string& fallback = "") {
        const char* val = std::getenv(name);
        return val ? val : fallback;
    }
    static int getenv_int(const char* name, int fallback = 0) {
        const char* val = std::getenv(name);
        return val ? std::atoi(val) : fallback;
    }
    // ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
    std::string api_url = getenv_str("API_URL");
    // ❗❗【非常重要】认证令牌，配置为应用 `TOKEN`
    std::string token = getenv_str("TOKEN");
    // ❗❗【非常重要】数据 `ID`
    int data_id = getenv_int("DATA_ID");
    // 目标设备IP
    std::string target_ip = getenv_str("TARGET_IP", "127.0.0.1");
    // 上报间隔（秒）
    int interval = getenv_int("INTERVAL", 60);

    void log(const std::string& msg) {
        auto t = std::chrono::system_clock::to_time_t(std::chrono::system_clock::now());
        std::cout << std::put_time(std::localtime(&t), "%Y-%m-%d %H:%M:%S") << " | " << msg << std::endl;
    }

    // 步骤1：构造事件数据
    json generate_event() {
        static std::random_device rd;
        static std::mt19937 gen(rd());
        static std::uniform_int_distribution<int> dis(80, 99);
        auto ts = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()).count();
        return {
            {"event_name", "cpu_alert"},
            // 事件内容（80-99%随机值）
            {"event", {{"content", "CPU告警: " + std::to_string(dis(gen)) + "%"}}},
            {"target", target_ip},
            // 事件维度
            {"dimension", {{"module", "db"}, {"location", "guangdong"}}},
            {"timestamp", ts}
        };
    }

    // 步骤2：发送事件到上报接口
    json send_event() {
        auto event = generate_event();
        // 组装上报 payload：
        // 包含 data_id：❗❗【非常重要】数据 `ID`
        // 包含 access_token：❗❗【非常重要】认证令牌，配置为应用 `TOKEN`
        // 包含 事件数据
        json payload = {{"data_id", data_id}, {"access_token", token}, {"data", json::array({event})}};
        log("生成事件数据:");
        std::cout << json::array({event}).dump(2) << std::endl;

        // 解析 URL（格式: "http://host:port/path"）
        // ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
        size_t p = api_url.find("://");
        size_t path = api_url.find('/', p + 3);
        std::string host = (path != std::string::npos) ? api_url.substr(0, path) : api_url;
        std::string route = (path != std::string::npos) ? api_url.substr(path) : "/";
        httplib::Client client(host.c_str());
        client.set_connection_timeout(5);
        client.set_read_timeout(5);

        // 发送 POST 请求并返回结果
        auto res = client.Post(route.c_str(), payload.dump(), "application/json");
        if (res && res->status == 200)
            return {{"status", "success"}, {"message", "上报成功"}};

        std::string err = res ? "HTTP " + std::to_string(res->status) : "连接失败";
        return {{"status", "error"}, {"message", err}};
    }

 public:
    // 步骤3：启动上报循环
    void run() {
        log("====== 事件上报服务启动 ======");
        log("目标设备: " + target_ip + " | 上报间隔: " + std::to_string(interval) + "秒");

        // 持续上报，每次上报后等待 interval 秒
        while (true) {
            auto res = send_event();
            std::string status = res["status"];
            std::string color = (status == "success") ? "\033[32m" : "\033[31m";
            log("上报结果: " + color + status + " " + res["message"].get<std::string>() + "\033[0m");

            std::this_thread::sleep_for(std::chrono::seconds(interval));
        }
    }
};

int main() {
    EventReporter reporter;
    reporter.run();
    return 0;
}
{% endraw %}
```

## 3. 了解更多

* <a href="{{docs.events.report_access}}" target="_blank">事件数据接入</a>。

* <a href="{{docs.events.Host_events}}" target="_blank">主机事件</a>。

* <a href="{{docs.events.Container_events}}" target="_blank">容器事件</a>。
