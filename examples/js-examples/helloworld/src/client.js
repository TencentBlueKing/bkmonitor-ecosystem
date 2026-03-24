// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

const http = require('http');
const opentelemetry = require('@opentelemetry/api');
const { SeverityNumber } = require('@opentelemetry/api-logs');
const { io } = require('socket.io-client');


function socketIoDemo(config, logger, tracer, data) {
    tracer.startActiveSpan('client/socket.io',  {kind: opentelemetry.SpanKind.CLIENT}, span => {
        // socket.io client 样例：https://socket.io/docs/v4/client-installation/
        const socket = io(`http://${config.serverAddress}:${config.serverPort}`, {
            // 透传 TraceID Context 到服务端
            auth: (cb) => {
                const carrier = {};
                opentelemetry.propagation.inject(opentelemetry.context.active(), carrier);
                cb(carrier);
            }
        });

        socket.on('connect', () => {
            socket.emit('chat message', data);
            socket.on('chat message', (msg) => {
                logger.emit({
                    severityNumber: SeverityNumber.INFO,
                    severityText: 'info',
                    body: `[client][socket.io] received message: ${msg}`,
                });
                socket.disconnect();
            });
        });

        socket.on('disconnect', () => {
            logger.emit({
                severityNumber: SeverityNumber.INFO,
                severityText: 'info',
                body: `[client][socket.io] disconnected.`,
            });
            span.setStatus({ code: opentelemetry.SpanStatusCode.OK });
            span.end();
        });

        socket.on('connect_error', (err) => {
            logger.emit({
                severityNumber: SeverityNumber.ERROR,
                severityText: 'error',
                body: `[client][socket.io] connect_error: ${err.message}`,
            });
            span.recordException(err);
            span.setStatus({ code: opentelemetry.SpanStatusCode.ERROR, message: err.message });
            span.end();
        });
    });
}

function makeRequest(config, logger) {
    const tracer = opentelemetry.trace.getTracer(config.serviceName);
    tracer.startActiveSpan('client/makeRequest',  {kind: opentelemetry.SpanKind.CLIENT},async span => {
        const options = {
            hostname: config.serverAddress,
            port: config.serverPort,
            path: '/helloworld',
            method: 'GET'
        };
        const req = http.request(options, res => {
            let data = '';
            res.on('data', chunk => {
                data += chunk;
            });
            res.on('end', () => {
                logger.emit({
                    severityNumber: SeverityNumber.INFO,
                    severityText: 'info',
                    body: `[client] [${new Date().toISOString()}] Status: ${res.statusCode}, Response: ${data}`,
                });
            });
            socketIoDemo(config, logger, tracer, data);
        });

        req.on('error', error => {
            logger.emit({
                severityNumber: SeverityNumber.ERROR,
                severityText: 'error',
                body: `[client] [${new Date().toISOString()}] Request error: ${error.message}`,
            });
            span.recordException(error);
        });

        req.end();
        span.end();
    });
}

function startClient(config, logger) {
    const id = setInterval(() => makeRequest(config, logger), 3000);
    return () => {
        clearInterval(id);
        console.log('[client] service stopped.');
    };
}

module.exports = { startClient };
