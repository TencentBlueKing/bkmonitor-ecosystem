// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// NOCC:build/header_guard(工具误报:)
#ifndef HELLOWORLD_INCLUDE_OTLP_LOGGER_COMMON_H_
#define HELLOWORLD_INCLUDE_OTLP_LOGGER_COMMON_H_

// 第三方库
#include "opentelemetry/exporters/otlp/otlp_grpc_exporter_factory.h"
#include "opentelemetry/exporters/otlp/otlp_grpc_log_record_exporter_factory.h"
#include "opentelemetry/exporters/otlp/otlp_http.h"
#include "opentelemetry/exporters/otlp/otlp_http_log_record_exporter_factory.h"
#include "opentelemetry/exporters/otlp/otlp_http_log_record_exporter_options.h"
#include "opentelemetry/logs/provider.h"
#include "opentelemetry/sdk/logs/logger.h"
#include "opentelemetry/sdk/logs/logger_context_factory.h"
#include "opentelemetry/sdk/logs/logger_provider_factory.h"
#include "opentelemetry/sdk/logs/simple_log_record_processor_factory.h"
#include "opentelemetry/sdk/resource/resource.h"

// 本地头文件
#include "config.h"

namespace nostd = opentelemetry::nostd;
namespace otlp = opentelemetry::exporter::otlp;
namespace logs_api = opentelemetry::logs;
namespace logs_sdk = opentelemetry::sdk::logs;
namespace resource_sdk = opentelemetry::sdk::resource;

namespace internal {
    inline void initLogger(const Config &config, const resource_sdk::Resource &resource) {
        if (!config.EnableLogs) { return; }


        // 初始化 Exporter
        std::unique_ptr<logs_sdk::LogRecordExporter> exporter;
        if (config.OtlpExporterType == "grpc") {
            otlp::OtlpGrpcLogRecordExporterOptions loggerOptions;
            loggerOptions.endpoint = config.OtlpEndpoint;
            // ❗️❗【非常重要】请传入应用 Token
            loggerOptions.metadata.insert({"x-bk-token",  config.Token});
            exporter = otlp::OtlpGrpcLogRecordExporterFactory::Create(loggerOptions);
        } else {
            otlp::OtlpHttpLogRecordExporterOptions loggerOptions;
            loggerOptions.url = config.OtlpEndpoint + "/v1/logs";
            // ❗️❗【非常重要】请传入应用 Token
            loggerOptions.http_headers.insert({"x-bk-token", config.Token});
            exporter = otlp::OtlpHttpLogRecordExporterFactory::Create(loggerOptions);
        }

        // 初始化 Processors
        auto processor = logs_sdk::SimpleLogRecordProcessorFactory::Create(std::move(exporter));
        std::vector<std::unique_ptr<logs_sdk::LogRecordProcessor>> processors;
        processors.push_back(std::move(processor));

        // 初始化 Provider
        auto context = logs_sdk::LoggerContextFactory::Create(std::move(processors), resource);
        std::shared_ptr<logs_api::LoggerProvider> provider = logs_sdk::LoggerProviderFactory::Create(
                std::move(context));
        logs_api::Provider::SetLoggerProvider(provider);
    }

    inline nostd::shared_ptr<opentelemetry::logs::Logger> getLogger(const std::string &name) {
        auto provider = logs_api::Provider::GetLoggerProvider();
        return provider->GetLogger(name + "_logger", name, OPENTELEMETRY_SDK_VERSION);
    }

    inline void cleanupLogger(const Config &config) {
        if (!config.EnableLogs) { return; }

        std::shared_ptr<logs_api::LoggerProvider> none;
        logs_api::Provider::SetLoggerProvider(none);
    }
}  // namespace internal

// NOCC:build/header_guard(工具误报:)
#endif  // HELLOWORLD_INCLUDE_OTLP_LOGGER_COMMON_H_
