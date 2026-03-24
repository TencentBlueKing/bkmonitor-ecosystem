// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//
// Created by sandrincai on 2024/9/8.
//

// NOCC:build/header_guard(工具误报:)
#ifndef HELLOWORLD_INCLUDE_OTLP_TRACER_COMMON_H_
#define HELLOWORLD_INCLUDE_OTLP_TRACER_COMMON_H_

// C++ 系统头文件
#include <grpcpp/grpcpp.h>
#include <cstring>
#include <iostream>
#include <vector>

// 第三方库
#include <oatpp/web/server/HttpRequestHandler.hpp>
#include "opentelemetry/context/propagation/global_propagator.h"
#include "opentelemetry/context/propagation/text_map_propagator.h"
#include "opentelemetry/exporters/ostream/span_exporter_factory.h"
#include "opentelemetry/exporters/otlp/otlp_grpc_exporter_factory.h"
#include "opentelemetry/exporters/otlp/otlp_http.h"
#include "opentelemetry/exporters/otlp/otlp_http_exporter_factory.h"
#include "opentelemetry/exporters/otlp/otlp_http_exporter_options.h"
#include "opentelemetry/nostd/shared_ptr.h"
#include "opentelemetry/sdk/resource/resource.h"
#include "opentelemetry/sdk/trace/batch_span_processor_factory.h"
#include "opentelemetry/sdk/trace/batch_span_processor_options.h"
#include "opentelemetry/sdk/trace/tracer_context.h"
#include "opentelemetry/sdk/trace/tracer_context_factory.h"
#include "opentelemetry/sdk/trace/tracer_provider_factory.h"
#include "opentelemetry/trace/propagation/http_trace_context.h"
#include "opentelemetry/trace/provider.h"

// 本地头文件
#include "config.h"

namespace nostd = opentelemetry::nostd;
namespace trace_api = opentelemetry::trace;
namespace trace_sdk = opentelemetry::sdk::trace;
namespace resource_sdk = opentelemetry::sdk::resource;
namespace otel_exporter = opentelemetry::exporter::otlp;

namespace internal {
class HttpTextMapCarrier : public opentelemetry::context::propagation::TextMapCarrier {
 public:
    explicit HttpTextMapCarrier(oatpp::web::protocol::http::Headers &headers) : headers_(headers) {}

    HttpTextMapCarrier() = default;

    virtual nostd::string_view Get(nostd::string_view key) const noexcept override {
        auto value = headers_.get(key.data());
        if (value) {
            return nostd::string_view(value->c_str(), value->size());
        }
        return "";
    }

    virtual void Set(nostd::string_view key, nostd::string_view value) noexcept override {
        headers_.put(oatpp::data::share::StringKeyLabelCI(oatpp::String(key.data(), key.size())),
                     oatpp::data::share::StringKeyLabel(oatpp::String(value.data(), value.size())));
    }

    oatpp::web::protocol::http::Headers headers_;
};

inline void initTracer(const Config &config, const resource_sdk::Resource &resource) {
    if (!config.EnableTraces) { return; }

    // 初始化 Exporter
    std::unique_ptr<trace_sdk::SpanExporter> exporter;
    if (config.OtlpExporterType == "grpc") {
        otel_exporter::OtlpGrpcExporterOptions otlpOptions;
        otlpOptions.endpoint = config.OtlpEndpoint;
        // ❗️❗【非常重要】请传入应用 Token
        otlpOptions.metadata.insert({"x-bk-token",  config.Token});
        exporter = otel_exporter::OtlpGrpcExporterFactory::Create(otlpOptions);
    } else {
        otel_exporter::OtlpHttpExporterOptions otlpOptions;
        otlpOptions.url = config.OtlpEndpoint + "/v1/traces";
        // ❗️❗【非常重要】请传入应用 Token
        otlpOptions.http_headers.insert({"x-bk-token", config.Token});
        exporter = otel_exporter::OtlpHttpExporterFactory::Create(otlpOptions);
    }

    // 初始化 Processors
    trace_sdk::BatchSpanProcessorOptions batchSpanProcessorOptions;
    auto processor = trace_sdk::BatchSpanProcessorFactory::Create(std::move(exporter), batchSpanProcessorOptions);
    std::vector<std::unique_ptr<trace_sdk::SpanProcessor>> processors;
    processors.push_back(std::move(processor));

    // 初始化 Provider
    auto context = trace_sdk::TracerContextFactory::Create(std::move(processors), resource);
    std::shared_ptr<trace_api::TracerProvider> provider = trace_sdk::TracerProviderFactory::Create(
            std::move(context));
    trace_api::Provider::SetTracerProvider(provider);

    // 初始化 Propagator
    opentelemetry::context::propagation::GlobalTextMapPropagator::SetGlobalPropagator(
            nostd::shared_ptr<opentelemetry::context::propagation::TextMapPropagator>(
                    new trace_api::propagation::HttpTraceContext()));
}

inline void cleanupTracer(const Config &config) {
    if (!config.EnableTraces) { return; }

    std::shared_ptr<trace_api::TracerProvider> none;
    trace_api::Provider::SetTracerProvider(none);
}

inline nostd::shared_ptr<trace_api::Tracer> get_tracer(const std::string &tracer_name) {
    auto provider = trace_api::Provider::GetTracerProvider();
    return provider->GetTracer(tracer_name);
}
}  // namespace internal

// NOCC:build/header_guard(工具误报:)
#endif  // HELLOWORLD_INCLUDE_OTLP_TRACER_COMMON_H_
