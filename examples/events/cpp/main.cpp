// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

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
            {"event", {{"content", "CPU 告警: " + std::to_string(dis(gen)) + "%"}}},
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
