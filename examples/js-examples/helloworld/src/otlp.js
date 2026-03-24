// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

const { NodeSDK } = require('@opentelemetry/sdk-node');
const { resourceFromAttributes, defaultResource } = require('@opentelemetry/resources');
const { ATTR_SERVICE_NAME } = require('@opentelemetry/semantic-conventions');
const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');

const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-http');
const { OTLPMetricExporter } = require('@opentelemetry/exporter-metrics-otlp-http');
const { OTLPLogExporter } = require('@opentelemetry/exporter-logs-otlp-http');

const { LoggerProvider, BatchLogRecordProcessor } = require('@opentelemetry/sdk-logs');
const { PeriodicExportingMetricReader, AggregationType, InstrumentType} = require('@opentelemetry/sdk-metrics');


let logger;

const setupOtlp = (config) => {
    const resource = defaultResource().merge(
        resourceFromAttributes({
            // ❗❗【非常重要】应用服务唯一标识。
            [ATTR_SERVICE_NAME]: config.serviceName,
        })
    );
    const sdkConfig = {
        resource: resource,
        // 自动检测资源信息，例如进程名称、所在操作系统等。
        autoDetectResources: true,
        // 对 express、socket.io、mysql2、mongodb 等常用库进行自动插桩。
        // 可根据需要选择性引入所需的插件，详见文档：
        // https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/auto-instrumentations-node
        instrumentations: [getNodeAutoInstrumentations({
            '@opentelemetry/instrumentation-socket.io': {
                // 对保留事件（connect、disconnect 等）也进行跟踪。
                traceReserved: true,
            },
        })],
        // 指定直方图（Histogram）的聚合配置。
        views: [{
            aggregation: {
                type: AggregationType.EXPLICIT_BUCKET_HISTOGRAM,
                // 请按埋点逻辑的实际耗时估算分桶。
                options: { boundaries: [0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0] },
            },
            // 匹配所有 Histogram 类型的指标。
            instrumentName: '*',
            instrumentType: InstrumentType.HISTOGRAM,
        }],
    };

    const commonExporterConfig = {
        // ❗❗【非常重要】请传入应用 Token。
        headers: {'x-bk-token': config.token},
    };
    if (config.enableTraces) {
        sdkConfig.traceExporter = new OTLPTraceExporter({
            ...commonExporterConfig,
            // ❗❗【非常重要】数据上报地址，otlpEndpoint 请根据页面指引提供的接入地址进行填写。
            url: `${config.otlpEndpoint}/v1/traces`,
        });
    }
    if (config.enableMetrics) {
        sdkConfig.metricReader = new PeriodicExportingMetricReader({
            exporter: new OTLPMetricExporter({
                ...commonExporterConfig,
                // ❗❗【非常重要】数据上报地址，otlpEndpoint 请根据页面指引提供的接入地址进行填写。
                url: `${config.otlpEndpoint}/v1/metrics`,
            }),
            // 指标上报周期：建议设置为 30 秒。
            // 上报周期越短，产生的点数越多，聚合耗时越长，如果只是分钟级别聚合，30 秒已经能保证较高的准确性。
            exportIntervalMillis: 30000,
        });
    }

    if (config.enableLogs) {
        const loggerProvider = new LoggerProvider({
            resource: resource,
            processors: [
                new BatchLogRecordProcessor(
                    new OTLPLogExporter({
                        // ❗❗【非常重要】数据上报地址，otlpEndpoint 请根据页面指引提供的接入地址进行填写。
                        ...commonExporterConfig, url: `${config.otlpEndpoint}/v1/logs`
                    })
                )
            ],
        });

        sdkConfig.loggerProvider = loggerProvider;
        logger = loggerProvider.getLogger(config.serviceName);
    }

    const sdk = new NodeSDK(sdkConfig);
    sdk.start();

    return () => {
        return sdk.shutdown().then(() => {
            console.log('[otlp] service stopped');
        }).catch((error) => {
            console.log('[otlp] ignored error during provider shutdown', error);
        });
    };
};

const getLogger = () => {
    // Logger 可能未启用，返回一个空的 emit 函数
    if (!logger) {
        return { emit: () => {} };
    }
    return logger;
};

module.exports = { setupOtlp, getLogger };
