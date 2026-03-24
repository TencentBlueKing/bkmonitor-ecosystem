# C++-指标（HTTP）上报

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
cd bkmonitor-ecosystem/examples/metrics/http/cpp
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
| `INTERVAL` | Integer | 数据上报间隔，默认值为 60 秒。    ​​                                                             |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/metrics/http/cpp" target="_blank">bkmonitor-ecosystem/examples/metrics/http/cpp</a> 中找到。

通过 docker build 构建名为 metrics-http-cpp 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、DATA_ID、API_URL 传递配置参数，实现周期上报指标：

```bash
docker build -t metrics-http-cpp .

docker run -e TOKEN="xxx" \
 -e DATA_ID=00000 \
 -e API_URL="http://127.0.0.1:10205/v2/push/" \
 -e INTERVAL=60 metrics-http-cpp
```

运行输出：

```bash
🔧 初始化指标上报服务...
🚀 启动指标上报服务
API地址: http://127.0.0.1:10205/v2/push/
数据ID: 00000
上报间隔: 60秒
=================================
[2025-12-29 08:59:30] 📡 HTTP状态码: 200
[2025-12-29 08:59:30] ✅ 上报成功 | CPU: 84.56% 内存: 17.64%
[2025-12-29 09:00:30] ✅ 上报成功 | CPU: 92.16% 内存: 44.53%
[2025-12-29 09:01:30] ✅ 上报成功 | CPU: 59.88% 内存: 95.42%
...
```

### 2.4 样例代码

该样例通过模拟周期上报 CPU 及内存使用率（数值随机生成），演示如何进行自定义指标上报：

```cpp
#include <cpr/cpr.h>
#include <nlohmann/json.hpp>
#include <iostream>
#include <string>
#include <random>
#include <chrono>
#include <thread>
#include <cstdlib>
#include <iomanip>
#include <atomic>
#include <csignal>

using json = nlohmann::json;

std::atomic<bool> g_running{true};

void signalHandler(int signal) {
    std::cout << "\n📡 接收到终止信号，正在退出..." << std::endl;
    g_running = false;
}

class ElegantMetricsReporter {
private:
    std::string api_url;
    std::string token;
    int data_id;
    int interval;
    std::random_device rd;
    std::mt19937 gen;
    std::uniform_real_distribution<double> dist;
    bool first_request;

public:
    ElegantMetricsReporter() : gen(rd()), dist(0.0, 100.0), first_request(true) {
        // ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写
        api_url = getEnv("API_URL", "");
        token = getEnv("TOKEN", ""); // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
        std::string data_id_str = getEnv("DATA_ID", ""); // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
        std::string interval_str = getEnv("INTERVAL", "60");  // 修改为60秒间隔

        data_id = std::stoi(data_id_str);
        interval = std::stoi(interval_str);
    }

    std::string getEnv(const std::string& key, const std::string& defaultValue = "") {
        const char* value = std::getenv(key.c_str());
        return value ? std::string(value) : defaultValue;
    }

    std::pair<double, double> collectMetrics() {
        double cpu_load = dist(gen);
        double mem_usage = dist(gen);
        return {cpu_load, mem_usage};
    }

    std::string buildJsonPayload(double cpu_load, double mem_usage) {
        json payload = {
            {"data_id", data_id},
            {"access_token", token},
            {"data", {{
                {"metrics", {
                    {"cpu_load", cpu_load},
                    {"memory_usage", mem_usage}
                }},
                {"target", "127.0.0.1"},
                {"dimension", {
                    {"module", "server"},
                    {"region", "guangdong"}
                }},
                {"timestamp", std::chrono::duration_cast<std::chrono::milliseconds>(
                    std::chrono::system_clock::now().time_since_epoch()).count()}
            }}}
        };

        return payload.dump();
    }

    std::string getCurrentTimestamp() {
        auto now = std::chrono::system_clock::now();
        auto time_t = std::chrono::system_clock::to_time_t(now);
        auto tm = *std::localtime(&time_t);

        std::ostringstream oss;
        oss << std::put_time(&tm, "%Y-%m-%d %H:%M:%S");
        return oss.str();
    }

    bool sendReport(double cpu_load, double mem_usage) {
        try {
            cpr::Response response = cpr::Post(
                cpr::Url{api_url},
                cpr::Header{{"Content-Type", "application/json"}},
                cpr::Body{buildJsonPayload(cpu_load, mem_usage)},
                cpr::Timeout{std::chrono::seconds(10)}
            );

            if (first_request) {
                std::cout << "[" << getCurrentTimestamp() << "] 📡 HTTP状态码: " << response.status_code << std::endl;
                first_request = false;
            }

            if (response.error) {
                return false;
            }

            return response.status_code == 200;

        } catch (const std::exception& e) {
            if (first_request) {
                first_request = false;
            }
            return false;
        }
    }

    void run() {
        std::cout << "🚀 启动指标上报服务" << std::endl;
        std::cout << "API地址: " << api_url << std::endl;
        std::cout << "数据ID: " << data_id << std::endl;
        std::cout << "上报间隔: " << interval << "秒" << std::endl;
        std::cout << "=================================" << std::endl;

        int success_count = 0;
        int total_count = 0;

        while (g_running) {
            auto metrics = collectMetrics();
            bool success = sendReport(metrics.first, metrics.second);

            std::string timestamp = getCurrentTimestamp();
            if (success) {
                std::cout << "[" << timestamp << "] ✅ 上报成功 | "
                          << "CPU: " << std::fixed << std::setprecision(2) << metrics.first << "% "
                          << "内存: " << metrics.second << "%" << std::endl;
                success_count++;
            } else {
                std::cout << "[" << timestamp << "] ❌ 上报失败 | "
                          << "CPU: " << metrics.first << "% "
                          << "内存: " << metrics.second << "%" << std::endl;
            }

            total_count++;

            for (int i = 0; i < interval && g_running; ++i) {
                std::this_thread::sleep_for(std::chrono::seconds(1));
            }
        }

        std::cout << "=================================" << std::endl;
        std::cout << "📊 服务运行统计:" << std::endl;
        std::cout << "总上报次数: " << total_count << std::endl;
        std::cout << "成功次数: " << success_count << std::endl;
        if (total_count > 0) {
            std::cout << "成功率: " << std::fixed << std::setprecision(1)
                      << (static_cast<double>(success_count) / total_count * 100) << "%" << std::endl;
        }
        std::cout << "👋 服务已停止" << std::endl;
    }
};

int main() {
    std::signal(SIGINT, signalHandler);
    std::signal(SIGTERM, signalHandler);

    try {
        std::cout << "🔧 初始化指标上报服务..." << std::endl;

        ElegantMetricsReporter reporter;
        reporter.run();

    } catch (const std::exception& e) {
        std::cerr << "💥 程序初始化异常: " << e.what() << std::endl;
        return 1;
    }

    return 0;
}

```

## 3. 了解更多

* 进行 <a href="#" target="_blank">指标检索</a>。

* 了解 <a href="#" target="_blank">怎么使用监控指标</a>。

* 了解如何 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/dashboard.md" target="_blank">配置仪表盘</a>。

* 了解如何使用 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/alarm-configurations/rules.md" target="_blank">监控告警</a>。