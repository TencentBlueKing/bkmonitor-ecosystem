// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.metric;

import java.time.Duration;
import java.util.Optional;

import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.AttributeKey;
import io.opentelemetry.exporter.otlp.http.metrics.OtlpHttpMetricExporter;
import io.opentelemetry.instrumentation.resources.ContainerResource;
import io.opentelemetry.instrumentation.resources.HostResource;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.metrics.SdkMeterProvider;
import io.opentelemetry.sdk.metrics.export.MetricExporter;
import io.opentelemetry.sdk.metrics.export.PeriodicMetricReader;
import io.opentelemetry.sdk.resources.Resource;

/**
 * OTLP 配置类.
 */
public class OtlpConfiguration {
    // 上报超时时间
    private static final Duration EXPORTER_DEFAULT_TIMEOUT = Duration.ofSeconds(5);
    // 指标上报间隔
    private static final Duration EXPORTER_DEFAULT_INTERVAL = Duration.ofSeconds(30);

    private static volatile OtlpConfiguration instance;
    private final OpenTelemetry openTelemetry;

    private static String getEnv(String key, String defaultValue) {
        return Optional.ofNullable(System.getenv(key)).orElse(defaultValue);
    }

    private OtlpConfiguration() {
        Resource extraResource = Resource.builder()
                // 应用服务唯一标识
                .put(AttributeKey.stringKey("service.name"),
                        getEnv("SERVICE_NAME", "helloworld"))
                // 在这里可以定义、补充一些服务相关的维度
                .put(AttributeKey.stringKey("app"), "flink")
                .put(AttributeKey.stringKey("version"), "1.0.0")
                .put(AttributeKey.stringKey("namespace"), "Production")
                .put(AttributeKey.stringKey("env_name"), "formal")
                .put(AttributeKey.stringKey("instance"), "127.0.0.1")
                .build();

        // getDefault 提供了部分 SDK 默认属性
        Resource resource = Resource.getDefault()
                .merge(extraResource)
                .merge(ContainerResource.get())
                .merge(HostResource.get());

        MetricExporter exporter = OtlpHttpMetricExporter.builder()
                // 配置为应用 Token
                .addHeader("x-bk-token", getEnv("TOKEN", "fixme"))
                // 数据上报地址，请根据页面指引提供的接入地址进行填写
                .setEndpoint(getEnv("OTLP_ENDPOINT", "http://localhost:4318")
                        + "/v1/metrics")
                .setTimeout(EXPORTER_DEFAULT_TIMEOUT)
                .build();

        OpenTelemetrySdk openTelemetrySdk = OpenTelemetrySdk.builder()
                .setMeterProvider(
                        SdkMeterProvider.builder()
                                .setResource(resource)
                                .registerMetricReader(
                                        PeriodicMetricReader.builder(exporter)
                                                .setInterval(EXPORTER_DEFAULT_INTERVAL)
                                                .build())
                                .build()
                )
                .buildAndRegisterGlobal();

        Runtime.getRuntime().addShutdownHook(new Thread(openTelemetrySdk::close));
        this.openTelemetry = openTelemetrySdk;
    }

    /**
     * 获取单例.
     *
     * @return OtlpConfiguration 实例
     */
    public static OtlpConfiguration getInstance() {
        if (instance == null) {
            synchronized (OtlpConfiguration.class) {
                if (instance == null) {
                    instance = new OtlpConfiguration();
                }
            }
        }
        return instance;
    }

    /**
     * 获取 OpenTelemetry 实例.
     *
     * @return OpenTelemetry 实例
     */
    public OpenTelemetry getOpenTelemetry() {
        return openTelemetry;
    }
}
