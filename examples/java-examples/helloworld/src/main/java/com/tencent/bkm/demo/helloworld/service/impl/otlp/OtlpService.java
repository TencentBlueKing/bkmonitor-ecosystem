// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.helloworld.service.impl.otlp;

import com.tencent.bkm.demo.helloworld.service.Config;
import com.tencent.bkm.demo.helloworld.service.Service;

import io.opentelemetry.api.common.AttributeKey;
import io.opentelemetry.api.trace.propagation.W3CTraceContextPropagator;
import io.opentelemetry.context.propagation.ContextPropagators;
import io.opentelemetry.instrumentation.resources.ContainerResource;
import io.opentelemetry.instrumentation.resources.HostResource;
import io.opentelemetry.instrumentation.resources.OsResource;
import io.opentelemetry.instrumentation.resources.ProcessResource;
import io.opentelemetry.sdk.logs.SdkLoggerProvider;
import io.opentelemetry.exporter.otlp.http.logs.OtlpHttpLogRecordExporter;
import io.opentelemetry.exporter.otlp.http.metrics.OtlpHttpMetricExporter;
import io.opentelemetry.exporter.otlp.http.trace.OtlpHttpSpanExporter;
import io.opentelemetry.exporter.otlp.logs.OtlpGrpcLogRecordExporter;
import io.opentelemetry.exporter.otlp.metrics.OtlpGrpcMetricExporter;
import io.opentelemetry.exporter.otlp.trace.OtlpGrpcSpanExporter;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.OpenTelemetrySdkBuilder;
import io.opentelemetry.sdk.metrics.SdkMeterProvider;
import io.opentelemetry.sdk.metrics.export.MetricExporter;
import io.opentelemetry.sdk.metrics.export.PeriodicMetricReader;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor;
import io.opentelemetry.sdk.trace.export.SpanExporter;
import io.opentelemetry.sdk.logs.export.BatchLogRecordProcessor;
import io.opentelemetry.sdk.logs.export.LogRecordExporter;

import java.time.Duration;
import java.util.HashMap;
import java.util.Map;

/**
 * OTLP 服务实现类.
 */
public class OtlpService implements Service {

    // https://javadoc.io/doc/io.opentelemetry

    private static final Duration EXPORTER_DEFAULT_TIMEOUT = Duration.ofSeconds(5);
    private static final Duration EXPORTER_DEFAULT_SCHEDULE_DELAY = Duration.ofSeconds(10);

    private OtlpConfig config;
    private OpenTelemetrySdk openTelemetrySdk;

    /**
     * 获取服务类型.
     *
     * @return 服务类型
     */
    @Override
    public String getType() {
        return "otlp";
    }

    /**
     * 初始化服务.
     *
     * @param config 配置对象
     * @throws Exception 异常
     */
    @Override
    public void init(Config config) throws Exception {
        this.config = new OtlpConfig(
                config.getToken(),
                config.getServiceName(),
                config.getOtlpEndpoint(),
                config.isEnableLogs(),
                config.isEnableTraces(),
                config.isEnableMetrics(),
                ExporterType.valueOf(
                        config.getOtlpExporterType().toUpperCase())
        );

        Resource resource = this.getResource();
        OpenTelemetrySdkBuilder openTelemetrySdkBuilder = OpenTelemetrySdk.builder();

        this.setUpLogs(openTelemetrySdkBuilder, resource);
        this.setUpTraces(openTelemetrySdkBuilder, resource);
        this.setUpMetrics(openTelemetrySdkBuilder, resource);

        // set global 需要放在任何模块的 get 前面
        this.openTelemetrySdk = openTelemetrySdkBuilder.buildAndRegisterGlobal();

        this.setUpLogsAppender();
    }

    /**
     * 启动服务.
     *
     * @throws Exception 异常
     */
    @Override
    public void start() throws Exception {

    }

    /**
     * 停止服务.
     *
     * @throws Exception 异常
     */
    @Override
    public void stop() throws Exception {
        if (this.openTelemetrySdk != null) {
            this.openTelemetrySdk.close();
        }
    }

    private void setUpTraces(OpenTelemetrySdkBuilder openTelemetrySdkBuilder, Resource resource) {
        if (this.config.isEnableTraces()) {
            openTelemetrySdkBuilder
                    .setTracerProvider(this.getTracerProvider(resource))
                    .setPropagators(ContextPropagators.create(
                            W3CTraceContextPropagator.getInstance()));
        }
    }

    private void setUpMetrics(OpenTelemetrySdkBuilder openTelemetrySdkBuilder,
                              Resource resource) {
        if (this.config.isEnableMetrics()) {
            openTelemetrySdkBuilder.setMeterProvider(this.getMeterProvider(resource));
        }
    }

    private void setUpLogs(OpenTelemetrySdkBuilder openTelemetrySdkBuilder,
                           Resource resource) {
        if (this.config.isEnableLogs()) {
            openTelemetrySdkBuilder.setLoggerProvider(this.getLoggerProvider(resource));
        }
    }

    private void setUpLogsAppender() {
        if (this.config.isEnableLogs()) {
            io.opentelemetry.instrumentation.log4j.appender.v2_17
                    .OpenTelemetryAppender.install(this.openTelemetrySdk);
        }
    }

    private SdkTracerProvider getTracerProvider(Resource resource) {
        SpanExporter exporter = this.getSpanExporter();
        return SdkTracerProvider.builder()
                .setResource(resource)
                .addSpanProcessor(
                        BatchSpanProcessor.builder(exporter)
                                .setScheduleDelay(EXPORTER_DEFAULT_SCHEDULE_DELAY)
                                .build())
                .build();
    }

    private SpanExporter getSpanExporter() {
        switch (config.getExporterType()) {
            case HTTP:
                return OtlpHttpSpanExporter.builder()
                        // 数据上报地址，请根据页面指引提供的接入地址进行填写
                        .setEndpoint(config.getEndpoint() + "/v1/traces")
                        .setTimeout(EXPORTER_DEFAULT_TIMEOUT)
                        .addHeader("x-bk-token", this.config.getToken())
                        .build();
            case GRPC:
                return OtlpGrpcSpanExporter.builder()
                        // 数据上报地址，请根据页面指引提供的接入地址进行填写
                        .setEndpoint(config.getEndpoint())
                        .setTimeout(EXPORTER_DEFAULT_TIMEOUT)
                        .addHeader("x-bk-token", this.config.getToken())
                        .build();
            default:
                throw new IllegalArgumentException(
                        "Unsupported exporter type: "
                                + config.getExporterType().toString().toLowerCase());
        }
    }

    private SdkMeterProvider getMeterProvider(Resource resource) {
        MetricExporter exporter = this.getMetricExporter();
        return SdkMeterProvider.builder()
                .setResource(resource)
                .registerMetricReader(
                        PeriodicMetricReader.builder(exporter)
                                .setInterval(EXPORTER_DEFAULT_SCHEDULE_DELAY)
                                .build())
                .build();
    }

    private MetricExporter getMetricExporter() {
        switch (config.getExporterType()) {
            case HTTP:
                return OtlpHttpMetricExporter.builder()
                        .setEndpoint(config.getEndpoint() + "/v1/metrics")
                        .setTimeout(EXPORTER_DEFAULT_TIMEOUT)
                        .addHeader("x-bk-token", this.config.getToken())
                        .build();
            case GRPC:
                return OtlpGrpcMetricExporter.builder()
                        .setEndpoint(config.getEndpoint())
                        .setTimeout(EXPORTER_DEFAULT_TIMEOUT)
                        .addHeader("x-bk-token", this.config.getToken())
                        .build();
            default:
                throw new IllegalArgumentException(
                        "Unsupported exporter type: "
                                + config.getExporterType().toString().toLowerCase());
        }
    }

    private SdkLoggerProvider getLoggerProvider(Resource resource) {
        LogRecordExporter exporter = this.getLogRecordExporter();
        return SdkLoggerProvider.builder()
                .setResource(resource)
                .addLogRecordProcessor(
                        BatchLogRecordProcessor.builder(exporter)
                                .setScheduleDelay(EXPORTER_DEFAULT_SCHEDULE_DELAY)
                                .build()
                )
                .build();
    }

    private LogRecordExporter getLogRecordExporter() {
        switch (config.getExporterType()) {
            case HTTP:
                return OtlpHttpLogRecordExporter.builder()
                        .setEndpoint(config.getEndpoint() + "/v1/logs")
                        .setTimeout(EXPORTER_DEFAULT_TIMEOUT)
                        .addHeader("x-bk-token", this.config.getToken())
                        .build();
            case GRPC:
                return OtlpGrpcLogRecordExporter.builder()
                        .setEndpoint(config.getEndpoint())
                        .setTimeout(EXPORTER_DEFAULT_TIMEOUT)
                        .addHeader("x-bk-token", this.config.getToken())
                        .build();
            default:
                throw new IllegalArgumentException(
                        "Unsupported exporter type: "
                                + config.getExporterType().toString().toLowerCase());
        }
    }

    private Resource getResource() {
        Resource extraResource = Resource.builder()
                // 应用服务唯一标识
                .put(AttributeKey.stringKey("service.name"),
                        this.config.getServiceName())
                .build();
        // getDefault 提供了部分 SDK 默认属性
        return Resource.getDefault()
                .merge(extraResource)
                .merge(ProcessResource.get())
                .merge(ContainerResource.get())
                .merge(OsResource.get())
                .merge(HostResource.get());
    }
}
