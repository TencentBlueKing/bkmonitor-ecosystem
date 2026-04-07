# Go（Jaeger - OpenTracing Bridge）接入

本指南将帮助使用 Jaeger client 上报数据的用户，平滑过渡到使用 OpenTelemetry SDK 接入蓝鲸应用性能监控，以 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/jaeger-ot-demo/" target="_blank">入门项目-jaeger-ot-demo</a> 为例，介绍调用链接入及 SDK 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速接入并体验蓝鲸应用性能监控相关功能。

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
cd bkmonitor-ecosystem/examples/go-examples/jaeger-ot-demo
docker build -t jaeger-ot-demo-go:latest .
```

## 2. 快速接入

Jaeger Client 是 OpenTracing API 规范的具体实现，目前业界标准已从 OpenTracing 演进为 OpenTelemetry（融合 OpenTracing 和 OpenCensus）。本文详解 Jaeger Client 到 OpenTelemetry SDK 的最小化迁移方案，我们也通过 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/jaeger-ot-demo/Dockerfile" target="_blank">Dockerfile</a> 演示了这一过程。

### 2.1 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

### 2.2 环境依赖

安装 OpenTelemetry API、OpenTelemetry SDK 和 OpenTracing Bridge，使用 OpenTracing Bridge 作为 OpenTracing 和 OpenTelemetry 之间的适配器。

```go
import (
	"go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	otelBridge "go.opentelemetry.io/otel/bridge/opentracing"
)
```
- 迁移完成后可移除 jaeger-client 依赖项，但应在过渡期保持其作为备用采集通道，通过流量灰度控制策略（如按服务版本/用户群分流）实现数据双向验证，保障可观测数据的完整性和一致性。
- <a href="https://github.com/open-telemetry/opentelemetry-go/tree/main/bridge/opentracing" target="_blank">OpenTracing Bridge</a> 提供双向适配层，维持 OpenTracing API 兼容性的同时桥接 OpenTelemetry SDK。

### 2.3 OpenTelemetry SDK 配置

OpenTelemetry 提供标准化的框架和工具包，用于创建和管理 Traces、Metrics、Logs 数据。示例项目集成 OpenTelemetry SDK 并将观测数据发送到 bk-collector。

样例代码 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/jaeger-ot-demo/service/otlp.go" target="_blank">jaeger-ot-demo service/otlp.go</a> 只演示上报 Traces 数据的配置，完整的配置可以参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/helloworld/service/otlp/otlp.go" target="_blank">helloworld service/otlp/otlp.go</a> 进行接入。

### 2.4 项目 tracer 修改

将 Jaeger tracer 替换为通过 OpenTelemetry 提供的 OpenTracing Bridge 实现的 tracer，在确保向后兼容性的前提下接入现代可观测性体系。

```go
// 初始化 Jaeger 追踪器
func initJaeger(serviceName string, conf *config.Config, otelService *service.OTelService, ctx context.Context) (opentracing.Tracer, func(), error) {
	// 1. 创建 Jaeger 配置
	// cfg := jaegercfg.Configuration{
	// 	ServiceName: serviceName, // 服务名称
	// 	Sampler: &jaegercfg.SamplerConfig{
	// 		Type:  jaeger.SamplerTypeConst, // 采样类型：全部采样
	// 		Param: 1,                       // 采样率：1=100%
	// 	},
	// 	Reporter: &jaegercfg.ReporterConfig{
	// 		LogSpans:           true,
	// 		// HTTPHeaders: map[string]string{
	// 		// 	"x-bk-token": conf.Token,
	// 		// },

	// 		CollectorEndpoint: collectorEndpoint,
	// 		BufferFlushInterval: 500 * time.Millisecond,
	// 	},
	// }

	// 2. 创建追踪器
	// tracer, closer, err := cfg.NewTracer(
	// 	jaegercfg.Logger(jaeger.StdLogger), // 使用标准日志
	// 	jaegercfg.Tag("bk.data.token", conf.Token),

	// )
	var err error
	if err = otelService.Init(conf, ctx); err != nil {
		log.Printf("[%v] failed to init: %v", otelService.Type(), err)
		return nil, nil, err
	}
	if err = otelService.Start(); err != nil {
		log.Printf("[%v] failed to start: %v", otelService.Type(), err)
		return nil, nil, err
	}
	tracerProvider := otel.GetTracerProvider()
	otelTracer := tracerProvider.Tracer("tracer_name")
	// Use the bridgeTracer as your OpenTracing tracer.
	bridgeTracer, wrapperTracerProvider := otelBridge.NewTracerPair(otelTracer)
	// Set the wrapperTracerProvider as the global OpenTelemetry
	// TracerProvider so instrumentation will use it by default.
	otel.SetTracerProvider(wrapperTracerProvider)
	// 设置为全局追踪器
	opentracing.SetGlobalTracer(bridgeTracer)

	// 返回追踪器和关闭函数
	return bridgeTracer, func() {
		if err := otelService.Stop(); err != nil {
			log.Printf("[%v] failed to stop: %v", otelService.Type(), err)
		}
	}, nil
}
```

### 2.5 关键配置

🌟 请仔细阅读本章节，以确保观测数据能准确上报到 APM。

#### 2.5.1 上报地址 & 应用 Token

请在创建 <a href="https://opentelemetry.io/docs/specs/otel/protocol/exporter/" target="_blank">Exporter</a> 时准确传入以下信息：

| 参数         | 说明                            |
|------------|-------------------------------|
| `endpoint` | 【必须】数据上报地址，请根据页面指引提供的接入地址进行填写。 |
| `x-bk-token`| 【必须】APM 应用 Token，作为 headers 传递。 |

示例项目在 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/jaeger-ot-demo/service/otlp.go" target="_blank">service/otlp.go setUpTraces</a> 提供了创建样例：

```go
// newTracerExporter Initialize a new tracer exporter based on ExporterType
func (ots *OTelService) newTracerExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	switch ots.ExporterType {
	case config.ExporterHttp:
		return newHttpTracerExporter(
			ctx,
			// ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
			// 格式为 ip:port 或 domain:port，不要带 schema
			ots.Endpoint,
			// ❗❗【非常重要】请传入应用 Token
			map[string]string{"x-bk-token": ots.Token},
		)
	case config.ExporterGRPC:
		return newGRPCTracerExporter(ctx, ots.gRPCConn, map[string]string{"x-bk-token": ots.Token})
	}
	return nil, fmt.Errorf("[%v] invalid exporter type", ots.Type())
}
```

#### 2.5.2 服务信息

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

示例项目在 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/jaeger-ot-demo/service/otlp.go" target="_blank">service/otlp.py newResource</a> 提供了创建样例：

```go
func (ots *OTelService) newResource() (*resource.Resource, error) {
	extraRes, err := resource.New(
		ots.ctx,
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			// ❗❗【非常重要】应用服务唯一标识
			semconv.ServiceNameKey.String(ots.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// resource.Default() 提供了部分 SDK 默认属性
	res, err := resource.Merge(resource.Default(), extraRes)
	if err != nil {
		return nil, err
	}

	return res, nil
}
```

## 3. 使用场景

当前示例项目使用 bridgeTracer 替换原有 Jaeger Client tracer 后，tracer 的使用无需任何改变。如果想要学习 Jaeger Client tracer 的使用方式，可参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/jaeger-client-demo/README.md" target="_blank">jaeger-client-demo「3. 使用场景」</a>实现。

## 4. 快速体验

### 4.1 运行样例

#### 4.1.1 运行

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="xxx" \
-e SERVICE_NAME="helloworld" \
-e OTLP_ENDPOINT="http://127.0.0.1:4318" \
-e ENABLE_TRACES="true" jaeger-ot-demo-go:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

#### 4.1.2 运行参数说明

运行参数说明：

| 参数                   | 推荐值                                | 说明                                        |
|----------------------|--------------------------------------|-------------------------------------------|
| `TOKEN`              | `""`                                 | APM 应用 `Token`。                            |
| `SERVICE_NAME`       | `"jaeger-client-demo-go"`                       | 服务唯一标识，一个应用可以有多个服务，通过该属性区分。                |
| `OTLP_ENDPOINT`      | `"http://127.0.0.1:4318"` | OT 数据上报地址，请根据页面指引提供的接入地址进行填写，支持以下协议：<br />`HTTP`：`http://127.0.0.1:4318`（demo 使用该协议演示上报）<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。        |
| `ENABLE_TRACES`      | `false`                              | 是否启用调用链上报。                                 |

### 4.2 查看数据

#### 4.2.1 Traces 检索

Tracing 检索功能主要用于对分布式系统中的请求链路进行跟踪和分析，请参考<a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/explore_traces.md" target="_blank">「应用性能监控 APM/调用链追踪」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/traces.png)

## 5. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>