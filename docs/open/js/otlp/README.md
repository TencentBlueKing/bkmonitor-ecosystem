# JavaScript（OpenTelemetry SDK）接入

本指南将帮助您使用 OpenTelemetry SDK 接入蓝鲸应用性能监控，以 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/examples/helloworld.md" target="_blank">入门项目-HelloWorld</a> 为例，介绍调用链、指标、日志数据接入及 SDK 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速入门 OpenTelemetry，接入并体验蓝鲸应用性能监控相关功能。

## 1. 前置准备

### 1.1 术语介绍

* Traces：<a href="https://opentelemetry.io/docs/concepts/signals/traces/" target="_blank">调用链</a>，表示请求在应用程序的执行路径。
* Metrics：<a href="https://opentelemetry.io/docs/concepts/signals/metrics/" target="_blank">指标</a>，表示对运行服务的测量。
* Logs: <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">日志</a>，表示对事件的记录。
* Profiling: <a href="https://grafana.com/docs/pyroscope/latest/introduction/profiling/" target="_blank">性能分析</a>，表示对应用程序运行时资源的持续测量。
* Telemetry Data：观测数据，指代 Traces、Metrics、Logs、Profiling 等。
* APM：蓝鲸观测平台应用性能监控，提供四类观测数据开箱即用的观测能力。
* <a href="https://github.com/TencentBlueKing/bkmonitor-datalink/tree/main/pkg/collector" target="_blank">bk-collector</a>：腾讯蓝鲸的 APM 服务端组件，负责接收 Prometheus、OpenTelemetry、Jaeger、Skywalking 等主流开源组件的观测数据，并对数据进行清洗转发到观测平台链路。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：
* Git
* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/js-examples/helloworld
docker build -t helloworld-js:latest .
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

### 2.2 开箱即用 SDK 接入示例

OpenTelemetry 提供标准化的框架和工具包，用于创建和管理 Traces、Metrics、Logs 数据。

示例项目提供集成 OpenTelemetry JavaScript SDK 并将观测数据发送到 bk-collector 的方式，可以参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/js-examples/helloworld/src/otlp.js" target="_blank">src/otlp.js</a> 进行接入。

### 2.3 关键配置

🌟 请仔细阅读本章节，以确保观测数据能准确上报到 APM。

#### 2.3.0 安装依赖

本样例通过 <a href="https://www.npmjs.com/package/@opentelemetry/sdk-node" target="_blank">OpenTelemetry SDK for Node.js</a> 进行接入，需安装以下依赖：

```shell
npm install @opentelemetry/api @opentelemetry/resources @opentelemetry/semantic-conventions
npm install @opentelemetry/sdk-node @opentelemetry/sdk-metrics @opentelemetry/sdk-logs @opentelemetry/api-logs
npm install @opentelemetry/exporter-trace-otlp-http @opentelemetry/exporter-metrics-otlp-http @opentelemetry/exporter-logs-otlp-http
# 自动埋点，支持 express、socket.io、mysql2、mongodb 等常用库。
# refer: https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/auto-instrumentations-node
npm install @opentelemetry/auto-instrumentations-node
```

#### 2.3.1 上报地址 & 应用 Token

请在创建 <a href="https://opentelemetry.io/docs/specs/otel/protocol/exporter/" target="_blank">Exporter</a> 时准确传入以下信息：

| 参数         | 说明                            |
|------------|-------------------------------|
| `endpoint` | 【必须】数据上报地址，请根据页面指引提供的接入地址进行填写。 |
| `x-bk-token`| 【必须】APM 应用 Token，作为 headers 传递。 |

示例项目在 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/js-examples/helloworld/src/otlp.js" target="_blank">src/otlp.js setupOtlp</a> 提供了包含调用链、指标、日志的 SDK 初始化样例：

```javascript
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
        sdkConfig.loggerProvider = loggerProvider
        logger = loggerProvider.getLogger(config.serviceName);
    }

    const sdk = new NodeSDK(sdkConfig);
    sdk.start();
}

const getLogger = () => {
    // Logger 可能未启用，返回一个空的 emit 函数
    if (!logger) {
        return { emit: () => {} };
    }
    return logger;
};

module.exports = { setupOtlp, getLogger };
```

`x-bk-token` 也可以通过「环境变量」的方式进行配置：

```shell
export OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo"
```

配置优先级：SDK > 环境变量，更多请参考 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#header-configuration" target="_blank">Header Configuration</a>。

#### 2.3.2 服务信息

请在 <a href="https://opentelemetry.io/docs/concepts/resources/" target="_blank">Resources</a> 添加以下属性，蓝鲸观测平台通过这些属性，将数据关联到具体的应用、资源实体：

| 属性                       | 说明                                          |
|--------------------------|---------------------------------------------|
| `service.name`           | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分              |
| `net.host.ip`            | 【可选】关联 CMDB 主机                              |
| `telemetry.sdk.language` | 【可选】标识应用对应的开发语言，SDK Default Resource 会提供该属性 |
| `telemetry.sdk.name`     | 【可选】OT SDK 名称，SDK Default Resource 会提供该属性   |
| `telemetry.sdk.version`  | 【可选】OT SDK 版本，SDK Default Resource 会提供该属性   |
| `k8s.bcs.cluster.id`     | 【可选】集群 ID，支持自动关联。                                        |
| `k8s.pod.name`           | 【可选】Pod 名称                                       |
| `k8s.namespace.name`     | 【可选】Pod 所在命名空间                                |

**a. 如何自动发现容器信息**

蓝鲸 APM 支持与 BCS 打通，你可以通过以下方式简单配置，将服务与容器信息进行关联，实现在 APM 查看服务所关联容器负载的监控、事件数据。

方案 1：🌟 通过集群内上报【推荐】

将上报域名切换为集群内域名，端口、上报路径与之前一致，即可自动获取关联。

方案 2：手动关联

手动补充上述的 `k8s.bcs.cluster.id`、`k8s.pod.name`、`k8s.namespace.name` 字段，也可以进行关联。

除了 `k8s.bcs.cluster.id` 外，可以在相应的 k8s 描述文件（Yaml）中，将 Pod 字段作为环境变量的值，然后在程序端读取，设置到 Resources：

```yaml
template:
  spec:
    containers:
      - name: xxx
        image: xxx
        env:
          - name: "K8S_POD_IP"
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
          - name: "K8S_POD_NAME"
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: "K8S_NAMESPACE"
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
```

示例项目在 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/js-examples/helloworld/src/otlp.js" target="_blank">src/otlp.js setupOtlp</a> 提供了创建样例：

```javascript
const setupOtlp = (config) => {
    const resource = defaultResource().merge(
        resourceFromAttributes({
            // ❗❗【非常重要】应用服务唯一标识。
            [ATTR_SERVICE_NAME]: config.serviceName,
        })
    );
    const sdkConfig = {
        // 自动检测资源信息，例如进程名称、所在操作系统等。
        autoDetectResources: true,
        // 设置自定义属性。
        resource: resource,
    };

    // ...
}
```
* `autoDetectResources` 设置为 `true` 时，将自动检测并上报资源信息，例如进程名称、所在操作系统等。
* `defaultResource` 提供了最基础的 SDK 和开发语言信息，建议保留，通过 `defaultResource().merge()` 增加自定义属性。

### 2.3.3 自动埋点

<a href="https://www.npmjs.com/package/@opentelemetry/sdk-node" target="_blank">OpenTelemetry SDK for Node.js</a> 提供了自动埋点（Auto Instrumentation）功能，对 express、socket.io、mysql2、mongodb 等常用库进行自动插桩。

请在初始化 SDK 时，传入「`instrumentations: [getNodeAutoInstrumentations()]`」。

```javascript
const opentelemetry = require("@opentelemetry/sdk-node");
const {
  getNodeAutoInstrumentations,
} = require("@opentelemetry/auto-instrumentations-node");

const sdk = new opentelemetry.NodeSDK({
    // ...
    instrumentations: [getNodeAutoInstrumentations()],
});
```

## 3. 使用场景

示例项目整理常见的使用场景，集中在：

```javascript
app.get("/helloworld", async (req, res) => {
    const country = countries[Math.floor(Math.random() * countries.length)];
    console.log(`[Server] get country -> ${country}`);

    // Logs（日志）- 打印日志
    logsDemo(req);

    // Metrics（指标） - Counter 类型
    metricsCounterDemo(country)
    // Metrics（指标） - Histograms 类型
    await metricsHistogramDemo()

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
        console.log(`[Server] Responding with error: ${err.message}`);
        res.status(500).send(err.message);
        return;
    }

    const greeting = `Hello World, ${country}!`;
    console.log(`[Server] Responding with greeting: ${greeting}`);
    res.status(200).send(greeting);
});
```

可以结合代码和下方说明进行使用：<a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/js-examples/helloworld/src/server.js" target="_blank">src/server.js</a>。

### 3.1 Traces

#### 3.1.1 创建 Span

Span 是 Traces 的构建块，代表一个工作或操作单元。

Span 通过 `otel.Tracer` 进行创建，<a href="https://opentelemetry.io/docs/specs/otel/trace/api/" target="_blank">`Tracer`</a> 是一个用于创建和管理 Span 的对象。它提供了 API 接口，开发人员可以用它来在应用程序代码中生成和记录 Span。

**后续样例提及的 `tracer` 创建方式如下：**

```javascript
const opentelemetry = require("@opentelemetry/api");

const serviceName = "helloworld";
const tracer = opentelemetry.trace.getTracer(config.serviceName);
````

通过 `startActiveSpan` 可以创建一个活跃的 Span，示例代码如下：

```javascript
const opentelemetry = require("@opentelemetry/api");

const serviceName = "helloworld";
const tracer = opentelemetry.trace.getTracer(config.serviceName);

// Traces（调用链）- 增加自定义 Span
// Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#create-spans
function tracesCustomSpanDemo() {
    return tracer.startActiveSpan("CustomSpanDemo/doSomething", (span) => {
        // 增加 Span 自定义属性
        // Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#attributes
        span.setAttributes({"helloworld.kind": 1, "helloworld.step": "tracesCustomSpanDemo"});
        return doSomething(100).then(() => {
            span.end();
        });
    });
}
```

* <a href="https://opentelemetry.io/docs/languages/js/instrumentation/#create-spans" target="_blank">Creating Spans</a>

#### 3.1.2 设置属性

Attributes（属性）是 Span 元数据，以 Key-Value 形式存在。

在 Span 设置属性，对问题定位、过滤、聚合非常有帮助。

```javascript
// 增加 Span 自定义属性
span.setAttributes({"helloworld.kind": 1, "helloworld.step": "tracesCustomSpanDemo"});
```

* <a href="https://opentelemetry.io/docs/languages/js/instrumentation/#attributes" target="_blank">Span Attributes</a>

#### 3.1.3 设置事件

Event（事件）是一段人类可读信息，用于记录 Span 生命周期内发生的事情。

```javascript
// Traces（调用链）- Span 事件
// Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#span-events
function tracesSpanEventDemo() {
    return tracer.startActiveSpan("SpanEventDemo/doSomething", (span) => {
        const opt = {"helloworld.kind": 2, "helloworld.step": "tracesSpanEventDemo"}
        span.addEvent("Before doSomething", opt);
        return doSomething(50).then(() => {
            span.addEvent("After doSomething", opt);
            span.end();
        });
    });
}
```

* <a href="https://opentelemetry.io/docs/languages/js/instrumentation/#span-events" target="_blank">Span Events</a>

#### 3.1.5 记录错误

当一个 Span 出现错误，可以对其进行错误记录。

```javascript
const opentelemetry = require("@opentelemetry/api");

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
        throw err;
    }
}
```

* <a href="https://opentelemetry.io/docs/languages/js/instrumentation/#recording-exceptions" target="_blank">Record errors</a>

#### 3.1.6 设置状态

当一个 Span 未能成功，可以通过设置状态进行显式标记。

```javascript
const opentelemetry = require("@opentelemetry/api");

// 设置 Span 状态为错误
// Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#span-status
currentSpan.setStatus({code: opentelemetry.SpanStatusCode.ERROR, message: err.message});
```
* <a href="https://opentelemetry.io/docs/languages/js/instrumentation/#span-status" target="_blank">Set span status</a>

#### 3.1.7 在当前 Span 上设置自定义属性

在部分场景下，Span 可能在框架入口、中间件等位置便被创建，如果你希望在当前的 Span 设置属性，可以使用 `getActiveSpan` 获取当前活跃的 Span，而不是新创建一个 Span，可以通过以下方式进行：

```javascript
const opentelemetry = require("@opentelemetry/api");

// Traces（调用链）- 在当前 Span 上设置自定义属性
// Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#get-the-current-span
function tracesSetCustomSpanAttributes() {
    const currentSpan = opentelemetry.trace.getActiveSpan();
    currentSpan.setAttributes({"ApiName": "ApiRequest", "actId": 12345});
}
```

* <a href="https://opentelemetry.io/docs/languages/js/instrumentation/#get-the-current-span" target="_blank">Get the current span</a>

### 3.2 Metrics

#### 3.2.1 创建 Meter

<a href="https://opentelemetry.io/docs/specs/otel/metrics/api/" target="_blank">`Meter`</a> 是一个负责创建 Instruments 的对象。它提供了 API 接口，允许开发人员在代码中定义和记录 Metrics。

后续样例提及的 `meter` 创建方式如下：

```javascript
const opentelemetry = require("@opentelemetry/api");

const serviceName = "helloworld"
const meter = opentelemetry.metrics.getMeter(config.serviceName);
```

#### 3.2.2 Counters

Counters（计数器）用于记录非负递增值。

例如，可以通过以下方式上报请求总数：

```javascript
//【建议】初始化指标再使用，而不是在业务逻辑里初始化
const requestsTotal = meter.createCounter("requests_total", {
    description: "Total number of HTTP requests"
});

// Metrics（指标）- 使用 Counter 类型指标
// Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#using-counters
function metricsCounterDemo(country) {
    requestsTotal.add(1, {country: country});
}
```
* <a href="https://opentelemetry.io/docs/languages/js/instrumentation/#using-counters" target="_blank">Using Counters</a>

#### 3.2.3 Histograms

Histograms（直方图）用于记录数值分布情况。

例如，可以通过以下方式上报某段逻辑的处理耗时：

```javascript
const taskExecuteDurationSeconds = meter.createHistogram("task_execute_duration_seconds", {
    description: "Task execute duration in seconds",
    unit: "s"
});

// Metrics（指标）- 使用 Histogram 类型指标
// Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#using-histograms
function metricsHistogramDemo() {
    const begin = Date.now();
    return doSomething(100).then(() => {
        const cost = (Date.now() - begin) / 1000;
        taskExecuteDurationSeconds.record(cost);
    });
}
```

* <a href="ttps://opentelemetry.io/docs/languages/js/instrumentation/#using-histograms" target="_blank">Using Histograms</a>

默认的分桶设置如果不满足需求，可以在初始化 SDK 时进行配置：

```javascript
const opentelemetry = require("@opentelemetry/sdk-node");
const {
  getNodeAutoInstrumentations,
} = require("@opentelemetry/auto-instrumentations-node");

const sdk = new opentelemetry.NodeSDK({
    // ...
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
});
```

* <a href="https://opentelemetry.io/docs/languages/js/instrumentation/#configure-metric-views" target="_blank">Configure Metric Views</a>

#### 3.2.4 Gauges

Gauges（仪表）用于记录瞬时值。

例如，可以通过以下方式，上报当前内存使用率：

```javascript
// Metrics（指标）- 使用 Gauge 类型指标
// Refer: https://opentelemetry.io/docs/languages/js/instrumentation/#using-observable-async-gauges
function metricsGaugeDemo() {
    const memoryUsage = meter.createObservableGauge("memory_usage", {
        description: "Memory usage"
    });
    meter.addBatchObservableCallback((observableResult) => {
        const usage = 0.1 + Math.random() * 0.2;
        observableResult.observe(memoryUsage, usage);
    }, [memoryUsage]);
}
```

* <a href="https://opentelemetry.io/docs/languages/js/instrumentation/#using-observable-async-gauges" target="_blank">Using Gauges</a>

### 3.3 Logs

#### 3.3.1 记录日志

```javascript
const { getLogger } = require("./otlp");
const { SeverityNumber } = require('@opentelemetry/api-logs');

const logger = getLogger();

// Logs（日志）- 打印日志
// Refer: https://github.com/open-telemetry/opentelemetry-js/tree/main/experimental/packages/exporter-logs-otlp-http
function logsDemo(req) {
    // 上报日志
    logger.emit({
        severityNumber: SeverityNumber.INFO,
        severityText: 'info',
        body: `received request: ${req.method} ${req.url}`,
    })

    // 添加自定义属性
    logger.emit({
        severityNumber: SeverityNumber.INFO,
        severityText: 'info',
        body: `report log with attrs, received request: ${req.method} ${req.url}`,
        attributes: {method: req.method, k1: 'v1', k2: 123}
    })
}
```

### 3.4 集成到 socket.io

<a href="https://socket.io/docs/v4/" target="_blank">socket.io</a> 是一个流行的库，用于在客户端和服务器之间实现实时、双向通信，以下将介绍如何使用 OpenTelemetry 对 socket.io 进行自动埋点。

#### 3.4.1 调用链自动埋点

OpenTelemetry 提供了对 socket.io 的自动埋点支持，可以通过 <a href="https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/instrumentation-socket.io" target="_blank">`@opentelemetry/auto-instrumentations-node`</a> 进行集成。

以下配置将对<a href="https://socket.io/docs/v4/emit-cheatsheet/" target="_blank">保留事件</a>（`connect`、`disconnect` 等）也进行跟踪：

```javascript
const opentelemetry = require("@opentelemetry/sdk-node");
const {
  getNodeAutoInstrumentations,
} = require("@opentelemetry/auto-instrumentations-node");

const sdk = new opentelemetry.NodeSDK({
    // ...
    instrumentations: [getNodeAutoInstrumentations({
        '@opentelemetry/instrumentation-socket.io': {
            // 对保留事件（connect、disconnect 等）也进行跟踪。
            traceReserved: true,
        },
    })],
});
```

* 完整代码请参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/js-examples/helloworld/src/otlp.js" target="_blank">src/otlp.js</a>。

#### 3.4.2 服务端 & 客户端调用链串联

通过「3.4.1 调用链自动埋点」后已经能采集到 socket.io 的调用链数据，如果需要将客户端和服务端的调用链进行串联，需要在客户端发起连接时，透传调用链上下文信息到服务端。

样例使用 <a href="https://socket.io/docs/v4/client-installation/" target="_blank">`socket.io-client`</a> 与 `socket.io` 进行通信：

```text
+---------+                  +---------+
| Client  |                  | Server  |
+---------+                  +---------+
     |  "Hello World, {country}!" |
     |--------------------------->|
     |           "Bye!"           |
     |<---------------------------|
     |                            |
```

在客户端发起建连时，可以通过 `auth` 或者 `headers` 透传调用链上下文信息到服务端：

```javascript
const opentelemetry = require('@opentelemetry/api');
const { io } = require('socket.io-client');

const tracer = opentelemetry.trace.getTracer('helloworld');

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
    // ...
});
```

服务端接收到 `connect` 后，可以通过 `socket.handshake.auth` 获取到调用链上下文信息，并创建一个新的 Span：

```javascript
const opentelemetry = require('@opentelemetry/api');
const { Server } = require('socket.io');
const { SpanKind } = require("@opentelemetry/api");

const tracer = opentelemetry.trace.getTracer('helloworld');

// socket.io: https://socket.io/docs/v4/server-installation/
io.on('connect', socket => {
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
            socket.emit('chat message', "Bye!");
        });
    });

    socket.on('disconnect', () => {
        span.setStatus({ code: opentelemetry.SpanStatusCode.OK });
        span.end();
    });
});
```

需要注意的是，socket.io 是异步通信，Span 需要在 `disconnect` 事件，或者根据业务场景自行结束，尽可能保证调用链能覆盖整个请求周期。

```javascript
socket.on('disconnect', () => {
    span.setStatus({ code: opentelemetry.SpanStatusCode.OK });
    span.end();
});
```

* 完整代码请参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/js-examples/helloworld/src/client.js" target="_blank">src/client.js</a> 和 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/js-examples/helloworld/src/server.js" target="_blank">src/server.js</a>。

#### 3.4.3 记录异常连接

在 socket.io 的 `connect_error` 事件中，可以获取到连接异常信息，并通过 `recordException` 进行记录：

```javascript
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
```

#### 3.4.4 记录指标

socket.io 是长连接，可以选取一些标志性事件，记录两个事件间的耗时，作为指标上报。

例如，可以在 `connect` 事件开始计时，在 `disconnect` 事件上报耗时：

```javascript
const opentelemetry = require('@opentelemetry/api');

const meter = opentelemetry.metrics.getMeter('default');
const socketIOHandledSeconds = meter.createHistogram('socket_io_handled_seconds', {
    description: 'Socket.IO message handled duration in seconds',
    unit: 's'
});

io.on('connect', socket => {
    const begin = Date.now();
    // ....
    socket.on('disconnect', () => {
        socketIOHandledSeconds.record( (Date.now() - begin) / 1000);
    });
});
```

## 4. 快速体验

### 4.1 运行样例

#### 4.1.1 运行

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="xxx" \
-e SERVICE_NAME="helloworld" \
-e OTLP_ENDPOINT="http://127.0.0.1:4318" \
-e ENABLE_TRACES="true" \
-e ENABLE_METRICS="true" \
-e ENABLE_LOGS="true" helloworld-js:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。


#### 4.1.2 运行参数说明

| 参数               | 值（根据所填写接入信息生成）                                          | 说明                                                                                                       |
|------------------|:--------------------------------------------------------|----------------------------------------------------------------------------------------------------------|
| `TOKEN`          | `"xxx"`                             | 【必须】APM 应用 `Token`。                                                                                      |
| `SERVICE_NAME`   | `"helloworld"`                                    | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分。                                                                          |
| `OTLP_ENDPOINT`  | `"127.0.0.1:4318"` | 【必须】OT 数据上报地址，支持以下协议：<br /> `HTTP`：`127.0.0.1:4318`（demo 使用该协议演示上报） |
| `ENABLE_TRACES`  | `true`                  | 是否启用调用链上报。                                                                                               |
| `ENABLE_METRICS` | `true`                 | 是否启用指标上报。                                                                                                |
| `ENABLE_LOGS`    | `true`                    | 是否启用日志上报。                                                                                                |

* *<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/" target="_blank">OTLP Exporter Configuration</a>*

### 4.2 查看数据

#### 4.2.1 Traces 检索

Tracing 检索功能主要用于对分布式系统中的请求链路进行跟踪和分析，请参考<a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/explore_traces.md" target="_blank">「应用性能监控 APM/调用链追踪」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/traces.png)

#### 4.2.2 指标检索

自定义指标功能旨在帮助用户针对特定应用及其服务进行深度性能指标监控，请参考<a href="#" target="_blank">「应用性能监控 APM/自定义指标」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/custom-metrics.png)

#### 4.2.3 日志检索

日志功能主要用于查看和分析对应服务（应用程序）运行过程中产生的各类日志信息，请参考<a href="#" target="_blank">「应用性能监控 APM/日志分析」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/logs.png)

## 5. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>