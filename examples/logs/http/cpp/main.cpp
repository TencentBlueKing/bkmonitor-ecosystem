// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// C 系统文件
#include <csignal>
#include <cstdlib>
#include <ctime>

// C++ 系统文件
#include <atomic>
#include <chrono>
#include <iostream>
#include <random>
#include <string>
#include <thread>
#include <vector>

// 其他库文件
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
    // ❗❗【非常重要】上报地址，国内站点默认是「 {{access_config.otlp.http_endpoint}}/v1/logs 」，
    // 其他环境、跨云场景请根据页面接入指引填写
    std::string api_url = get_env("API_URL",
                                  "{{access_config.otlp.http_endpoint}}/v1/logs");

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
        // curl API 要求 CURLINFO_RESPONSE_CODE 必须传 long*，故此处保留 long 而非改用 int64_t
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
