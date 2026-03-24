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

// 对应的头文件
#include "client.h"

// C++ 系统头文件
#include <string>
#include <thread>

// 第三方库
#include "opentelemetry/trace/span_id.h"
#include "opentelemetry/trace/semantic_conventions.h"
#include "oatpp/web/client/HttpRequestExecutor.hpp"
#include "oatpp/network/tcp/client/ConnectionProvider.hpp"

// 本地头文件
#include "otlp/logger_common.h"
#include "otlp/tracer_common.h"


namespace otlp_trace_api = opentelemetry::trace;
namespace otlp_context = opentelemetry::context;

int queryHelloWorld(std::string url) {
    const internal::Config &config = internal::Config::getInstance();
    auto logger = internal::getLogger(config.ServiceName);

    auto connectionProvider = oatpp::network::tcp::client::ConnectionProvider::createShared(
            {config.ServerAddress, static_cast<v_uint16>(config.ServerPort), oatpp::network::Address::IP_4
            });
    auto requestExecutor = oatpp::web::client::HttpRequestExecutor::createShared(connectionProvider);

    otlp_trace_api::StartSpanOptions options;
    options.kind = otlp_trace_api::SpanKind::kClient;

    std::string method = "GET";
    std::string scheme = "http";
    std::string path = "helloworld";
    auto span = internal::get_tracer(config.ServiceName)
            ->StartSpan("HTTP GET", {
                    {trace_api::SemanticConventions::kUrlFull,           url},
                    {trace_api::SemanticConventions::kUrlScheme,         scheme},
                    {trace_api::SemanticConventions::kNetworkProtocolName, scheme},
                    {trace_api::SemanticConventions::kHttpRequestMethod, method},
                    {trace_api::SemanticConventions::kNetworkTransport, "tcp"},
                    {trace_api::SemanticConventions::kServerPort, config.ServerPort},
                    {trace_api::SemanticConventions::kServerAddress, config.ServerAddress},
                    {trace_api::SemanticConventions::kNetworkPeerPort, config.ServerPort},
                    {trace_api::SemanticConventions::kNetworkPeerAddress, config.ServerAddress},
            }, options);

    auto scope = internal::get_tracer(config.ServiceName)->WithActiveSpan(span);

    auto current_ctx = otlp_context::RuntimeContext::GetCurrent();
    auto prop = opentelemetry::context::propagation::GlobalTextMapPropagator::GetGlobalPropagator();

    internal::HttpTextMapCarrier carrier;
    prop->Inject(carrier, current_ctx);

    logger->Info("[queryHelloWorld] send request");
    auto response = requestExecutor->executeOnce(method, path, carrier.headers_, nullptr,
                                                 requestExecutor->getConnection());

    int code = response->getStatusCode();
    auto body = response->readBodyToString().getValue("");
    span->SetAttribute(trace_api::SemanticConventions::kHttpResponseStatusCode, code);

    if (code != 200) {
        logger->Error("[queryHelloWorld] got error -> " + body);
        span->SetStatus(otlp_trace_api::StatusCode::kError);
        span->End();
        return code;
    }

    logger->Info("[queryHelloWorld] received: " + body);
    span->SetStatus(otlp_trace_api::StatusCode::kOk);
    span->End();

    return code;
}

void loopQueryHelloWorld(std::atomic<bool> &running) {
    const internal::Config &config = internal::Config::getInstance();

    std::string url = "http://" + config.ServerAddress + ":" + std::to_string(config.ServerPort) + "/helloworld";
    OATPP_LOGI("[http]", "start LoopQueryHelloWorld to periodically request %s", static_cast<const char *>(url.data()));

    while (running) {
        std::this_thread::sleep_for(std::chrono::seconds(3));

        auto span = internal::get_tracer(config.ServiceName)->StartSpan("Caller/queryHelloWorld");
        auto scope = internal::get_tracer(config.ServerAddress)->WithActiveSpan(span);

        try {
            queryHelloWorld(url);
        } catch (const std::exception& e) {
            OATPP_LOGE("[http]", "failed to execute queryHelloWorld: %s", e.what());
        }

        span->End();
    }

    OATPP_LOGI("[http]", "LoopQueryHelloWorld stopped");
}
