// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.springbootstarter.service.http;

import com.tencent.bkm.demo.springbootstarter.service.http.exception.FileNotFoundException;
import com.tencent.bkm.demo.springbootstarter.service.http.exception.MySQLConnectTimeoutException;
import com.tencent.bkm.demo.springbootstarter.service.http.exception.NetworkUnreachableException;
import com.tencent.bkm.demo.springbootstarter.service.http.exception.UserNotFoundException;

import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.AttributeKey;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.metrics.DoubleHistogram;
import io.opentelemetry.api.metrics.LongCounter;
import io.opentelemetry.api.metrics.Meter;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.StatusCode;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.Scope;
import io.opentelemetry.semconv.HttpAttributes;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.context.ApplicationContext;
import org.springframework.stereotype.Service;
import org.springframework.scheduling.annotation.Async;

import java.util.List;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Random;

/**
 * 旅行服务类.
 */
@Service
public class TravelService {
    private static final String[] COUNTRIES = {
            "United States", "Canada", "United Kingdom", "Germany",
            "France", "Japan", "Australia", "China", "India", "Brazil"
    };
    private static final float ERROR_RATE = 0.1F;
    private static final Logger logger = LogManager.getLogger(TravelService.class);
    private final Tracer tracer;
    private final Meter meter;
    private final LongCounter visitRequestsTotal;
    private final DoubleHistogram visitExecuteDurationSeconds;
    private final ApplicationContext applicationContext;

    /**
     * 构造函数.
     *
     * @param openTelemetry OpenTelemetry 实例
     * @param applicationContext Spring 应用上下文
     */
    public TravelService(OpenTelemetry openTelemetry,
                         ApplicationContext applicationContext) {
        this.applicationContext = applicationContext;
        this.tracer = openTelemetry.getTracer(getClass().getName());
        this.meter = openTelemetry.getMeter(getClass().getName());
        this.visitRequestsTotal = this.meter
                .counterBuilder("visit_requests_total")
                .setDescription("Total calls to the visit function")
                .setUnit("requests")
                .build();
        this.visitExecuteDurationSeconds = this.meter
                .histogramBuilder("visit_execute_duration_seconds")
                .setDescription("Visit function execute duration in seconds")
                .setExplicitBucketBoundariesAdvice(
                        List.of(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0))
                .setUnit("seconds")
                .build();
    }

    /**
     * 处理旅行请求.
     *
     * @return 处理结果
     */
    public String handle() {
        Span span = this.tracer.spanBuilder("Travel/handle").startSpan();
        try (Scope scope = span.makeCurrent()) {
            List<String> countries = choiceCountries();
            logger.info("get countries -> {}", countries);

            this.parallelVisit(countries);
            this.serialVisit(countries);
        } catch (Exception e) {
            span.recordException(e);
            span.setStatus(StatusCode.ERROR, e.getMessage());
            span.setAttribute(HttpAttributes.HTTP_RESPONSE_STATUS_CODE, 500L);
        } finally {
             span.end();
        }
        return "Travel Success";
    };

    private List<String> choiceCountries() {
        ArrayList<String> copyCountries = new ArrayList<>(List.of(COUNTRIES));
        Collections.shuffle(copyCountries, new Random());
        return copyCountries.subList(0, 3);
    };

    private void parallelVisit(List<String> countries) throws Exception {
        Span span = this.tracer.spanBuilder("Travel/parallelVisit").startSpan();
        try (Scope scope = span.makeCurrent()) {
            // 获取当前的 Span
            Context currentContext = Context.current();
            for (String country : countries) {
                TravelService travelBean = this.applicationContext
                        .getBean(TravelService.class);
                travelBean.parallelTask(currentContext, country);
            }
        } catch (Exception e) {
            throw e;
        } finally {
            span.end();
        }
    };

    private void serialVisit(List<String> countries) throws Exception {
        Span span = this.tracer.spanBuilder("Travel/serialVisit").startSpan();
        try (Scope scope = span.makeCurrent()) {
            logger.info("Travel serialVisit start");
            for (String country : countries) {
                this.visit(country);
            }
            logger.info("Travel serialVisit end");
        } catch (Exception e) {
            throw e;
        } finally {
            span.end();
        }
    }

    /**
     * 并行任务处理.
     *
     * @param parentContext 父上下文
     * @param country 国家名称
     * @throws Exception 异常
     */
    @Async
    public void parallelTask(Context parentContext, String country) throws Exception {
        // 从父上下文中获取父 Span
        Span span = this.tracer
                .spanBuilder("Travel/parallelTask")
                .setParent(parentContext)
                .startSpan();
        try (Scope scope = span.makeCurrent()) {
            logger.info("Travel parallelTask start");
                this.visit(country);
            logger.info("Travel parallelTask end");
        } catch (Exception e) {
            throw e;
        } finally {
            span.end();
        }
    };

    private void visit(String country) throws Exception {
        long startTime = System.nanoTime();
        Span span = this.tracer.spanBuilder("Travel/visit").startSpan();
        try (Scope scope = span.makeCurrent()) {
            // Metrics（指标） - Counter 类型
            // 记录 visit 函数的调用次数，并按国家进行分类
            this.metricsCounterDemo(country);
            // Traces（调用链）- 模拟任务耗时
            // 随机休眠 200 ms~500 ms
            this.doSomething();
            // Traces（调用链）- 模拟错误
            // 10% 的概率抛出一个随机的错误
            this.tracesRandomErrorDemo();
            // Metrics（指标） - Histograms 类型
            // 记录 visit 函数的耗时
            this.metricsHistogramDemo(startTime);
        } catch (Exception e) {
            throw e;
        } finally {
            span.end();
        }
    };

    /**
     * Metrics（指标）- 使用 Counter 类型指标.
     * Refer:
     * https://opentelemetry.io/docs/languages/java/instrumentation/#using-counters
     *
     * @param country 国家名称
     */
    private void metricsCounterDemo(String country) {
        this.visitRequestsTotal.add(1,
                Attributes.of(AttributeKey.stringKey("country"), country));
    };

    private void doSomething() {
        Random random = new Random();
        int sleepTime = 200 + random.nextInt(300);
        try {
            Thread.sleep(sleepTime);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    };

    /**
     * Traces（调用链）- 异常事件、状态.
     * Refer:
     * https://opentelemetry.io/docs/languages/java/instrumentation/#set-span-status
     *
     * @throws Exception 异常
     */
    private void tracesRandomErrorDemo() throws Exception {
        try {
            throwRandomError(ERROR_RATE);
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
    };

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

    /**
     * Metrics（指标）- 使用 Histogram 类型指标.
     * Refer:
     * https://opentelemetry.io/docs/languages/java/instrumentation/#using-histograms
     *
     * @param startTime 开始时间
     */
    private void metricsHistogramDemo(Long startTime) {
        long endTime = System.nanoTime();
        double durationInSeconds = (endTime - startTime) / 1_000_000_000.0;
        // 记录耗时
        this.visitExecuteDurationSeconds.record(durationInSeconds);
    }

}