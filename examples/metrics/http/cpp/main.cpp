// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

#include <cstdlib>
#include <cpr/cpr.h>
#include <nlohmann/json.hpp>
#include <chrono>
#include <iomanip>
#include <iostream>
#include <random>
#include <string>
#include <thread>

using json = nlohmann::json;

// Helper: get environment variable with optional default value
std::string getEnv(const std::string& key, const std::string& defaultValue = "") {
    const char* value = std::getenv(key.c_str());
    return value ? std::string(value) : defaultValue;
}

// Helper: get current timestamp string
std::string getCurrentTimestamp() {
    auto now = std::chrono::system_clock::now();
    auto time_t = std::chrono::system_clock::to_time_t(now);
    auto tm = *std::localtime(&time_t);
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%d %H:%M:%S");
    return oss.str();
}

// Build JSON payload for metrics report
std::string buildJsonPayload(int data_id, const std::string& token,
                             double cpu_load, double mem_usage) {
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

// Send metrics report to API, returns HTTP status code
int sendReport(const std::string& api_url, const std::string& token,
               int data_id, double cpu_load, double mem_usage) {
    cpr::Response response = cpr::Post(
        cpr::Url{api_url},
        cpr::Header{{"Content-Type", "application/json"}},
        cpr::Body{buildJsonPayload(data_id, token, cpu_load, mem_usage)},
        cpr::Timeout{std::chrono::seconds(10)});
    return response.status_code;
}

int main() {
    // ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 {{access_config.custom.http}} 」，
    // 其他环境、跨云场景请根据页面接入指引填写
    std::string api_url = getEnv("API_URL");
    // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
    std::string token = getEnv("TOKEN");
    // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
    std::string data_id_str = getEnv("DATA_ID");
    // 上报间隔，默认为60秒
    int interval = std::stoi(getEnv("INTERVAL", "60"));
    int data_id = std::stoi(data_id_str);

    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_real_distribution<double> dist(0.0, 100.0);

    std::cout << "🚀 启动指标上报服务" << std::endl;
    std::cout << "API地址: " << api_url << std::endl;
    std::cout << "数据ID: " << data_id << std::endl;
    std::cout << "上报间隔: " << interval << "秒" << std::endl;
    std::cout << "=================================" << std::endl;

    while (true) {
        double cpu_load = dist(gen);
        double mem_usage = dist(gen);
        int status = sendReport(api_url, token, data_id, cpu_load, mem_usage);
        std::string timestamp = getCurrentTimestamp();
        if (status == 200) {
            std::cout << "[" << timestamp << "] ✅ 上报成功 | "
                      << "CPU: " << std::fixed << std::setprecision(2) << cpu_load << "% "
                      << "内存: " << mem_usage << "%" << std::endl;
        } else {
            std::cout << "[" << timestamp << "] ❌ 上报失败 | "
                      << "状态码: " << status << std::endl;
        }
        std::this_thread::sleep_for(std::chrono::seconds(interval));
    }
    return 0;
}
