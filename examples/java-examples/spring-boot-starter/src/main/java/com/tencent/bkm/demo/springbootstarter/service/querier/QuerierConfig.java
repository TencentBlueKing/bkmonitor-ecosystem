// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.springbootstarter.service.querier;

import org.springframework.context.annotation.Configuration;
import org.springframework.core.env.Environment;
import org.springframework.stereotype.Component;

@Configuration
public class QuerierConfig {
    private final String serverAddress;
    private final String serverPort;
    private static final String schema = "http";
    private static final String QUERIER_PATH = "/travel";

    public QuerierConfig(Environment environment) {
        this.serverAddress = environment.getProperty("server.address");
        this.serverPort = environment.getProperty("server.port");
    }

    public String getEndpoint() {
        return schema + "://" + serverAddress + ":" + serverPort + QUERIER_PATH;
    }
}