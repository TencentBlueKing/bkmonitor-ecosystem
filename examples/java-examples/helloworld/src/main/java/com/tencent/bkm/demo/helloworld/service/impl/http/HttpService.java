// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.helloworld.service.impl.http;

import com.sun.net.httpserver.HttpServer;
import com.tencent.bkm.demo.helloworld.service.Config;
import com.tencent.bkm.demo.helloworld.service.Service;

import java.net.InetSocketAddress;
import java.util.logging.Logger;

public class HttpService implements Service {

    private static final String PATH = "/helloworld";
    private static final Logger logger = Logger.getLogger(HttpService.class.getName());

    private HttpConfig config;
    private HttpServer server;

    @Override
    public String getType() {
        return "http";
    }

    @Override
    public void init(Config config) throws Exception {
        this.config = new HttpConfig(
                config.getServiceName(),
                config.getHttpPort(),
                config.getHttpAddress(),
                config.getHttpScheme()
        );
        this.server = HttpServer.create(new InetSocketAddress(this.config.getAddress(), this.config.getPort()), 0);
        this.server.createContext(PATH, new HelloWorldHttpHandler(this.config));
        this.server.setExecutor(null);
    }

    @Override
    public void start() throws Exception {
        if (server != null) {
            logger.info(
                    String.format(
                            "[%s] start to listen http server at %s:%d",
                            this.getType(), this.config.getAddress(), this.config.getPort()
                    )
            );
            server.start();
        }
    }

    @Override
    public void stop() throws Exception {
        if (server != null) {
            server.stop(0);
            logger.info(String.format("[%s] server stopped", this.getType()));
        }
    }
}
