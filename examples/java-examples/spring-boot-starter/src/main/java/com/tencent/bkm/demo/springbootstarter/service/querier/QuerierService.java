// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.springbootstarter.service.querier;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.client.RestClient;
import org.springframework.stereotype.Service;
import org.springframework.scheduling.annotation.Scheduled;

import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.context.Scope;


@Service
public class QuerierService{
    private final Tracer tracer;
    private final QuerierConfig querierConfig;
    private final RestClient restClient;
    private final Logger logger = LoggerFactory.getLogger(getClass().getName());

    // RestClient.Builder 参考：https://opentelemetry.io/docs/zero-code/java/spring-boot-starter/out-of-the-box-instrumentation/#spring-web-autoconfiguration
    public QuerierService(OpenTelemetry openTelemetry, QuerierConfig querierConfig, RestClient.Builder restClientBuilder) {
        this.tracer = openTelemetry.getTracer(getClass().getName());
        this.querierConfig = querierConfig;
        this.restClient = restClientBuilder.build();
    };

    @Scheduled(fixedRate = 1000)
    private void queryHelloWorld() {
        Span span = this.tracer.spanBuilder("Caller/queryTravel").startSpan();

        try (Scope ignored = span.makeCurrent()) {
            logger.info("[queryTravel] send request");
            ResponseEntity<String> response = this.restClient.get()
                    .uri(this.querierConfig.getEndpoint())
                    .retrieve()
                    .toEntity(String.class);
            logger.info("[queryTravel] received: {}", response.getBody());
        } catch (Exception e) {
            logger.error("[queryTravel] got error -> {}", e.getMessage());
        } finally {
            span.end();
        }
    }
}