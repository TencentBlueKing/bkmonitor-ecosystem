// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//
// Created by sandrincai on 2024/9/10.
//

// NOCC:build/header_guard(工具误报:)
#ifndef HELLOWORLD_INCLUDE_CONFIG_H_
#define HELLOWORLD_INCLUDE_CONFIG_H_

// C++ 系统头文件
#include <memory>
#include <random>
#include <stdexcept>
#include <string>
#include <utility>
#include <vector>

// 第三方库
#include "oatpp/web/server/HttpRequestHandler.hpp"

namespace internal {
class Config {
 public:
    // 维护单例
    Config(const Config &) = delete;

    Config &operator=(const Config &) = delete;

    static Config &getInstance() {
        static Config instance;
        return instance;
    }

    bool Debug;
    std::string Token;
    std::string ServiceName;
    std::string OtlpEndpoint;
    std::string OtlpExporterType;
    bool EnableLogs;
    bool EnableTraces;
    bool EnableMetrics;
    bool EnableProfiling;
    std::string ProfilingLocalEndpoint;

    int ServerPort;
    std::string ServerAddress;

 private:
    Config() {
        Debug = getEnvAsBool("DEBUG", true);
        Token = getEnv("TOKEN", "todo");
        ServiceName = getEnv("SERVICE_NAME", "helloworld");
        OtlpEndpoint = getEnv("OTLP_ENDPOINT", "localhost:4317");
        OtlpExporterType = getEnv("OTLP_EXPORTER_TYPE", "grpc");
        EnableLogs = getEnvAsBool("ENABLE_LOGS", true);
        EnableTraces = getEnvAsBool("ENABLE_TRACES", true);
        EnableMetrics = getEnvAsBool("ENABLE_METRICS", true);
        EnableProfiling = getEnvAsBool("ENABLE_PROFILING", false);
        ServerAddress = getEnv("SERVER_ADDRESS", "localhost");
        ProfilingLocalEndpoint = "http://localhost:4040";
        ServerPort = 8080;
    }

    static std::string getEnv(const std::string &key, const std::string &defaultValue) {
        const char *val = std::getenv(key.c_str());
        if (val) {
            return {val};
        }
        return defaultValue;
    }

    static bool getEnvAsBool(const std::string &key, bool defaultValue) {
        const char *val = std::getenv(key.c_str());
        if (val) {
            std::string strVal(val);
            if (strVal == "true" || strVal == "1") {
                return true;
            } else if (strVal == "false" || strVal == "0") {
                return false;
            }
        }
        return defaultValue;
    }
};
}  // namespace internal

// NOCC:build/header_guard(工具误报:)
#endif  // HELLOWORLD_INCLUDE_CONFIG_H_
