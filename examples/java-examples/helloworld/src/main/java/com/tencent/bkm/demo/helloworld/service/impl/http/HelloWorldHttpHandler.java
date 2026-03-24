// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.helloworld.service.impl.http;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.tencent.bkm.demo.helloworld.service.impl.http.exception.FileNotFoundException;
import com.tencent.bkm.demo.helloworld.service.impl.http.exception.MySQLConnectTimeoutException;
import com.tencent.bkm.demo.helloworld.service.impl.http.exception.NetworkUnreachableException;
import com.tencent.bkm.demo.helloworld.service.impl.http.exception.UserNotFoundException;
import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.common.AttributeKey;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.metrics.DoubleHistogram;
import io.opentelemetry.api.metrics.LongCounter;
import io.opentelemetry.api.metrics.Meter;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.SpanKind;
import io.opentelemetry.api.trace.StatusCode;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.Scope;
import io.opentelemetry.context.propagation.TextMapGetter;
import io.opentelemetry.semconv.ClientAttributes;
import io.opentelemetry.semconv.HttpAttributes;
import io.opentelemetry.semconv.UrlAttributes;
import io.opentelemetry.semconv.UserAgentAttributes;
import io.opentelemetry.semconv.ServerAttributes;
import io.opentelemetry.semconv.NetworkAttributes;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.Charset;
import java.util.List;
import java.util.Random;

/**
 * HelloWorld HTTP 处理器.
 */
public class HelloWorldHttpHandler implements HttpHandler {

    private static final String[] COUNTRIES = {
            "United States", "Canada", "United Kingdom", "Germany",
            "France", "Japan", "Australia", "China", "India", "Brazil"
    };

    private static final Logger logger = LogManager.getLogger(HelloWorldHttpHandler.class);

    private final HttpConfig config;
    private final Tracer tracer;
    private final Meter meter;

    private final LongCounter requestsTotal;
    private final DoubleHistogram taskExecuteDurationSeconds;

    /**
     * 构造函数.
     *
     * @param config HTTP 配置
     */
    public HelloWorldHttpHandler(HttpConfig config) {
        this.config = config;
        this.tracer = GlobalOpenTelemetry.getTracer(this.config.getServiceName());
        this.meter = GlobalOpenTelemetry.getMeter(this.config.getServiceName());

        this.requestsTotal = this.meter.counterBuilder("requests_total")
                .setDescription("Total number of HTTP requests")
                .setUnit("requests")
                .build();

        this.taskExecuteDurationSeconds = this.meter
                .histogramBuilder("task_execute_duration_seconds")
                .setDescription("Task execute duration in seconds")
                .setExplicitBucketBoundariesAdvice(
                    List.of(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0)
                 )
                .setUnit("seconds")
                .build();

        // Metrics（指标） - Gauge 类型
        this.metricsGaugeDemo();
    }

    /**
     * 处理 HTTP 请求.
     *
     * @param exchange HTTP 交换对象
     * @throws IOException IO 异常
     */
    @Override
    public void handle(HttpExchange exchange) throws IOException {
        // Refer:
        // https://opentelemetry.io/docs/languages/java/instrumentation/#context-propagation-between-http-requests
        Context extractedContext = GlobalOpenTelemetry.getPropagators()
                .getTextMapPropagator()
                .extract(Context.current(), exchange, new TextMapGetter<>() {
                    @Override
                    public Iterable<String> keys(HttpExchange carrier) {
                        return carrier.getRequestHeaders().keySet();
                    }

                    @Override
                    public String get(HttpExchange carrier, String key) {
                        if (carrier.getRequestHeaders().containsKey(key)) {
                            return carrier.getRequestHeaders().get(key).get(0);
                        }
                        return null;
                    }
                });

        Long serverPort = Integer.toUnsignedLong(
                exchange.getLocalAddress().getPort());
        Long clientPort = Integer.toUnsignedLong(
                exchange.getRemoteAddress().getPort());
        String serverAddress = exchange.getLocalAddress()
                .getAddress().getHostAddress();
        String clientAddress = exchange.getRemoteAddress()
                .getAddress().getHostAddress();
        String userAgent = exchange.getRequestHeaders()
                .getOrDefault("user-agent", List.of("unknown")).get(0);
        Span span = this.tracer.spanBuilder("HTTP Server")
                .setSpanKind(SpanKind.SERVER)
                .setParent(extractedContext)
                .setAttribute(UserAgentAttributes.USER_AGENT_ORIGINAL, userAgent)
                .setAttribute(HttpAttributes.HTTP_REQUEST_METHOD, exchange.getRequestMethod())
                .setAttribute(UrlAttributes.URL_SCHEME, exchange.getRequestURI().getScheme())
                .setAttribute(UrlAttributes.URL_PATH, exchange.getRequestURI().getPath())
                .setAttribute(UrlAttributes.URL_QUERY, exchange.getRequestURI().getQuery())
                .setAttribute(ClientAttributes.CLIENT_PORT, clientPort)
                .setAttribute(ServerAttributes.SERVER_PORT, serverPort)
                .setAttribute(ClientAttributes.CLIENT_ADDRESS, clientAddress)
                .setAttribute(ServerAttributes.SERVER_ADDRESS, serverAddress)
                .setAttribute(NetworkAttributes.NETWORK_PROTOCOL_NAME, exchange.getProtocol())
                .setAttribute(NetworkAttributes.NETWORK_TRANSPORT, NetworkAttributes.NetworkTransportValues.TCP)
                .setAttribute(NetworkAttributes.NETWORK_LOCAL_ADDRESS, serverAddress)
                .setAttribute(NetworkAttributes.NETWORK_LOCAL_PORT, serverPort)
                .setAttribute(NetworkAttributes.NETWORK_PEER_ADDRESS, clientAddress)
                .setAttribute(NetworkAttributes.NETWORK_PEER_PORT, clientPort)
                .startSpan();

        try (Scope scope = span.makeCurrent()) {
            String greeting = handleHelloWorld(exchange);
            exchange.sendResponseHeaders(200, greeting.length());

            OutputStream os = exchange.getResponseBody();
            os.write(greeting.getBytes(Charset.defaultCharset()));
            os.close();

            span.setStatus(StatusCode.OK);
            span.setAttribute(HttpAttributes.HTTP_RESPONSE_STATUS_CODE, 200L);

        } catch (Exception e) {
            exchange.sendResponseHeaders(500, e.getMessage().length());

            OutputStream os = exchange.getResponseBody();
            os.write(e.getMessage().getBytes(Charset.defaultCharset()));
            os.close();

            span.recordException(e);
            span.setStatus(StatusCode.ERROR, e.getMessage());
            span.setAttribute(HttpAttributes.HTTP_RESPONSE_STATUS_CODE, 500L);

        } finally {
            span.end();
        }
    }

    /**
     * 处理 HelloWorld 业务逻辑.
     *
     * @param exchange HTTP 交换对象
     * @return 问候语
     * @throws Exception 异常
     */
    public String handleHelloWorld(HttpExchange exchange) throws Exception {
        Span span = this.tracer.spanBuilder("Handle/HelloWorld").startSpan();
        try (Scope ignored = span.makeCurrent()) {
            // Logs（日志）
            this.logsDemo(exchange);

            String country = choiceCountry();
            logger.info("get country -> {}", country);

            // Metrics（指标） - Counter 类型
            this.metricsCounterDemo(country);
            // Metrics（指标） - Histograms 类型
            this.metricsHistogramDemo();

            // Traces（调用链）- 自定义 Span
            this.tracesCustomSpanDemo();
            // Traces（调用链）- Span 事件
            this.tracesSpanEventDemo();
            // Traces（调用链）- 模拟错误
            tracesRandomErrorDemo();

            return generateGreeting(country);
        } catch (Exception e) {
            span.recordException(e);
            throw e;
        } finally {
            span.end();
        }
    }

    /**
     * Logs（日志）打印日志.
     * Refer:
     * https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/
     * instrumentation/log4j/log4j-appender-2.17/library/README.md
     *
     * @param exchange HTTP 交换对象
     */
    private void logsDemo(HttpExchange exchange) {
        logger.info("received request: {} {}",
                exchange.getRequestMethod(), exchange.getRequestURI());
    }

    /**
     * Metrics（指标）- 使用 Counter 类型指标.
     * Refer: https://opentelemetry.io/docs/languages/java/instrumentation/#using-counters
     *
     * @param country 国家名称
     */
    private void metricsCounterDemo(String country) {
        this.requestsTotal.add(1,
                Attributes.of(AttributeKey.stringKey("country"), country));
    }

    /**
     * Metrics（指标）- 使用 Histogram 类型指标.
     * Refer:
     * https://opentelemetry.io/docs/languages/java/instrumentation/#using-histograms
     */
    private void metricsHistogramDemo() {
        long begin = System.nanoTime();
        doSomething(100);
        long end = System.nanoTime();
        double durationInSeconds = (end - begin) / 1_000_000_000.0;
        // 记录耗时
        taskExecuteDurationSeconds.record(durationInSeconds);
    }

    /**
     * Metrics（指标）- 使用 Gauge 类型指标.
     * Refer:
     * https://opentelemetry.io/docs/languages/java/instrumentation/#using-observable-async-gauges
     */
    private void metricsGaugeDemo() {
        this.meter.gaugeBuilder("memory_usage")
                .setDescription("Memory usage")
                .buildWithCallback(
                        result -> {
                            Random random = new Random();
                            result.record(0.1 + random.nextDouble() * 0.2);
                        }
                );
    }

    /**
     * Traces（调用链）- 增加自定义 Span.
     * Refer: https://opentelemetry.io/docs/languages/java/instrumentation/#create-spans
     */
    private void tracesCustomSpanDemo() {
        Span span = tracer.spanBuilder("CustomSpanDemo/doSomething").startSpan();
        try (Scope ignored = span.makeCurrent()) {
            // 增加 Span 自定义属性
            // Refer:
            // https://opentelemetry.io/docs/languages/java/instrumentation/#span-attributes
            span.setAttribute(AttributeKey.longKey("helloworld.kind"), 1L);
            span.setAttribute(AttributeKey.stringKey("helloworld.step"),
                    "tracesCustomSpanDemo");
            doSomething(50);
        } finally {
            span.end();
        }
    }

    /**
     * Traces（调用链）- Span 事件.
     * Refer:
     * https://opentelemetry.io/docs/languages/java/instrumentation/#create-spans-with-events
     */
    private void tracesSpanEventDemo() {
        Span span = tracer.spanBuilder("tracesSpanEventDemo/doSomething").startSpan();
        try (Scope ignored = span.makeCurrent()) {
            Attributes evnetAttributes = Attributes.of(
                    AttributeKey.longKey("helloworld.kind"), 2L,
                    AttributeKey.stringKey("helloworld.step"), "tracesSpanEventDemo"
            );
            span.addEvent("Before doSomething", evnetAttributes);
            doSomething(50);
            span.addEvent("After doSomething", evnetAttributes);
        } finally {
            span.end();
        }
    }

    /**
     * Traces（调用链）- 异常事件、状态.
     * Refer:
     * https://opentelemetry.io/docs/languages/java/instrumentation/#set-span-status
     *
     * @throws Exception 异常
     */
    private void tracesRandomErrorDemo() throws Exception {
        try {
            throwRandomError(0.1F);
        } catch (Exception e) {
            logger.error("[tracesRandomErrorDemo] got error -> {}",
                    e.getMessage());
            // 获取当前 Span
            // Refer:
            // https://opentelemetry.io/docs/languages/java/instrumentation/#get-the-current-span
            Span span = Span.current();
            // 设置状态
            // Refer:
            // https://opentelemetry.io/docs/languages/java/instrumentation/#set-span-status
            span.setStatus(StatusCode.ERROR, e.getMessage());
            // 记录异常事件
            // Refer:
            // https://opentelemetry.io/docs/languages/java/instrumentation/#record-exceptions-in-spans
            span.recordException(e);
            throw e;
        }
    }

    private String generateGreeting(String country) {
        return "Hello World, " + country + "!";
    }

    private String choiceCountry() {
        Random random = new Random();
        int index = random.nextInt(COUNTRIES.length);
        return COUNTRIES[index];
    }

    private void doSomething(int maxMs) {
        Random random = new Random();
        int sleepTime = 10 + random.nextInt(maxMs);
        try {
            Thread.sleep(sleepTime);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    private void throwRandomError(float rate) throws Exception {
        Random random = new Random();
        if (random.nextFloat() < rate) {
            int index = random.nextInt(4);
            switch (index) {
                case 0:
                    throw new FileNotFoundException();
                case 1:
                    throw new MySQLConnectTimeoutException();
                case 2:
                    throw new NetworkUnreachableException();
                case 3:
                    throw new UserNotFoundException();
                default:
                    break;
            }
        }
    }
}
