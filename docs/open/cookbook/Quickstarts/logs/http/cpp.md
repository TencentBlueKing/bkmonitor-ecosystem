# C++-日志（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/http/README.md" target="_blank">自定义日志 HTTP 上报</a>

* <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs（OTel 日志）</a>：OpenTelemetry 中用于描述离散日志事件的信号类型。

* <a href="https://opentelemetry.io/docs/specs/otel/logs/data-model/" target="_blank">Logs Data Model（OTel 日志数据模型）</a>：定义 `resourceLogs`、`scopeLogs`、`logRecords`、`body`、`attributes`、`severityNumber` 等字段含义。

* <a href="https://opentelemetry.io/docs/specs/otlp/#otlphttp" target="_blank">OTLP/HTTP（OpenTelemetry HTTP 上报协议）</a>：定义通过 HTTP 上报 OTel 数据的协议方式，本示例使用 `/v1/logs` 上报日志。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/logs/http/cpp
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/http/README.md" target="_blank">自定义日志 HTTP 上报</a> 创建一个上报协议为 `JSON` 的自定义日志，关注创建后提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。

* `API_URL`：国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，其他环境、跨云场景请根据页面接入指引填写。

### 2.2 样例运行参数

运行参数说明：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `API_URL` | 是 | ❗❗【非常重要】日志上报接口地址（`Access URL`），请根据页面接入指引填写；如果页面提供的是 OTLP HTTP Endpoint 根地址，请在末尾追加 `/v1/logs`。 |
| `TOKEN` | 是 | ❗❗【非常重要】日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。 |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/master/examples/logs/http/cpp" target="_blank">bkmonitor-ecosystem/examples/logs/http/cpp</a> 中找到。

通过 docker build 构建名为 logs-http-cpp 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、API_URL 传递配置参数，实现周期上报日志：

```bash
docker build -t logs-http-cpp .

docker run -e TOKEN="fixme" \
 -e API_URL="http://127.0.0.1:4318/v1/logs" \
 logs-http-cpp
```

运行输出：

```bash
2026-08-11 09:04:46 - INFO - Starting log reporter (press Ctrl+C to stop)...
2026-08-11 09:04:46 - INFO - Sending log level: WARN (13)
2026-08-11 09:04:46 - INFO - response.status_code=200, body={}
2026-08-11 09:04:46 - INFO - Sending log level: DEBUG (5)
2026-08-11 09:04:46 - INFO - response.status_code=200, body={}
2026-08-11 09:04:46 - INFO - Sending log level: ERROR (17)
2026-08-11 09:04:46 - INFO - response.status_code=200, body={}
2026-08-11 09:04:46 - INFO - Sending log level: INFO (9)
2026-08-11 09:04:46 - INFO - response.status_code=200, body={}
...
```

### 2.4 样例代码

上报代码示例：

```cpp
#include <atomic>
#include <chrono>
#include <csignal>
#include <cstdlib>
#include <ctime>
#include <iostream>
#include <random>
#include <string>
#include <thread>
#include <vector>

#include <curl/curl.h>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

static std::atomic<bool> g_running{true};


// ==================== 工具函数 ====================
std::string get_env(const char* key, const std::string& fallback = "") {
    const char* val = std::getenv(key);
    return val ? val : fallback;
}

std::string current_timestamp() {
    auto now = std::chrono::system_clock::now();
    std::time_t t = std::chrono::system_clock::to_time_t(now);
    char buf[32];
    std::strftime(buf, sizeof(buf), "%Y-%m-%d %H:%M:%S", std::localtime(&t));
    return buf;
}

// ==================== 日志工具 ====================
void log_info(const std::string& msg) {
    std::cout << current_timestamp() << " - INFO - " << msg << std::endl;
}

void log_error(const std::string& msg) {
    std::cerr << current_timestamp() << " - ERROR - " << msg << std::endl;
}

// ==================== 时间戳 ====================
std::string get_current_nano_timestamp() {
    auto now = std::chrono::system_clock::now();
    auto ns = std::chrono::duration_cast<std::chrono::nanoseconds>(
        now.time_since_epoch()).count();
    return std::to_string(ns);
}

// ==================== 随机日志级别 ====================
struct LogLevel {
    int severity_number;
    std::string severity_text;
    std::string message;
};

LogLevel get_random_level() {
    static const std::vector<LogLevel> levels = {
        {5,  "DEBUG", "debug log from cpp http"},
        {9,  "INFO",  "info log from cpp http"},
        {13, "WARN",  "warn log from cpp http"},
        {17, "ERROR", "error log from cpp http"},
    };
    static std::random_device rd;
    static std::mt19937 gen(rd());
    static std::uniform_int_distribution<size_t> dist(0, levels.size() - 1);
    return levels[dist(gen)];
}

// ==================== 构建 OTLP payload ====================
json build_payload() {
    std::string now_ns = get_current_nano_timestamp();
    LogLevel level = get_random_level();

    return {
        {"resourceLogs", {
            {
                {"resource", {
                    {"attributes", {
                        {{"key", "service.name"},
                         {"value", {{"stringValue", "custom-log-demo"}}}},
                        {{"key", "deployment.environment.name"},
                         {"value", {{"stringValue", "local"}}}}
                    }}
                }},
                {"scopeLogs", {
                    {
                        {"scope", {{"name", "cpp-http-demo"}}},
                        {"logRecords", {
                            {
                                {"timeUnixNano", now_ns},
                                {"observedTimeUnixNano", now_ns},
                                {"severityNumber", level.severity_number},
                                {"severityText", level.severity_text},
                                {"body", {{"stringValue", level.message}}},
                                {"attributes", {
                                    {{"key", "demo.source"},
                                     {"value", {{"stringValue", "cpp"}}}}
                                }}
                            }
                        }}
                    }
                }}
            }
        }}
    };
}

// ==================== HTTP POST ====================
static size_t write_cb(void* data, size_t size, size_t nmemb, std::string* out) {
    out->append(static_cast<char*>(data), size * nmemb);
    return size * nmemb;
}



void do_post(const json& payload) {
    // ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
    std::string token = get_env("TOKEN", "fixme");
    // ❗❗【非常重要】上报地址，国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，
    // 其他环境、跨云场景请根据页面接入指引填写
    std::string api_url = get_env("API_URL",
                                  "http://127.0.0.1:4318/v1/logs");

    auto rec = payload["resourceLogs"][0]["scopeLogs"][0]["logRecords"][0];
    std::string msg = "Sending log level: " + rec["severityText"].get<std::string>() +
                      " (" + std::to_string(rec["severityNumber"].get<int>()) + ")";
    log_info(msg);

    std::string body = payload.dump();

    CURL* curl = curl_easy_init();
    if (!curl) {
        log_error("curl init failed");
        return;
    }

    struct curl_slist* hdr = nullptr;
    hdr = curl_slist_append(hdr, "Content-Type: application/json");
    std::string auth = "x-bk-token: " + token;
    hdr = curl_slist_append(hdr, auth.c_str());

    std::string resp;
    curl_easy_setopt(curl, CURLOPT_URL, api_url.c_str());
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, body.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, hdr);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_cb);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &resp);

    CURLcode rc = curl_easy_perform(curl);
    if (rc != CURLE_OK) {
        log_error(std::string("failed to post request: ") + curl_easy_strerror(rc));
    } else {
        long code;
        curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &code);
        log_info("response.status_code=" + std::to_string(code) + ", body=" + resp);
    }

    curl_slist_free_all(hdr);
    curl_easy_cleanup(curl);
}


// ==================== 信号处理 ====================
void signal_handler(int) {
    log_info("Received keyboard interrupt, exiting...");
    g_running = false;
}

// ==================== main ====================
int main() {
    curl_global_init(CURL_GLOBAL_ALL);

    std::signal(SIGINT, signal_handler);
    std::signal(SIGTERM, signal_handler);

    log_info("Starting log reporter (press Ctrl+C to stop)...");

    while (g_running) {
        json payload = build_payload();
        do_post(payload);
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }

    curl_global_cleanup();
    return 0;
}

```

## 3. 了解更多

进一步了解以下内容：

* 进行 <a href="#" target="_blank">日志检索</a>。

* 了解 <a href="#" target="_blank">容器日志自定义上报使用文档</a>。

* 了解 <a href="#" target="_blank">容器日志采集器安装</a>。