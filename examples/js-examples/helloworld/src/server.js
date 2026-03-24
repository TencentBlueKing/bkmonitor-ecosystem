// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

const express = require('express');
const opentelemetry = require('@opentelemetry/api');
const { Server } = require('socket.io');
const { SpanKind } = require("@opentelemetry/api");
const { SeverityNumber } = require('@opentelemetry/api-logs');

const countries = [
    'United States',
    'Canada',
    'United Kingdom',
    'Germany',
    'France',
    'Japan',
    'Australia',
    'China',
    'India',
    'Brazil'
];

const errors = [
    'mysql connect timeout',
    'user not found',
    'network unreachable',
    'file not found'
];

function startServer(config, logger) {
    function sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    async function doSomething(maxMs) {
        const randomDelay = 10 + Math.floor(Math.random() * maxMs);
        await sleep(randomDelay);
    }

    function throwRandomError(rate) {
        if (Math.random() < rate) {
            throw new Error(errors[Math.floor(Math.random() * errors.length)]);
        }
    }

    const meter = opentelemetry.metrics.getMeter(config.serviceName);
    const tracer = opentelemetry.trace.getTracer(config.serviceName);

    const requestsTotal = meter.createCounter('requests_total', {
       description: 'Total number of HTTP requests'
    });
    const taskExecuteDurationSeconds = meter.createHistogram('task_execute_duration_seconds', {
         description: 'Task execute duration in seconds',
         unit: 's'
    });
    const socketIOHandledSeconds = meter.createHistogram('socket_io_handled_seconds', {
        description: 'Socket.IO message handled duration in seconds',
        unit: 's'
    });

    metricsGaugeDemo();

    // Logs（日志）- 打印日志
    // Refer: https://github.com/open-telemetry/opentelemetry-js/tree/main/experimental/packages/exporter-logs-otlp-http
    function logsDemo(req) {
        // 上报日志
        logger.emit({
            severityNumber: SeverityNumber.INFO,
            severityText: 'info',
            body: `received request: ${req.method} ${req.url}`,
        });

        // 添加自定义属性
        logger.emit({
            severityNumber: SeverityNumber.INFO,
            severityText: 'info',
            body: `report log with attrs, received request: ${req.method} ${req.url}`,
            attributes: {method: req.method, k1: 'v1', k2: 123}
        });
    }

    // Metrics（指标）- 使用 Counter 类型指标
    // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#using-counters
    function metricsCounterDemo(country) {
        requestsTotal.add(1, {country: country});
    }

    // Metrics（指标）- 使用 Histogram 类型指标
    // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#using-histograms
    function metricsHistogramDemo() {
        const begin = Date.now();
        return doSomething(100).then(() => {
            const cost = (Date.now() - begin) / 1000;
            taskExecuteDurationSeconds.record(cost);
        });
    }

    // Metrics（指标）- 使用 Gauge 类型指标
    // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#using-observable-async-gauges
    function metricsGaugeDemo() {
        const memoryUsage = meter.createObservableGauge('memory_usage', {
            description: 'Memory usage'
        });
        meter.addBatchObservableCallback((observableResult) => {
            const usage = 0.1 + Math.random() * 0.2;
            observableResult.observe(memoryUsage, usage);
        }, [memoryUsage]);
    }

    // Traces（调用链）- 增加自定义 Span
    // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#create-spans
    function tracesCustomSpanDemo() {
        return tracer.startActiveSpan('CustomSpanDemo/doSomething', (span) => {
            // 增加 Span 自定义属性
            // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#attributes
            span.setAttributes({'helloworld.kind': 1, 'helloworld.step': 'tracesCustomSpanDemo'});
            return doSomething(100).then(() => {
                span.end();
            });
        });
    }

    // Traces（调用链）- 在当前 Span 上设置自定义属性
    // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#get-the-current-span
    function tracesSetCustomSpanAttributes() {
        const currentSpan = opentelemetry.trace.getActiveSpan();
        currentSpan.setAttributes({ApiName: 'ApiRequest', actId: 12345});
    }

    // Traces（调用链）- Span 事件
    // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#span-events
    function tracesSpanEventDemo() {
        return tracer.startActiveSpan('SpanEventDemo/doSomething', (span) => {
            const opt = {'helloworld.kind': 2, 'helloworld.step': 'tracesSpanEventDemo'};
            span.addEvent('Before doSomething', opt);
            return doSomething(50).then(() => {
                span.addEvent('After doSomething', opt);
                span.end();
            });
        });
    }

    // Traces（调用链）- 异常事件、状态
    // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#span-status
    function tracesRandomErrorDemo() {
        try {
            throwRandomError(0.1);
        } catch (err) {
            // 获取当前 Span
            // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#get-the-current-span
            const currentSpan = opentelemetry.trace.getActiveSpan();
            // 增加异常事件
            // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#recording-exceptions
            currentSpan.recordException(err);
            // 设置 Span 状态为错误
            // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#span-status
            currentSpan.setStatus({code: opentelemetry.SpanStatusCode.ERROR, message: err.message});
            throw err;
        }
    }

    const app = express();
    app.get('/helloworld', async (req, res) => {
        const country = countries[Math.floor(Math.random() * countries.length)];
        logger.emit({
            severityNumber: SeverityNumber.INFO,
            severityText: 'info',
            body: `[server] get country -> ${country}`,
        });

        // Logs（日志）- 打印日志
        logsDemo(req);

        // Metrics（指标） - Counter 类型
        metricsCounterDemo(country);
        // Metrics（指标） - Histograms 类型
        await metricsHistogramDemo();

        try {
            // Traces（调用链）- 自定义 Span
            await tracesCustomSpanDemo();
            // Traces（调用链）- 在当前 Span 上设置自定义属性
            tracesSetCustomSpanAttributes();
            // Traces（调用链）- Span 事件
            await tracesSpanEventDemo();
            // Traces（调用链）- 模拟错误
            tracesRandomErrorDemo();
        } catch (err) {
            logger.emit({
                severityNumber: SeverityNumber.ERROR,
                severityText: 'error',
                body: `[server] error happened: ${err.message}`,
            });
            res.status(500).send(err.message);
            return;
        }

        res.status(200).send(`Hello World, ${country}!`);
    });

    const server = app.listen(config.serverPort, config.serverAddress, () => {
        console.log(`[server] Server running at http://${config.serverAddress}:${config.serverPort}/`);
    });
    const io = new Server(server);


    // socket.io: https://socket.io/docs/v4/server-installation/
    io.on('connect', socket => {
        const begin = Date.now();
        logger.emit({
            severityNumber: SeverityNumber.INFO,
            severityText: 'info',
            body: `[server][socket.io] client connected, id: ${socket.id}`,
        });

        // 从 socket.handshake.auth 中提取 Trace，并激活上下文，用于将 client、server 两端的调用链关联起来。
        // socket.io 已通过 instrumentation-socket.io 自动埋点，这里只需传递上下文即可。
        // 了解更多: https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/instrumentation-socket.io
        const parentCtx = opentelemetry.propagation.extract(opentelemetry.context.active(), socket.handshake.auth);
        const span = tracer.startSpan('server/socket.io', {kind: SpanKind.SERVER}, parentCtx);

        socket._otel_context = {};
        opentelemetry.propagation.inject(
            opentelemetry.trace.setSpan(opentelemetry.context.active(), span),
            socket._otel_context
        );

        socket.on('chat message', (msg) => {
            const ctx = opentelemetry.propagation.extract(opentelemetry.context.active(), socket._otel_context);
            opentelemetry.context.with(ctx, () => {
                logger.emit({
                    severityNumber: SeverityNumber.INFO,
                    severityText: 'info',
                    body: `[server][socket.io] received message: ${msg}`,
                });
                socket.emit('chat message', "Bye!");
            });
        });

        socket.on('disconnect', () => {
            socketIOHandledSeconds.record((Date.now() - begin) / 1000);
            logger.emit({
                severityNumber: SeverityNumber.INFO,
                severityText: 'info',
                body: `[server][socket.io] client disconnected, id: ${socket.id}`,
            });

            span.setStatus({ code: opentelemetry.SpanStatusCode.OK });
            span.end();
        });
    });


    return () => {
        server.close(() => {
            console.log('[http] service stopped');
        });
    };
}

module.exports = {startServer};
