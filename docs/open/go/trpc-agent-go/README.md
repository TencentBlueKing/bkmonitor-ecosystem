# Go（tRPC Agent Go）接入

本指南介绍如何将 tRPC Agent Go 的 Traces 和 Metrics 上报到蓝鲸应用性能监控（APM）。

tRPC Agent Go 已在 Runner、Agent、LLM 和 Tool 等执行路径中实现 OpenTelemetry 埋点。业务只需在创建 Agent 前初始化 Trace 和 Metric Provider，无需重复插桩。本文使用 OTLP/HTTP 上报。

## 1. 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

## 2. 接入

### 2.1 添加依赖

在 `go.mod` 中添加 tRPC Agent Go 和 tRPC-Go：

```go
require (
    trpc.group/trpc-go/trpc-agent-go v1.10.1-0.20260708011736-814595b55df0
    trpc.group/trpc-go/trpc-go v1.1.0
)
```

### 2.2 初始化 SDK

在创建 Agent 前初始化 Provider。以下代码同时启用 Traces 和 Metrics：

```go
import (
    "context"
    "errors"
    "fmt"
    "net/url"
    "os"
    "strings"

    ametric "trpc.group/trpc-go/trpc-agent-go/telemetry/metric"
    atrace "trpc.group/trpc-go/trpc-agent-go/telemetry/trace"
)

const serviceName = "trpc.test.trpcagent.api"

func setupTelemetry(ctx context.Context) (func() error, error) {
    endpoint := strings.TrimSpace(os.Getenv("OTLP_ENDPOINT"))
    if endpoint == "" {
        return nil, errors.New("OTLP_ENDPOINT is required")
    }
    token := strings.TrimSpace(os.Getenv("TOKEN"))
    if token == "" {
        return nil, errors.New("TOKEN is required")
    }
    otelServiceName := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
    if otelServiceName == "" {
        otelServiceName = serviceName
    }

    header := "x-bk-token=" + url.QueryEscape(token)
    if err := os.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", header); err != nil {
        return nil, fmt.Errorf("set metrics headers: %w", err)
    }
    meterProvider, err := ametric.NewMeterProvider(
        ctx,
        // ❗❗【非常重要】请填写 APM 接入指引提供的 HTTP OTLP 地址，格式为 host:port。
        ametric.WithEndpoint(endpoint),
        ametric.WithProtocol("http"),
        ametric.WithServiceName(otelServiceName),
    )
    if err != nil {
        return nil, fmt.Errorf("create meter provider: %w", err)
    }
    if err := ametric.InitMeterProvider(meterProvider); err != nil {
        shutdownErr := meterProvider.Shutdown(ctx)
        return nil, errors.Join(fmt.Errorf("initialize meter provider: %w", err), shutdownErr)
    }

    shutdownTrace, err := atrace.Start(
        ctx,
        // ❗❗【非常重要】请填写 APM 接入指引提供的 HTTP OTLP 地址，格式为 host:port。
        atrace.WithEndpoint(endpoint),
        atrace.WithProtocol("http"),
        // ❗❗【非常重要】请将 APM 应用 Token 作为 x-bk-token Header 传入。
        atrace.WithHeaders(map[string]string{"x-bk-token": token}),
        atrace.WithServiceName(otelServiceName),
    )
    if err != nil {
        shutdownErr := meterProvider.Shutdown(ctx)
        return nil, errors.Join(fmt.Errorf("start trace provider: %w", err), shutdownErr)
    }

    return func() error {
        return errors.Join(
            shutdownTrace(),
            meterProvider.Shutdown(context.Background()),
        )
    }, nil
}
```

在 `main` 中先初始化 Provider，再创建 Agent：

```go
ctx := context.Background()
shutdownTelemetry, err := setupTelemetry(ctx)
if err != nil {
    log.Fatalf("failed to set up telemetry: %v", err)
}
defer func() {
    if err := shutdownTelemetry(); err != nil {
        log.Errorf("failed to shut down telemetry: %v", err)
    }
}()

agent := newAgent()
```

### 2.3 关联配置

| 环境变量 | 必需 | 推荐值 | 说明 |
| --- | --- | --- | --- |
| `OTLP_ENDPOINT` | 是 | `"127.0.0.1:4318"` | APM 接入指引提供的 HTTP OTLP 地址，格式为 `host:port`，不包含 `http://`。其他环境、跨云场景请根据页面接入指引填写。 |
| `TOKEN` | 是 | `"xxx"` | APM 接入指引提供的应用 Token。 |
| `SERVICE_NAME` | 否 | `"helloworld"` | APM 服务名；未设置时使用代码中的 `serviceName`。 |
| `OPENAI_API_KEY` | 是 | `<模型 API Key>` | OpenAI 兼容模型服务的 API Key。 |
| `OPENAI_BASE_URL` | 否 | `<OpenAI 兼容 Base URL>` | OpenAI 兼容模型服务的 Base URL。 |

APM 不需要额外的 tRPC-Go Plugin 配置，服务继续使用常规 `trpc_go.yaml`。

### 2.4 运行 Demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/go-examples/trpc-agent-go
docker build -t trpc-agent-go-apm:latest .

docker run --rm --name trpc-agent-go-demo \
  -p 8080:8080 \
  -e OTLP_ENDPOINT="127.0.0.1:4318" \
  -e TOKEN="xxx" \
  -e SERVICE_NAME="helloworld" \
  -e OPENAI_API_KEY="<模型 API Key>" \
  -e OPENAI_BASE_URL="<可选的 OpenAI 兼容 Base URL>" \
  trpc-agent-go-apm:latest
```

Demo 会启动 tRPC-Agent HTTP 服务，并通过 `loopQuery` 周期请求本地 Agent，持续产生 Traces 和 Metrics。

### 2.5 LLM 指标

tRPC Agent Go 会自动上报以下指标：

| 指标 | 说明 |
| --- | --- |
| `trpc_agent_go.client.request_cnt` | Agent、模型或工具请求次数。 |
| `gen_ai.client.operation.duration` | GenAI 操作耗时。 |
| `gen_ai.client.token.usage` | 输入、输出和缓存 Token 用量。 |
| `gen_ai.server.time_to_first_token` | 流式响应首 Token 耗时。 |
| `trpc_agent_go.client.time_per_output_token` | 平均每个输出 Token 的耗时。 |
| `trpc_agent_go.client.output_token_per_time` | 单位时间的输出 Token 数。 |
| `gen_ai.workflow.elapsed_time` | Workflow 生命周期区间耗时。 |

## 3. 查看数据

| 查看调用链 | 查看指标 |
| --- | --- |
| ![查看调用链](./images/trpc-agent-go-traces.png) | ![查看指标](./images/trpc-agent-go-metrics.png) |

## 4. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>
* <a href="https://github.com/trpc-group/trpc-agent-go" target="_blank">tRPC Agent Go</a>
* <a href="https://github.com/trpc-group/trpc-go" target="_blank">tRPC-Go</a>
* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/master/examples/go-examples/trpc-agent-go" target="_blank">本接入样例源码</a>