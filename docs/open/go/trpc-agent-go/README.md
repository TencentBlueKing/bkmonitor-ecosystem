# Go（tRPC Agent Go）接入

本指南介绍如何复用 tRPC Agent Go 内置的 OpenTelemetry 埋点，将 **Agent、模型调用和工具执行** 产生的 Traces、Metrics、Logs 上报到蓝鲸应用性能监控（APM）。

示例使用 tRPC Agent Go `v1.10.0`，通过 OTLP/HTTP 接入，覆盖初始化、配置、运行和生产注意事项。示例通过 OpenTelemetry `slog` Bridge 上报不包含 Prompt 和模型输出的结构化业务日志。

> ⚠️ **请先区分两个同名前缀的框架**：本示例接入的是 <a href="https://github.com/trpc-group/trpc-agent-go" target="_blank">tRPC-Agent（Agent 框架）</a>，而不是 <a href="https://github.com/trpc-group/trpc-go" target="_blank">tRPC-Go（RPC 框架）</a>。两者名称前缀一致，但一个是构建大模型 Agent 的框架，另一个是通用 RPC 通信框架。本文关注的是 Agent 框架自身的可观测。

## 0. 示例演示了什么可观测

本示例采用 **A2A（Agent-to-Agent）** 的 server-client 形态：

* `server/`：使用 `server/a2a` 把一个带工具的 Agent 暴露为可远程调用的服务。**Agent 在 server 进程内执行**（A2A Server 内部为该 Agent 构建 Runner 并运行），因此 Agent、模型调用、工具执行产生的 Span 和 GenAI 指标都在 server 端产生并上报。
* `client/`：使用 `agent/a2aagent` 读取远端 Agent Card，再用 Runner **定期发起远程调用**，消费返回的流式事件。client 默认按固定间隔循环请求（数据自生成），让 server 端持续产生可观测数据，便于在 APM 中直接观察到连续的调用链和指标。

因此，**这套示例主要体现的是 tRPC-Agent 框架自身的可观测**——Agent / LLM 调用 / Tool 执行的调用链和 GenAI 指标由框架内置埋点自动产生，业务代码无需手动埋点。OTel Provider 必须安装在 **server 端**，才能采集到这些信号。

## 1. 前置准备

### 1.1 术语介绍

* Traces：<a href="https://opentelemetry.io/docs/concepts/signals/traces/" target="_blank">调用链</a>，表示请求在应用程序的执行路径。
* Metrics：<a href="https://opentelemetry.io/docs/concepts/signals/metrics/" target="_blank">指标</a>，表示对运行服务的测量。
* Logs: <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">日志</a>，表示对事件的记录。
* Profiling: <a href="https://grafana.com/docs/pyroscope/latest/introduction/profiling/" target="_blank">性能分析</a>，表示对应用程序运行时资源的持续测量。
* Telemetry Data：观测数据，指代 Traces、Metrics、Logs、Profiling 等。
* APM：蓝鲸观测平台应用性能监控，提供四类观测数据开箱即用的观测能力。
* <a href="https://github.com/TencentBlueKing/bkmonitor-datalink/tree/main/pkg/collector" target="_blank">bk-collector</a>：腾讯蓝鲸的 APM 服务端组件，负责接收 Prometheus、OpenTelemetry、Jaeger、Skywalking 等主流开源组件的观测数据，并对数据进行清洗转发到观测平台链路。

此外，本示例涉及以下 Agent 观测概念：

* Agent Span：表示 Runner、Agent、模型调用或工具执行等操作。
* GenAI Metrics：以 `gen_ai.*` 为主的模型调用次数、耗时、首 Token 耗时和 Token 用量指标。
* A2A（Agent-to-Agent）：Agent 之间通过 Agent Card 相互发现、远程调用的协议。本示例用它拆分出 server 与 client。
* 高基数属性：取值数量很大的属性，例如用户 ID、会话 ID。直接作为指标维度可能增加存储和查询开销。

### 1.2 开发环境要求

在开始之前，请确保已经安装以下软件：

* Git。
* Docker 或其他兼容的容器工具。
* 一个 OpenAI API 兼容的模型服务及其 API Key。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/go-examples/trpc-agent-go
docker build -t trpc-agent-go-apm:latest -f server/Dockerfile .
# 可选：构建 client 镜像（定期请求 server，自生成数据）
docker build -t trpc-agent-go-apm-client:latest -f client/Dockerfile .
```

代码结构：

```text
trpc-agent-go/
├── server/            # A2A Server，Agent 在此执行，负责上报观测数据
│   ├── main.go        # server 入口：装 Provider、创建 Agent、启动 A2A Server
│   ├── telemetry.go   # Trace / Metric / Log Provider 初始化与关闭
│   ├── main_test.go   # 单元测试（配置校验、计算器工具、Provider 关闭）
│   └── Dockerfile
└── client/
    ├── main.go        # A2A Client：定期远程调用 server 暴露的 Agent，持续自生成数据
    └── Dockerfile
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

### 2.2 接入方式

tRPC Agent Go 已在 Runner、Agent、LLM、Tool 和 Workflow 等执行路径中创建 OpenTelemetry Span，并记录常用 GenAI 指标。这些内置埋点使用的是 **OTel 全局 Provider**：`telemetry/trace.Start` 会 `otel.SetTracerProvider`，`telemetry/metric.InitMeterProvider` 会注入框架全局 Meter。因此业务侧只需在创建 Agent 前完成以下初始化：

1. 使用 `telemetry/trace.Start` 安装 Trace Provider。
2. 使用 `telemetry/metric.NewMeterProvider` 创建 Metric Provider。
3. 调用 `telemetry/metric.InitMeterProvider`，让框架内部指标绑定到该 Provider。
4. 创建 Log Provider，并通过 OpenTelemetry `slog` Bridge 发送结构化日志。
5. 退出前关闭 3 个 Provider，确保批量数据被刷新。

上述初始化都发生在 **server 端**（`server/main.go` 的 `run` 会先调用 `setupTelemetry`，再创建 Agent、启动 A2A Server），因为 Agent 就在 server 进程执行。

完整实现请参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/trpc-agent-go/server/main.go" target="_blank">server/main.go</a> 和 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/trpc-agent-go/server/telemetry.go" target="_blank">server/telemetry.go</a>。

#### 两种接入姿势

本示例采用「代码内 OTLP 直连 APM」的方式，即在应用代码里显式创建 Provider 并直连蓝鲸 APM 的 OTLP/HTTP 端点。这是最通用、依赖最少的方式，也是本文档演示的方式。

如果你的服务本身还使用 **tRPC-Go（RPC 框架）** 承载流量，也可以改用 tRPC-Go 的插件式接入：在 `trpc_go.yaml` 中配置 `opentelemetry` 相关插件，由框架在启动时统一装配 OTel Provider，无需在代码中手写 `setupTelemetry`。此时只需保证插件装配的全局 Provider 在 Agent 创建之前就绪，tRPC-Agent 的内置埋点即可复用同一个 Provider。两种方式二选一即可，本示例为保持「最小可跑、零额外框架依赖」选择前者。

### 2.3 关键配置

🌟 请仔细阅读本章节，以确保观测数据能够准确上报到 APM。

#### 2.3.1 上报地址和应用 Token

请在创建 <a href="https://opentelemetry.io/docs/specs/otel/protocol/exporter/" target="_blank">Exporter</a> 时准确传入以下信息：

| 参数         | 说明                            |
|------------|-------------------------------|
| `endpoint` | 【必须】数据上报地址，请根据页面指引提供的接入地址进行填写。 |
| `x-bk-token`| 【必须】APM 应用 Token，作为 headers 传递。 |

Go 示例使用 OTLP/HTTP。`OTLP_ENDPOINT` 必须是 `host:port`，不包含 `http://`、`https://` 和 `/v1/traces` 等路径。SDK 会分别补充 `/v1/traces`、`/v1/metrics` 和 `/v1/logs`。

Trace 和 Log 导出器直接设置 Header，Metric 导出器通过 OpenTelemetry 标准环境变量读取同一 Token：

```go
header := "x-bk-token=" + url.QueryEscape(cfg.token)
os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", header)
os.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", header)

// atrace.Start 返回的关闭函数会复用这里传入的 Context；单独传 shutdownCtx
// 无法约束它。因此使用可取消的生命周期 Context，并在关闭截止时间到达时取消。
traceLifecycleCtx, cancelTraceLifecycle := context.WithCancel(ctx)
shutdownTrace, err := atrace.Start(
    traceLifecycleCtx,
    // ❗❗【非常重要】APM 接入指引提供的 HTTP OTLP 地址，格式为 host:port。
    atrace.WithEndpoint(cfg.otlpEndpoint),
    atrace.WithProtocol("http"),
    // ❗❗【非常重要】APM 应用 Token。
    atrace.WithHeaders(map[string]string{"x-bk-token": cfg.token}),
    atrace.WithServiceName(cfg.serviceName),
)
```

样例退出时会同时关闭 Trace、Metric 和 Log Provider，并为整个过程设置 10 秒截止时间；到期后会取消 `traceLifecycleCtx`，从而约束 `atrace.Start` 返回的 Trace 关闭函数。

* <a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#header-configuration" target="_blank">OpenTelemetry OTLP Header Configuration</a>

#### 2.3.2 服务信息和容器关联

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

示例将 `SERVICE_NAME` 映射为 `service.name`，并支持通过环境变量补充 Kubernetes 资源属性：

| 环境变量 | Resource 属性 | 说明 |
| --- | --- | --- |
| `SERVICE_NAME` | `service.name` | 【必须】APM 内的服务唯一标识。 |
| `SERVICE_NAMESPACE` | `service.namespace` | 【可选】服务命名空间。 |
| `SERVICE_VERSION` | `service.version` | 【可选】服务版本。 |
| `K8S_BCS_CLUSTER_ID` | `k8s.bcs.cluster.id` | 【可选】蓝鲸容器集群 ID。 |
| `K8S_POD_NAME` | `k8s.pod.name` | 【可选】Pod 名称。 |
| `K8S_NAMESPACE` | `k8s.namespace.name` | 【可选】Pod 所在命名空间。 |

#### 2.3.3 Provider 初始化

仅设置全局 OpenTelemetry Meter Provider 不足以启用 tRPC Agent Go 的内置指标。必须额外调用 `InitMeterProvider`：

```go
meterProvider, err := ametric.NewMeterProvider(
    ctx,
    ametric.WithEndpoint(cfg.otlpEndpoint),
    ametric.WithProtocol("http"),
    ametric.WithServiceName(cfg.serviceName),
)
if err != nil {
    return err
}
if err := ametric.InitMeterProvider(meterProvider); err != nil {
    return err
}
```

初始化必须发生在处理 Agent 请求之前，即在 A2A Server 启动之前。

Log Provider 使用相同 Resource、Endpoint 和 Token，并通过 `otelslog` 将 Go `slog` 日志桥接到 OpenTelemetry Logs：

```go
exporter, err := otlploghttp.New(
    ctx,
    // ❗❗【非常重要】APM 接入指引提供的 HTTP OTLP 地址，格式为 host:port。
    otlploghttp.WithEndpoint(cfg.otlpEndpoint),
    otlploghttp.WithInsecure(),
    // ❗❗【非常重要】APM 应用 Token。
    otlploghttp.WithHeaders(map[string]string{"x-bk-token": cfg.token}),
)
loggerProvider := sdklog.NewLoggerProvider(
    sdklog.WithResource(res),
    sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
)
logger := otelslog.NewLogger(appName, otelslog.WithLoggerProvider(loggerProvider))
```

* <a href="https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog" target="_blank">OpenTelemetry Go `slog` Bridge</a>

#### 2.3.4 模型配置

样例使用 OpenAI API 兼容接口。`MODEL_BASE_URL` 为空时使用 OpenAI 默认地址；使用其他兼容服务时填写完整 API Base URL。

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `MODEL_API_KEY` | 无 | 【必须】模型服务 API Key。 |
| `MODEL_BASE_URL` | 空 | 【可选】OpenAI API 兼容地址。 |
| `MODEL_NAME` | `gpt-4o-mini` | 模型名称。 |

模型 API Key 只用于访问模型服务，不要与蓝鲸 APM 的 `TOKEN` 混用。

#### 2.3.5 Server / Client 地址

| 环境变量 | 作用于 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `HOST` | server | `127.0.0.1:8080` | A2A Server 监听地址，容器内建议设为 `0.0.0.0:8080`。 |
| `TARGET` | client | `http://127.0.0.1:8080` | client 要连接的 A2A Server 地址。 |
| `PROMPT` | client | 内置计算题 | 每次远程 Agent 请求内容。 |
| `INTERVAL_SECONDS` | client | `30` | client 循环发起请求的间隔秒数（数据自生成）；设为 `0` 时只跑一次就退出，便于本地调试。 |

## 3. 使用场景

### 3.1 Traces

client 定期发起远程调用后，server 端框架会自动生成 Runner、Agent 和模型调用等 Span。使用 Tool 或 Workflow 时，还会产生对应的子 Span。调用链可用于分析：

* 一次 Agent 请求经过了哪些 Agent、模型和工具。
* 模型调用或工具执行的耗时分布。
* 错误发生在哪个执行阶段。
* 多 Agent、Graph 或 Workflow 的父子调用关系。

业务代码不需要为这些框架步骤重复创建 Span。需要补充业务步骤时，可以复用 `atrace.Tracer` 创建自定义 Span，并沿用 Runner 的 `context.Context`。

### 3.2 Metrics

tRPC Agent Go 内置指标包括：

| 指标 | 类型 | 说明 |
| --- | --- | --- |
| `trpc_agent_go.client.request_cnt` | Counter | Agent、模型或工具请求次数。 |
| `gen_ai.client.operation.duration` | Histogram | GenAI 操作（模型调用、工具执行等）耗时。 |
| `gen_ai.client.token.usage` | Histogram | Token 用量，按 `gen_ai.token.type` 区分输入、输出和缓存。 |
| `gen_ai.server.time_to_first_token` | Histogram | 流式响应首 Token 耗时。 |
| `trpc_agent_go.client.time_to_first_token` | Histogram | 客户端侧首 Token 耗时，与 `gen_ai.server.time_to_first_token` 同值。 |
| `trpc_agent_go.client.time_per_output_token` | Histogram | 平均每个输出 Token 的耗时。 |
| `trpc_agent_go.client.output_token_per_time` | Histogram | 单位时间的输出 Token 数（生成吞吐）。 |
| `gen_ai.workflow.elapsed_time` | Histogram | Workflow 生命周期区间耗时，仅在使用 Workflow 时产生。 |

这些指标由框架内置埋点自动记录，无需业务代码手动上报。其中 Histogram 类型在 APM 自定义指标页会展开为 `_bucket`、`_count`、`_sum`、`_min`、`_max` 等一组序列。示例默认每 30 秒导出一次指标，可通过 `OTEL_METRIC_EXPORT_INTERVAL` 调整，单位为毫秒。

### 3.3 Logs

样例通过 `otelslog` 将 `slog` 日志桥接到 OpenTelemetry Log Provider。server 在启动 A2A Server 时会写入一条结构化事件日志：

| 字段 | 说明 |
| --- | --- |
| `event.name` | 固定为 `a2a.server.starting`。 |
| `agent.name` | Agent 的稳定名称。 |
| `host` | A2A Server 监听地址。 |

日志由 OpenTelemetry SDK 附带 Trace 上下文。日志正文和属性不会包含 Prompt、模型输出、API Key 或 APM Token。tRPC Agent Go 框架自身的其他运行日志仍需通过标准输出采集或单独桥接。

## 4. 生产环境注意事项

### 4.1 采样

本示例依赖的 tRPC Agent Go `v1.10.0` 内置 `trace.Start` 使用 AlwaysSample，适合接入验证和低流量场景。在高流量生产环境中，应先评估数据量，再决定是否：

* 在框架中增加可配置 Sampler。
* 使用自定义 OpenTelemetry Trace Provider，并将框架的 `TracerProvider`、`Tracer` 绑定到该 Provider。
* 在采集链路后端执行尾部采样。

在未确认 SDK 实现前，不要仅设置采样环境变量并假设它一定覆盖框架的显式 Sampler。

### 4.2 提示词和模型输出安全

Agent Span 可能包含用户输入、模型输出、工具参数和工具结果。可通过设置 `OTEL_SPAN_ATTRIBUTE_VALUE_LENGTH_LIMIT=2048` 限制单个属性长度，但截断不等于脱敏。

上线前应确认数据合规要求，避免在提示词或工具参数中传入密码、Token、个人敏感信息。若业务必须处理敏感数据，应在进入 Agent 前脱敏，或在导出前增加属性过滤策略。

### 4.3 指标基数

用户 ID、会话 ID、IP、容器实例等字段具有高基数风险。不要将它们复制为自定义指标标签；优先保留在 Trace 中用于单次请求定位。模型名、Agent 名和 Tool 名应使用稳定、有限的取值集合。

## 5. 运行和查看数据

### 5.1 测试所需配置

无论接入本地观测后端还是真实蓝鲸 APM，都需要一个可用的 OpenAI API 兼容模型服务：

| 配置项 | 本地观测后端 | 真实蓝鲸 APM |
| --- | --- | --- |
| `TOKEN` | 可填写 `local-test`，本地 Collector 不校验。 | APM 应用接入指引提供的 Token。 |
| `OTLP_ENDPOINT` | Docker 内使用 `host.docker.internal:4318`。 | APM 应用接入指引提供的 HTTP OTLP `host:port`。 |
| `MODEL_API_KEY` | 模型服务 API Key；本地兼容服务按其要求填写。 | 模型服务 API Key。 |
| `MODEL_BASE_URL` | OpenAI 兼容模型服务 Base URL；使用 OpenAI 默认地址时留空。 | 同左。 |
| `MODEL_NAME` | 模型服务支持的模型名称。 | 同左。 |
| `SERVICE_NAME` | 建议使用 `trpc-agent-go-local`。 | 在目标 APM 应用内唯一。 |

### 5.2 运行单元测试

单元测试不需要 APM、Collector 或模型服务：

```shell
cd examples/go-examples/trpc-agent-go
go test ./...
go vet ./...
```

测试覆盖配置校验、布尔环境变量解析、计算器工具，以及 Trace、Metric、Log Provider 受关闭截止时间约束的行为。

### 5.3 使用本地观测后端验证

仓库内置 `ob-all-in-one`，包含 OpenTelemetry Collector、Jaeger、Prometheus、Grafana 和 OpenSearch。先启动本地观测后端：

```shell
cd examples/common/ob-all-in-one
docker compose up -d
```

构建并启动 A2A Server（Agent 在此执行并上报观测数据）：

```shell
cd ../../go-examples/trpc-agent-go
docker build -t trpc-agent-go-apm:latest -f server/Dockerfile .

docker run --rm --name trpc-agent-go-server \
  -p 8080:8080 \
  -e HOST="0.0.0.0:8080" \
  -e TOKEN="local-test" \
  -e OTLP_ENDPOINT="host.docker.internal:4318" \
  -e SERVICE_NAME="trpc-agent-go-local" \
  -e MODEL_API_KEY="<模型 API Key>" \
  -e MODEL_BASE_URL="<可选的 OpenAI 兼容 Base URL>" \
  -e MODEL_NAME="<模型名称>" \
  trpc-agent-go-apm:latest
```

如果 Linux 环境无法解析 `host.docker.internal`，在 `docker run` 后增加 `--add-host=host.docker.internal:host-gateway`。

Server 就绪后，在另一个终端运行 client 定期发起远程调用（client 不上报观测数据，只负责触发 server 端 Agent 执行）：

```shell
cd examples/go-examples/trpc-agent-go
go run ./client
# 每 30 秒自动请求一次，持续自生成数据；按 Ctrl+C 停止
# 只跑一次：INTERVAL_SECONDS=0 go run ./client
# 换个问题：PROMPT="Use the calculator to compute 25 * 4." go run ./client
```

本地查看入口：

* Jaeger Traces：<a href="http://localhost:16686" target="_blank">http://localhost:16686</a>。
* Prometheus Metrics：<a href="http://localhost:9090" target="_blank">http://localhost:9090</a>。
* Grafana：<a href="http://localhost:3000" target="_blank">http://localhost:3000</a>。
* OpenSearch Logs：<a href="http://localhost:5601" target="_blank">http://localhost:5601</a>。

验证结束后执行：

```shell
cd examples/common/ob-all-in-one
docker compose down
```

### 5.4 接入真实蓝鲸 APM

不需要自行搭建 Collector。先在蓝鲸 APM 创建应用，获取应用 Token 和 HTTP OTLP Endpoint，再启动 server：

```shell
docker run --rm --name trpc-agent-go-server \
  -p 8080:8080 \
  -e HOST="0.0.0.0:8080" \
  -e TOKEN="<APM 应用 Token>" \
  -e OTLP_ENDPOINT="<HTTP OTLP host:port>" \
  -e SERVICE_NAME="trpc-agent-go-demo" \
  -e MODEL_API_KEY="<模型 API Key>" \
  -e MODEL_BASE_URL="<可选的 OpenAI 兼容 Base URL>" \
  -e MODEL_NAME="gpt-4o-mini" \
  trpc-agent-go-apm:latest
```

随后用 client 定期发起远程调用触发 Agent 执行（`go run ./client`，默认每 30 秒一次，持续自生成数据）。server 常驻运行，Metrics 和 Logs 使用批量或周期导出，收到 `SIGINT`/`SIGTERM` 后会优雅关闭并刷新 Trace、Metric、Log Provider。

### 5.5 查看 Traces

Tracing 检索功能主要用于对分布式系统中的请求链路进行跟踪和分析，请参考<a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/explore_traces.md" target="_blank">「应用性能监控 APM/调用链追踪」</a> 进一步了解相关功能。

![](./images/trpc-agent-go-traces.png)

可以使用 `service.name=trpc-agent-go-demo` 过滤服务，再按 Agent、模型或操作名称分析 Span。

### 5.6 查看 Metrics

自定义指标功能旨在帮助用户针对特定应用及其服务进行深度性能指标监控，请参考<a href="#" target="_blank">「应用性能监控 APM/自定义指标」</a> 进一步了解相关功能。

![](./images/trpc-agent-go-metrics.png)

进入 APM 应用的自定义指标页面，在 `trpc_agent_go.internal.chat` 等 Meter 分组下检索 `gen_ai.client.operation.duration`、`gen_ai.client.token.usage`、`gen_ai.server.time_to_first_token` 和 `trpc_agent_go.client.request_cnt` 等框架内置指标，确认 GenAI 观测数据已进入目标 APM 应用。Histogram 类型指标会展开为 `_bucket`、`_count`、`_sum` 等序列，属正常现象。

### 5.7 查看 Logs

日志功能主要用于查看和分析对应服务（应用程序）运行过程中产生的各类日志信息，请参考<a href="#" target="_blank">「应用性能监控 APM/日志分析」</a> 进一步了解相关功能。

![](./images/trpc-agent-go-logs.png)

进入 APM 应用的日志分析页面，使用 `service.name=trpc-agent-go-demo` 和 `event.name=a2a.server.starting` 过滤，确认 server 端结构化日志已上报。

## 6. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>
* <a href="https://github.com/trpc-group/trpc-agent-go" target="_blank">tRPC Agent Go</a>
* <a href="https://github.com/trpc-group/trpc-agent-go/tree/main/server/a2a" target="_blank">tRPC Agent Go A2A Server</a>
* <a href="https://opentelemetry.io/docs/languages/go/" target="_blank">OpenTelemetry Go</a>
* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/trpc-agent-go" target="_blank">本接入样例源码</a>
* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/docs/open/go/otlp/README.md" target="_blank">Go（OpenTelemetry SDK）接入</a>