// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

const Pyroscope = require('@pyroscope/nodejs');

let started = false;

function normalizeEndpoint(endpoint) {
    return endpoint.replace(/\/+$/, '');
}

function now() {
    return new Date().toISOString();
}

// Node.js SDK 成功上报默认不打日志（DEBUG=pyroscope 也只打印失败）。
// 样例包装 fetch，打印 ingest HTTP 状态，便于确认上报。
function installIngestLogger() {
    const originalFetch = globalThis.fetch.bind(globalThis);
    globalThis.fetch = async (input, init) => {
        const url = typeof input === 'string' ? input : (input && input.url) || String(input);
        const isIngest = url.includes('/ingest');
        const pathname = url.split('?')[0];
        try {
            const response = await originalFetch(input, init);
            if (isIngest) {
                console.log(`${now()} [profiling] ingest ${pathname} -> ${response.status}`);
            }
            return response;
        } catch (err) {
            if (isIngest) {
                console.error(`${now()} [profiling] ingest failed: ${err.message}`);
            }
            throw err;
        }
    };
}

function startProfiling(config) {
    if (!config.enableProfiling) {
        return;
    }

    installIngestLogger();
    Pyroscope.init({
        // ❗❗【非常重要】请传入应用 Token
        authToken: config.token,
        // ❗❗【非常重要】应用服务唯一标识
        appName: config.serviceName,
        // ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
        serverAddress: normalizeEndpoint(config.profilingEndpoint),
        // 上报周期，默认 15 s
        flushIntervalMs: 15 * 1000,
        wall: {
            // 单次 wall profile 时长，官方默认 60000
            samplingDurationMs: 15 * 1000,
            // 采样间隔（微秒），官方默认 10000
            samplingIntervalMicros: 10000,
            // 采集 CPU 时间；官方文档要求开启后才有 CPU profile
            collectCpuTime: true,
        },
        heap: {
            // 堆采样平均间隔字节数，官方默认 524288
            samplingIntervalBytes: 524288,
            // 堆采样最大栈深度，官方默认 64
            stackDepth: 64,
        },
        // 静态标签，可用于页面过滤
        tags: {
            'service.name': config.serviceName,
        },
    });

    // Node.js SDK 支持的采集类型：CPU、Wall、Heap。
    // Pyroscope.start() 会同时启动 wall（含 CPU）和 heap。
    // 如只需其中一类，可改为：
    // Pyroscope.startWallProfiling();
    // Pyroscope.startHeapProfiling();
    Pyroscope.start();
    started = true;
    console.log('[profiling] started');
}

async function stopProfiling() {
    if (!started) {
        return;
    }
    await Pyroscope.stop();
    started = false;
}

function wrapWithLabels(labels, fn) {
    if (!started) {
        return fn();
    }
    return Pyroscope.wrapWithLabels(labels, fn);
}

module.exports = {
    startProfiling,
    stopProfiling,
    wrapWithLabels,
};
