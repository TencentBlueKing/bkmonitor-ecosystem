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

// 对应的头文件
#include "server.h"

// C++ 系统头文件
#include <chrono>
#include <iostream>

// 第三方库
#include "opentelemetry/baggage/baggage.h"
#include "opentelemetry/trace/semantic_conventions.h"
#include "opentelemetry/trace/span_metadata.h"
#include "opentelemetry/trace/span_startoptions.h"

// 本地头文件
#include "otlp/logger_common.h"
#include "otlp/tracer_common.h"
#include "otlp/meter_common.h"

namespace otlp_trace_api = opentelemetry::trace;
namespace otlp_context = opentelemetry::context;
namespace otlp_common = opentelemetry::common;

using oatpp::web::server::HttpRequestHandler;
using oatpp::web::protocol::http::Status;
using oatpp::web::protocol::http::outgoing::ResponseFactory;

// Constructor to initialize the countries and customErrors
HelloWorldHelper::HelloWorldHelper() {
    countries = {
            "United States", "Canada", "United Kingdom", "Germany", "France",
            "Japan", "Australia", "China", "India", "Brazil"
    };

    customErrors = {
            std::make_shared<std::runtime_error>("mysql connect timeout"),
            std::make_shared<std::runtime_error>("user not found"),
            std::make_shared<std::runtime_error>("network unreachable"),
            std::make_shared<std::runtime_error>("file not found")
    };

    const internal::Config &config = internal::Config::getInstance();

    auto meter = internal::get_meter(config.ServiceName);
    requestsTotal = meter->CreateUInt64Counter("requests_total", "Total number of HTTP requests");
    taskExecuteDurationSeconds = meter->CreateDoubleHistogram(
        "task_execute_duration_seconds",
        "Task execute duration in seconds");

    logger = internal::getLogger(config.ServiceName);
}

void doSomething(int maxMs) {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(10, maxMs + 9);
    int r = dis(gen);
    std::this_thread::sleep_for(std::chrono::milliseconds(r));
}

std::string HelloWorldHelper::choiceCountry() {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> distr(0, countries.size() - 1);
    return countries[distr(gen)];
}

std::shared_ptr<std::runtime_error> HelloWorldHelper::choiceErr() {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> distr(0, customErrors.size() - 1);
    return customErrors[distr(gen)];
}

std::string HelloWorldHelper::generateGreeting(const std::string &country) {
    return "Hello World, " + country + "!";
}

std::shared_ptr<std::runtime_error> HelloWorldHelper::randErr(double errRate) {
    if (static_cast<double>(rand()) / RAND_MAX < errRate) {
        return choiceErr();
    }
    return nullptr;
}

void HelloWorldHelper::logsDemo(
    const std::shared_ptr<oatpp::web::server::HttpRequestHandler::IncomingRequest> &request) {
    std::string url = request->getPathTail().getValue("");
    std::string method = request->getStartingLine().method.toString();
    logger->Info("received request: " + method + " " + url);
}

void HelloWorldHelper::metricsCounterDemo(const std::string &country) {
    requestsTotal->Add(1, {{"country", country}});
}

void HelloWorldHelper::metricsHistogramDemo() {
    auto begin = std::chrono::high_resolution_clock::now();
    doSomething(100);
    auto end = std::chrono::high_resolution_clock::now();

    std::chrono::duration<double> duration = end - begin;
    taskExecuteDurationSeconds->Record(duration.count(), {});
}

void HelloWorldHelper::tracesCustomSpanDemo() {
    const internal::Config &config = internal::Config::getInstance();
    auto span = internal::get_tracer(config.ServiceName)
            ->StartSpan("CustomSpanDemo/doSomething");
    auto scope = internal::get_tracer(config.ServiceName)->WithActiveSpan(span);

    doSomething(50);

    // 增加 Span 自定义属性
    span->SetAttribute("helloworld.kind", 1);
    span->SetAttribute("helloworld.step", "tracesCustomSpanDemo");
    span->End();
}

void HelloWorldHelper::tracesSpanEventDemo() {
    const internal::Config &config = internal::Config::getInstance();
    auto span = internal::get_tracer(config.ServiceName)
            ->StartSpan("SpanEventDemo/doSomething");
    auto scope = internal::get_tracer(config.ServiceName)->WithActiveSpan(span);

    std::map<std::string, otlp_common::AttributeValue> attributes = {
            {"helloworld.kind", 2}, {"helloworld.step", "tracesCustomSpanDemo"}
    };

    span->AddEvent("Before doSomething", attributes);
    doSomething(50);
    span->AddEvent("After doSomething", attributes);
    span->End();
}


std::shared_ptr<std::runtime_error> HelloWorldHelper::tracesRandomErrorDemo() {
    if (auto err = randErr(0.1)) {
        auto ctx = opentelemetry::context::RuntimeContext::GetCurrent();
        auto span = trace_api::GetSpan(ctx);

        auto exceptionMessage = err->what();
        auto exceptionType = typeid(err).name();
        logger->Error("[tracesRandomErrorDemo] got error -> " + std::string(exceptionMessage));
        span->SetStatus(trace_api::StatusCode::kError, exceptionMessage);
        span->AddEvent("exception", {
            {trace_api::SemanticConventions::kExceptionMessage, exceptionMessage},
            {trace_api::SemanticConventions::kExceptionType, exceptionType}
        });
        return err;
    }
    return nullptr;
}


std::shared_ptr<HttpRequestHandler::OutgoingResponse>
Handler::handle(const std::shared_ptr<HttpRequestHandler::IncomingRequest> &request) {
    const internal::Config &config = internal::Config::getInstance();

    std::string path = request->getStartingLine().path.toString();
    std::string url = request->getPathTail().getValue("");
    std::string method = request->getStartingLine().method.toString();
    std::string protocol = request->getStartingLine().protocol.toString();
    std::string userAgent = request->getHeader("user-agent").getValue("");
    std::string clientAddress = request->getHeader("x-forwarded-for").getValue("");
    int clientPort = std::stoi(request->getHeader("x-forwarded-port").getValue("0"));
    const oatpp::data::stream::Context::Properties connectionProperties
            = request->getConnection()->getInputStreamContext().getProperties();
    if (clientAddress.empty()) {
        clientAddress = connectionProperties.get("peer_address").getValue("");
    }
    if (clientPort == 0) {
        clientPort = std::stoi(connectionProperties.get("peer_port").getValue("0"));
    }

    oatpp::web::protocol::http::Headers headers = request->getHeaders();
    internal::HttpTextMapCarrier carrier(headers);

    auto prop = otlp_context::propagation::GlobalTextMapPropagator::GetGlobalPropagator();
    auto currentCtx = otlp_context::RuntimeContext::GetCurrent();
    auto newCtx = prop->Extract(carrier, currentCtx);

    otlp_trace_api::StartSpanOptions startSpanOptions;
    startSpanOptions.kind = otlp_trace_api::SpanKind::kServer;
    startSpanOptions.parent = otlp_trace_api::GetSpan(newCtx)->GetContext();

    auto span = internal::get_tracer(config.ServiceName)
            ->StartSpan("HTTP Server", {
                    {trace_api::SemanticConventions::kNetworkProtocolName, protocol},
                    {trace_api::SemanticConventions::kServerAddress,       config.ServerAddress},
                    {trace_api::SemanticConventions::kServerPort,          config.ServerPort},
                    {trace_api::SemanticConventions::kNetworkLocalAddress, config.ServerAddress},
                    {trace_api::SemanticConventions::kNetworkLocalPort,    config.ServerPort},
                    {trace_api::SemanticConventions::kNetworkTransport,    "tcp"},
                    {trace_api::SemanticConventions::kHttpRequestMethod,   method},
                    {trace_api::SemanticConventions::kUrlScheme,           "http"},
                    {trace_api::SemanticConventions::kUrlPath,             path},
                    {trace_api::SemanticConventions::kUrlQuery,            url},
                    {trace_api::SemanticConventions::kClientAddress,       clientAddress},
                    {trace_api::SemanticConventions::kClientPort,          clientPort},
                    {trace_api::SemanticConventions::kNetworkPeerAddress,  clientAddress},
                    {trace_api::SemanticConventions::kNetworkPeerPort,     clientPort},
                    {trace_api::SemanticConventions::kUserAgentOriginal,   userAgent},
            }, startSpanOptions);
    auto scope = internal::get_tracer(config.ServiceName)->WithActiveSpan(span);

    auto response = handleHelloWorld(request);

    if (response->getStatus() == Status::CODE_200) {
        span->SetStatus(otlp_trace_api::StatusCode::kOk);
    } else {
        span->SetStatus(otlp_trace_api::StatusCode::kError);
    }

    span->SetAttribute(trace_api::SemanticConventions::kHttpResponseStatusCode, response->getStatus().code);
    span->End();

    return response;
}


std::shared_ptr<HttpRequestHandler::OutgoingResponse>
Handler::handleHelloWorld(const std::shared_ptr<HttpRequestHandler::IncomingRequest> &request) {
    const internal::Config &config = internal::Config::getInstance();
    auto logger = internal::getLogger(config.ServiceName);

    auto span = internal::get_tracer(config.ServiceName)->StartSpan("Handle/HelloWorld");
    auto scope = internal::get_tracer(config.ServiceName)->WithActiveSpan(span);

    // Logs（日志）
    helloWorldHelper.logsDemo(request);

    auto country = helloWorldHelper.choiceCountry();
    logger->Info("get country -> " + country);

    // Metrics（指标） - Counter 类型
    helloWorldHelper.metricsCounterDemo(country);
    // Metrics（指标） - Histograms 类型
    helloWorldHelper.metricsHistogramDemo();

    // Traces（调用链）- 自定义 Span
    HelloWorldHelper::tracesCustomSpanDemo();
    // Traces（调用链）- Span 事件
    HelloWorldHelper::tracesSpanEventDemo();

    // Traces（调用链）- 模拟错误
    if (auto err = helloWorldHelper.tracesRandomErrorDemo()) {
        auto response = ResponseFactory::createResponse(Status::CODE_500, err->what());

        span->End();
        return response;
    }

    auto greeting = HelloWorldHelper::generateGreeting(country);
    auto response = ResponseFactory::createResponse(Status::CODE_200, greeting.c_str());

    span->End();
    return response;
}