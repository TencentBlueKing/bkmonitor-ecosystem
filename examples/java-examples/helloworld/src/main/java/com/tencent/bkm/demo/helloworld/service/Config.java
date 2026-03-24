// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.helloworld.service;

import lombok.Getter;

import java.util.Optional;


@Getter
public class Config {
    private final boolean debug;
    private final String token;
    private final String serviceName;
    private final String otlpEndpoint;
    private final String otlpExporterType;
    private final boolean enableLogs;
    private final boolean enableTraces;
    private final boolean enableMetrics;
    private final boolean enableProfiling;
    private final String profilingEndpoint;
    private final String httpScheme;
    private final String httpAddress;
    private final int httpPort;

    public Config() {
        this.debug = getEnvAsBool("DEBUG", false);
        this.token = getEnv("TOKEN", "todo");
        this.serviceName = getEnv("SERVICE_NAME", "helloworld");
        this.otlpEndpoint = getEnv("OTLP_ENDPOINT", "http://localhost:4317");
        this.otlpExporterType = getEnv("OTLP_EXPORTER_TYPE", "grpc");
        this.profilingEndpoint = getEnv("PROFILING_ENDPOINT", "http://localhost:4040") ;
        this.httpScheme = "http";
        this.httpAddress = "0.0.0.0";
        this.httpPort = 8080;
        this.enableLogs = getEnvAsBool("ENABLE_LOGS", this.debug);
        this.enableTraces = getEnvAsBool("ENABLE_TRACES", this.debug);
        this.enableMetrics = getEnvAsBool("ENABLE_METRICS", this.debug);
        this.enableProfiling = getEnvAsBool("ENABLE_PROFILING", this.debug);
    }

    private static String getEnv(String key, String defaultValue) {
        return Optional.ofNullable(System.getenv(key)).orElse(defaultValue);
    }

    private static boolean getEnvAsBool(String key, boolean defaultValue) {
        String value = System.getenv(key);
        if (value == null) {
            return defaultValue;
        }
        return Boolean.parseBoolean(value);
    }
}
