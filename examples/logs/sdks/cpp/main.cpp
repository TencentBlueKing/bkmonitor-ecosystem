// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// 该示例演示使用 OpenTelemetry C++ SDK 通过 OTLP/HTTP 上报日志到蓝鲸监控平台。
#include <atomic>
#include <iostream>
#include <string>
#include <csignal>
#include <cstdlib>
#include <memory>
#include <chrono>
#include <thread>
#include <ctime>

#include "opentelemetry/logs/provider.h"
#include "opentelemetry/logs/severity.h"
#include "opentelemetry/sdk/logs/logger_provider.h"
#include "opentelemetry/sdk/logs/batch_log_record_processor.h"
#include "opentelemetry/sdk/resource/resource.h"
#include "opentelemetry/exporters/otlp/otlp_http_log_record_exporter.h"
#include "opentelemetry/exporters/otlp/otlp_http_log_record_exporter_options.h"

namespace logs_api      = opentelemetry::logs;
namespace logs_sdk      = opentelemetry::sdk::logs;
namespace otlp_exporter = opentelemetry::exporter::otlp;
namespace nostd         = opentelemetry::nostd;

std::atomic<bool> running{true};

void SignalHandler(int) { running = false; }

struct LogLevel { logs_api::Severity severity; const char* text; };
const LogLevel kLevels[] = {
    {logs_api::Severity::kInfo, "INFO"},  {logs_api::Severity::kWarn, "WARN"},
    {logs_api::Severity::kError, "ERROR"}, {logs_api::Severity::kDebug, "DEBUG"},
};

std::string CurrentTimeStr() {
    auto now = std::chrono::system_clock::now();
    auto tt = std::chrono::system_clock::to_time_t(now);
    auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(now.time_since_epoch()) % 1000;
    char buf[64];
    std::strftime(buf, sizeof(buf), "%Y-%m-%d %H:%M:%S", std::localtime(&tt));
    return std::string(buf) + "," + std::to_string(ms.count());
}

std::string GetEnv(const char* key, const std::string& fallback) {
    const char* val = std::getenv(key);
    return val ? std::string(val) : fallback;
}

int main() {
    std::signal(SIGTERM, SignalHandler);
    std::signal(SIGINT,  SignalHandler);

    // ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
    std::string token   = GetEnv("TOKEN", "fixme");
    // ❗❗【非常重要】上报地址，国内站点默认是「 {{access_config.otlp.http_endpoint}}/v1/logs 」，
    // 其他环境、跨云场景请根据页面接入指引填写。
    std::string api_url = GetEnv("API_URL", "{{access_config.otlp.http_endpoint}}/v1/logs");

    // 配置 OTLP HTTP 日志导出器
    otlp_exporter::OtlpHttpLogRecordExporterOptions opts;
    opts.url            = api_url;
    opts.content_type   = otlp_exporter::HttpRequestContentType::kBinary;
    opts.http_headers.insert({"x-bk-token", token});

    auto exporter = std::unique_ptr<logs_sdk::LogRecordExporter>(
        new otlp_exporter::OtlpHttpLogRecordExporter(opts));

    // 创建批量处理器
    auto processor = std::unique_ptr<logs_sdk::LogRecordProcessor>(
        new logs_sdk::BatchLogRecordProcessor(std::move(exporter)));

    // 声明资源属性，用于在平台侧区分日志来源
    auto resource = opentelemetry::sdk::resource::Resource::Create({
        {"service.name", "custom-log-sdk-demo"},
        {"deployment.environment.name", "local"},
    });

    // 创建 LoggerProvider
    auto provider = nostd::shared_ptr<logs_sdk::LoggerProvider>(
        new logs_sdk::LoggerProvider(std::move(processor), resource));

    nostd::shared_ptr<logs_sdk::LoggerProvider> g_provider = provider;

    // 设置为全局 LoggerProvider（必须使用 std::move 转为右值）
    logs_api::Provider::SetLoggerProvider(std::move(provider));

    // logger_name, library_name, library_version
    auto logger = g_provider->GetLogger("otel", "otel", "1.0.0");

    std::cout << CurrentTimeStr() << " - INFO - Starting log reporter (press Ctrl+C/SIGTERM to stop)..." << std::endl;

    int count = 0;
    while (running) {
        auto record = logger->CreateLogRecord();
        const auto& level = kLevels[count % 4];

        // 显式设置事件发生时间，避免上报的时间戳为 0
        auto now = std::chrono::system_clock::now();
        record->SetTimestamp(now);
        record->SetSeverity(level.severity);
        record->SetBody("log from cpp sdk");
        logger->EmitLogRecord(std::move(record));

        std::cout << CurrentTimeStr() << " - " << level.text << " - log from cpp sdk" << std::endl;

        count++;
        std::this_thread::sleep_for(std::chrono::seconds(1));
    }

    std::cout << CurrentTimeStr() << " - INFO - shutting down..." << std::endl;
    g_provider->Shutdown();
    return 0;
}
