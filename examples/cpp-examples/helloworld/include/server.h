// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// NOCC:build/header_guard(工具误报:)
#ifndef HELLOWORLD_INCLUDE_SERVER_H_
#define HELLOWORLD_INCLUDE_SERVER_H_

// C++ 系统头文件
#include <memory>
#include <random>
#include <stdexcept>
#include <string>
#include <utility>
#include <vector>

// 第三方库
#include <oatpp/web/server/HttpRequestHandler.hpp>
#include "opentelemetry/sdk/logs/logger.h"
#include "opentelemetry/sdk/metrics/meter.h"

// 本地头文件
#include "config.h"

namespace nostd         = opentelemetry::nostd;
namespace metrics_api   = opentelemetry::metrics;
namespace logs_api   = opentelemetry::logs;

// HelloWorldHelper class declaration
class HelloWorldHelper {
 private:
    std::vector<std::string> countries;
    std::vector<std::shared_ptr<std::runtime_error>> customErrors;

    nostd::shared_ptr<logs_api::Logger> logger;

    nostd::unique_ptr<metrics_api::Counter<uint64_t>> requestsTotal;
    nostd::unique_ptr<metrics_api::Histogram<double_t>> taskExecuteDurationSeconds;

 public:
    HelloWorldHelper();

    std::string choiceCountry();

    std::shared_ptr<std::runtime_error> choiceErr();

    static std::string generateGreeting(const std::string &country);

    std::shared_ptr<std::runtime_error> randErr(double errRate);

    // Logs（日志）打印日志
    void logsDemo(const std::shared_ptr<oatpp::web::server::HttpRequestHandler::IncomingRequest> &request);

    // Metrics（指标）- 使用 Counter 类型指标
    // Refer: https://opentelemetry.io/docs/languages/cpp/instrumentation/#create-a-counter
    void metricsCounterDemo(const std::string &country);

    // Metrics（指标）- 使用 Histogram 类型指标
    // Refer: https://opentelemetry.io/docs/languages/cpp/instrumentation/#create-a-histogram
    void metricsHistogramDemo();

    // Traces（调用链）- 自定义 Span
    // Refer: https://opentelemetry.io/docs/languages/cpp/instrumentation/#start-a-span
    static void tracesCustomSpanDemo();

    // Traces（调用链）- Span 事件
    // Refer:
    // https://opentelemetry-cpp.readthedocs.io/en/latest/otel_docs/classopentelemetry_1_1trace_1_1Span.html
    static void tracesSpanEventDemo();

    // Traces（调用链）- 异常事件、状态
    // Refer:
    // https://opentelemetry-cpp.readthedocs.io/en/latest/otel_docs/classopentelemetry_1_1trace_1_1Span.html
    std::shared_ptr<std::runtime_error> tracesRandomErrorDemo();
};

// Handler class declaration
class Handler : public oatpp::web::server::HttpRequestHandler {
 private:
    HelloWorldHelper helloWorldHelper;

 public:
    std::shared_ptr<OutgoingResponse> handle(const std::shared_ptr<IncomingRequest> &request) override;

    std::shared_ptr<OutgoingResponse> handleHelloWorld(
        const std::shared_ptr<HttpRequestHandler::IncomingRequest> &request);
};


// NOCC:build/header_guard(工具误报:)
#endif  // HELLOWORLD_INCLUDE_SERVER_H_
