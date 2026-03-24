// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// C++ 系统头文件
#include <csignal>
#include <string>

// 第三方库
#include "oatpp/web/server/HttpConnectionHandler.hpp"
#include "oatpp/network/Server.hpp"
#include "oatpp/network/tcp/server/ConnectionProvider.hpp"

// 本地头文件
#include "client.h"
#include "otlp/logger_common.h"
#include "otlp/meter_common.h"
#include "otlp/resource_common.h"
#include "otlp/tracer_common.h"
#include "server.h"

std::atomic<bool> running(true);
std::vector<std::thread> threads;

void StartServer() {
    threads.emplace_back([] {
        internal::Config &config = internal::Config::getInstance();
        // useExtendedConnections=true：开启以便通过连接获取 Client 信息
        auto connectionProvider = oatpp::network::tcp::server::ConnectionProvider::createShared(
                {config.ServerAddress, static_cast<v_uint16>(config.ServerPort), oatpp::network::Address::IP_4},
                true);

        auto router = oatpp::web::server::HttpRouter::createShared();
        router->route("GET", "/helloworld", std::make_shared<Handler>());
        auto connectionHandler = oatpp::web::server::HttpConnectionHandler::createShared(router);

        oatpp::network::Server server(connectionProvider, connectionHandler);
        OATPP_LOGI("[http]", "start to listen http server at %s:%s",
                   static_cast<const char *>(config.ServerAddress.data()),
                   static_cast<const char *>(connectionProvider->getProperty("port").getData()));

        std::function<bool()> condition = [](){
            return running.load();
        };
        server.run(condition);

        connectionProvider->stop();
        connectionHandler->stop();
        OATPP_LOGI("[http]", "service stopped");
    });

    OATPP_LOGI("[http]", "service started");
}

void StartQuerier() {
    threads.emplace_back([] {
        loopQueryHelloWorld(running);
    });
}

void StartOtlp(const internal::Config& config) {
    auto resource = internal::CreateResource(config);
    internal::initTracer(config, resource);
    internal::initMeter(config, resource);
    internal::initLogger(config, resource);
    OATPP_LOGI("[otlp]", "service started");
}

void Stop() {
    running.store(false);
    for (auto& t : threads) {
        if (t.joinable()) {
            t.join();
        }
    }
}

void StopOtlp(const internal::Config& config) {
    internal::cleanupTracer(config);
    internal::cleanupMeter(config);
    internal::cleanupLogger(config);
    OATPP_LOGI("[otlp]", "service stopped");
}

void signalHandler(int signal) {
    Stop();
}

int main() {
    oatpp::base::Environment::init();
    std::signal(SIGINT, signalHandler);

    const internal::Config &config = internal::Config::getInstance();
    StartOtlp(config);
    StartServer();
    StartQuerier();
    OATPP_LOGI("[main]", "🚀");

    while (running.load()) {
        std::this_thread::yield();
    }

    StopOtlp(config);
    OATPP_LOGI("[main]", "👋");

    oatpp::base::Environment::destroy();
    exit(0);
}
