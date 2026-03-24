// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// C++ 系统头文件
#include <cpr/cpr.h>
#include <atomic>
#include <chrono>
#include <csignal>
#include <cstdlib>
#include <iomanip>
#include <iostream>
#include <random>
#include <string>
#include <thread>

// 第三方库
#include <nlohmann/json.hpp>

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
        // ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 {{access_config.custom.http}} 」，
        // 其他环境、跨云场景请根据页面接入指引填写
        api_url = getEnv("API_URL", "");
        token = getEnv("TOKEN", "");  // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
        // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
        std::string data_id_str = getEnv("DATA_ID", "");
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
                cpr::Timeout{std::chrono::seconds(10)});

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
