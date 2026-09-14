# Profiling-JavaScript（Pyroscope SDK）接入

本指南将帮助您使用 Pyroscope Node.js SDK 接入蓝鲸应用性能监控，以入门项目-Profiling 为例，介绍性能分析数据接入及 SDK 使用场景。

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
* 若本地直接运行，需要 Node.js `20.20.2` 或更高版本。当前 `@pyroscope/nodejs@0.6.3` 声明了该最低版本，样例 Docker 镜像使用 Node 22。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/js-examples/profiling
docker build -t profiling-js:latest .
```

## 2. 快速体验

### 2.1 运行样例

运行参数说明：

| 参数                   | 推荐值                                | 说明                                        |
|----------------------|------------------------------------|-------------------------------------------|
| `TOKEN`              | `""`                               | APM 应用 `Token`。                            |
| `PROFILING_ENDPOINT` | `"http://127.0.0.1:4318/pyroscope"` | Profiling 数据上报地址，请根据页面指引提供的 HTTP 接入地址进行填写。<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。 |
| `SERVICE_NAME`       | `"helloworld"`                     | 服务唯一标识，一个应用可以有多个服务，通过该属性区分。                |
| `ENABLE_PROFILING`   | `false`                            | 是否启用性能分析上报。                                |

💡 为保证数据能上报到平台，`TOKEN`、`PROFILING_ENDPOINT` 请务必根据应用接入指引提供的实际值填写。

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="xxx" \
-e SERVICE_NAME="helloworld" \
-e PROFILING_ENDPOINT="http://127.0.0.1:4318/pyroscope" \
-e ENABLE_PROFILING="true" profiling-js:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

### 2.2 查看数据

等待片刻，便可在「服务详情-Profiling」看到应用数据。

![](image/img.png)

## 3. 快速接入

### 3.1 Pyroscope SDK

<a href="https://grafana.com/docs/pyroscope/latest/" target="_blank">Pyroscope</a> 是 Grafana 旗下用于聚合连续分析数据的开源软件项目。

请在创建 `PyroscopeConfig` 时，准确传入以下信息：

| 属性                | 说明                                            |
|-------------------|-----------------------------------------------|
| `AuthToken`       | 【必须】APM 应用 `Token`                            |
| `ApplicationName` | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分                |
| `ServerAddress`   | 【必须】Profiling 数据上报地址，请根据页面指引提供的 HTTP 接入地址进行填写 |

Node.js SDK 对应字段为 `authToken`、`appName`、`serverAddress`。Token 通过 `Authorization: Bearer` 发送，与 Go / Java Pyroscope SDK 的 `AuthToken` 一致。

在项目中引入模块依赖：

```shell
npm install @pyroscope/nodejs@0.6.3
```

示例项目提供集成 Pyroscope Node.js SDK 并将性能数据发送到 bk-collector 的方式，可以参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/master/examples/js-examples/profiling/src/profiling.js" target="_blank">src/profiling.js</a> 进行接入：

```javascript
const Pyroscope = require('@pyroscope/nodejs');

Pyroscope.init({
    // ❗❗【非常重要】请传入应用 Token
    authToken: config.token,
    // ❗❗【非常重要】应用服务唯一标识
    appName: config.serviceName,
    // ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
    serverAddress: config.profilingEndpoint,
    wall: {
        collectCpuTime: true,
    },
    heap: {
        samplingIntervalBytes: 524288,
        stackDepth: 64,
    },
});
Pyroscope.start();
```

参考官方文档以获得更多信息：<a href="https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/nodejs/" target="_blank">Configure the client to send profiles - Node.js</a>

### 3.2 采集类型

Node.js SDK 支持的采集类型为 CPU、Wall、Heap，与官方 <a href="https://grafana.com/docs/pyroscope/latest/configure-client/profile-types/" target="_blank">Profile types</a> 表一致。不支持 Goroutine、Mutex、Block 等类型。

| 采集类型 | 说明 |
|----------|------|
| CPU | 需设置 `wall.collectCpuTime: true`，否则页面看不到 CPU profile |
| Wall | 墙钟采样，`Pyroscope.start()` 默认启动 |
| Heap | 堆内存采样，`Pyroscope.start()` 默认启动 |

`Pyroscope.start()` 会同时启动 wall（含 CPU）和 heap。如只需其中一类，可改为 `startWallProfiling()` 或 `startHeapProfiling()`。

完整采样参数（上报周期、wall / heap 采样间隔等）见 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/master/examples/js-examples/profiling/src/profiling.js" target="_blank">src/profiling.js</a>。

### 3.3 动态标签

Wall / CPU profile 支持动态标签，用于区分不同代码路径。

示例项目在 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/master/examples/js-examples/profiling/src/server.js" target="_blank">src/server.js</a> 提供了创建样例：

```javascript
Pyroscope.wrapWithLabels({ handler: 'tasks' }, () => {
    intensiveTask();
});
```

参考官方文档以获得更多信息：<a href="https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/nodejs/" target="_blank">Dynamic labels for Wall/CPU profiles</a>

## 4. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>