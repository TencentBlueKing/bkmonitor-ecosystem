# Go（Jaeger - OpenTracing Bridge）接入

本指南将帮助使用 Jaeger client 上报数据的用户，平滑过渡到使用 OpenTelemetry SDK 接入蓝鲸应用性能监控，以 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/jaeger-ot-demo/" target="_blank">入门项目-jaeger-ot-demo</a> 为例，介绍调用链接入及 SDK 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速接入并体验蓝鲸应用性能监控相关功能。

## 1. 前置准备

### 1.1 术语介绍

{{TERM_INTRO}}

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：
* Git
* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/go-examples/jaeger-ot-demo
docker build -t jaeger-ot-demo-go:latest .
```

## 2. 快速接入

Jaeger Client 是 OpenTracing API 规范的具体实现，目前业界标准已从 OpenTracing 演进为 OpenTelemetry（融合 OpenTracing 和 OpenCensus）。本文详解 Jaeger Client 到 OpenTelemetry SDK 的最小化迁移方案，我们也通过 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/jaeger-ot-demo/Dockerfile" target="_blank">Dockerfile</a> 演示了这一过程。

### 2.1 创建应用

{{APPLICATION_SET_UP}}

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

样例代码 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/jaeger-ot-demo/service/otlp.go" target="_blank">jaeger-ot-demo service/otlp.go</a> 只演示上报 Traces 数据的配置，完整的配置可以参考 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/helloworld/service/otlp/otlp.go" target="_blank">helloworld service/otlp/otlp.go</a> 进行接入。

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

{{MUST_CONFIG_EXPORTER}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/jaeger-ot-demo/service/otlp.go" target="_blank">service/otlp.go setUpTraces</a> 提供了创建样例：

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

{{MUST_CONFIG_RESOURCES}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/jaeger-ot-demo/service/otlp.go" target="_blank">service/otlp.py newResource</a> 提供了创建样例：

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

当前示例项目使用 bridgeTracer 替换原有 Jaeger Client tracer 后，tracer 的使用无需任何改变。如果想要学习 Jaeger Client tracer 的使用方式，可参考 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/jaeger-client-demo/README.md" target="_blank">jaeger-client-demo「3. 使用场景」</a>实现。

## 4. 快速体验

### 4.1 运行样例

#### 4.1.1 运行

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="{{access_config.token}}" \
-e SERVICE_NAME="{{service_name}}" \
-e OTLP_ENDPOINT="{{access_config.otlp.http_endpoint}}" \
-e ENABLE_TRACES="{{access_config.otlp.enable_traces}}" jaeger-ot-demo-go:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

#### 4.1.2 运行参数说明

{{JAEGER_DEMO_RUN_PARAMETERS}}

### 4.2 查看数据

#### 4.2.1 Traces 检索

{{VIEW_TRACES_DATA}}

## 5. 了解更多

{{LEARN_MORE}}