// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//
// Created by sandrincai on 2024/9/9.
//

// NOCC:build/header_guard(工具误报:)
#ifndef HELLOWORLD_INCLUDE_OTLP_METER_COMMON_H_
#define HELLOWORLD_INCLUDE_OTLP_METER_COMMON_H_

// C++ 系统头文件
#include <memory>

// 第三方库
#include "opentelemetry/exporters/otlp/otlp_grpc_metric_exporter_factory.h"
#include "opentelemetry/exporters/otlp/otlp_http.h"
#include "opentelemetry/exporters/otlp/otlp_http_metric_exporter_factory.h"
#include "opentelemetry/exporters/otlp/otlp_http_metric_exporter_options.h"
#include "opentelemetry/metrics/provider.h"
#include "opentelemetry/sdk/metrics/export/periodic_exporting_metric_reader.h"
#include "opentelemetry/sdk/metrics/meter.h"
#include "opentelemetry/sdk/metrics/meter_context_factory.h"
#include "opentelemetry/sdk/metrics/meter_provider.h"
#include "opentelemetry/sdk/resource/resource.h"

// 本地头文件
#include "config.h"


namespace nostd = opentelemetry::nostd;
namespace common = opentelemetry::common;
namespace metrics_api = opentelemetry::metrics;
namespace metrics_sdk = opentelemetry::sdk::metrics;
namespace otlp_exporter = opentelemetry::exporter::otlp;
namespace resource_sdk = opentelemetry::sdk::resource;

namespace internal {
    inline void initMeter(const Config &config, const resource_sdk::Resource &resource) {
        if (!config.EnableMetrics) { return; }

        // 初始化 Exporter
        std::unique_ptr<metrics_sdk::PushMetricExporter> exporter;
        if (config.OtlpExporterType == "grpc") {
            otlp_exporter::OtlpGrpcMetricExporterOptions otlpOptions;
            otlpOptions.endpoint = config.OtlpEndpoint;
            // ❗️❗【非常重要】请传入应用 Token
            otlpOptions.metadata.insert({"x-bk-token",  config.Token});
            exporter = otlp_exporter::OtlpGrpcMetricExporterFactory::Create(otlpOptions);
        } else {
            otlp_exporter::OtlpHttpMetricExporterOptions otlpOptions;
            otlpOptions.url = config.OtlpEndpoint + "/v1/metrics";
            // ❗️❗【非常重要】请传入应用 Token
            otlpOptions.http_headers.insert({"x-bk-token", config.Token});
            exporter = otlp_exporter::OtlpHttpMetricExporterFactory::Create(otlpOptions);
        }

        // 初始化 Reader
        metrics_sdk::PeriodicExportingMetricReaderOptions options;
        std::unique_ptr<metrics_sdk::MetricReader> reader{
                new metrics_sdk::PeriodicExportingMetricReader(std::move(exporter), options)
        };

        // 初始化 Provider
        auto provider = std::shared_ptr<metrics_api::MeterProvider>(
                new metrics_sdk::MeterProvider(std::make_unique<metrics_sdk::ViewRegistry>(), resource));
        auto p = std::static_pointer_cast<metrics_sdk::MeterProvider>(provider);
        p->AddMetricReader(std::move(reader));

        // 初始化视图
        std::shared_ptr<metrics_sdk::HistogramAggregationConfig> aggConf(new metrics_sdk::HistogramAggregationConfig());
        aggConf->boundaries_ = {0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0};

        std::unique_ptr<metrics_sdk::View> view{new metrics_sdk::View(
                config.ServiceName, config.ServiceName, "seconds", metrics_sdk::AggregationType::kHistogram, aggConf)};
        std::unique_ptr<metrics_sdk::InstrumentSelector> instrument_selector{
                new metrics_sdk::InstrumentSelector(metrics_sdk::InstrumentType::kHistogram, "*", "seconds")};
        std::unique_ptr<metrics_sdk::MeterSelector> histogram_meter_selector{
                new metrics_sdk::MeterSelector(config.ServiceName, "version1", "schema1")};
        p->AddView(std::move(instrument_selector), std::move(histogram_meter_selector),
                   std::move(view));

        metrics_api::Provider::SetMeterProvider(provider);
    }

    inline void cleanupMeter(const Config &config) {
        if (!config.EnableMetrics) { return; }

        std::shared_ptr<metrics_api::MeterProvider> none;
        metrics_api::Provider::SetMeterProvider(none);
    }

    inline nostd::shared_ptr<metrics_api::Meter> get_meter(const std::string &meter_name) {
        auto provider = metrics_api::Provider::GetMeterProvider();
        return provider->GetMeter(meter_name, "version1", "schema1");
    }
}  // namespace internal

// NOCC:build/header_guard(工具误报:)
#endif  // HELLOWORLD_INCLUDE_OTLP_METER_COMMON_H_
