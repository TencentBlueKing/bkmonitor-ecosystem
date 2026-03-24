// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.helloworld.service.impl.querier;

import com.tencent.bkm.demo.helloworld.service.Config;
import com.tencent.bkm.demo.helloworld.service.Service;
import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Scope;
import io.opentelemetry.instrumentation.httpclient.JavaHttpClientTelemetry;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

/**
 * Querier 服务实现类.
 */
public class QuerierService implements Service {

    private static final Logger logger = LogManager.getLogger(QuerierService.class);

    private QuerierConfig config;

    private Tracer tracer;
    private HttpClient httpClient;

    /**
     * 获取服务类型.
     *
     * @return 服务类型
     */
    @Override
    public String getType() {
        return "querier";
    }

    /**
     * 初始化服务.
     *
     * @param config 配置对象
     * @throws Exception 异常
     */
    @Override
    public void init(Config config) throws Exception {

        this.config = new QuerierConfig(
                config.getServiceName(), config.getHttpPort(),
                config.getHttpAddress(), config.getHttpScheme());

        this.tracer = GlobalOpenTelemetry.getTracer(this.config.getServiceName());

        this.httpClient = JavaHttpClientTelemetry.builder(GlobalOpenTelemetry.get())
                .build()
                .newHttpClient(HttpClient.newBuilder().build());
    }

    /**
     * 启动服务.
     *
     * @throws Exception 异常
     */
    @Override
    public void start() throws Exception {
        loopQueryHelloWorld();
    }

    /**
     * 停止服务.
     *
     * @throws Exception 异常
     */
    @Override
    public void stop() throws Exception {

    }

    private void loopQueryHelloWorld() {
        ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(1);
        scheduler.scheduleAtFixedRate(this::queryHelloWorld,
                0, 3, TimeUnit.SECONDS);
        Runtime.getRuntime().addShutdownHook(new Thread(scheduler::shutdown));
    }

    private void queryHelloWorld() {
        HttpRequest request = HttpRequest.newBuilder()
                .GET()
                .uri(URI.create(this.config.getEndpoint()))
                .build();

        Span span = this.tracer.spanBuilder("Caller/queryHelloWorld").startSpan();
        try (Scope ignored = span.makeCurrent()) {
            logger.info("[queryHelloWorld] send request");
            HttpResponse<String> response = this.httpClient.send(request,
                    HttpResponse.BodyHandlers.ofString());
            logger.info("[queryHelloWorld] received: {}", response.body());
        } catch (Exception e) {
            logger.error("[queryHelloWorld] got error -> {}", e.getMessage());
        } finally {
            span.end();
        }
    }
}
