# Go（Jaeger Client）接入

本指南将帮助您使用 Jaeger client 接入蓝鲸应用性能监控，以 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/jaeger-client-demo/" target="_blank">入门项目-jaeger-client-demo</a> 为例，介绍调用链接入及 SDK 使用场景。

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
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/go-examples/jaeger-client-demo
docker build -t jaeger-client-demo-go:latest .
```

## 2. 快速接入

Jaeger Client 是 OpenTracing API 规范的具体实现，目前业界标准已从 OpenTracing 演进为 OpenTelemetry（融合 OpenTracing 和 OpenCensus），推荐从 Jaeger client 迁移到 OTel SDK。本文详解 Jaeger Client 通过 HTTP Thrift 协议上报的方案，我们也通过 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/jaeger-client-demo/Dockerfile" target="_blank">Dockerfile</a> 演示了这一过程。

### 2.1 创建应用

{{APPLICATION_SET_UP}}

### 2.2 关键配置（上报地址 & 应用 Token）

🌟 请仔细阅读本章节，以确保观测数据能准确上报到 APM。

初始化 Jaeger Tracer 时，需通过 CollectorEndpoint 配置上报地址，并在 HTTPHeaders 中添加认证 Header。

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/jaeger-client-demo/main.go" target="_blank">jaeger-client-demo/main.go initJaeger</a> 提供了创建样例：

```go
collectorEndpoint := ""
if strings.HasPrefix(conf.BKEndpoint, "http://") || strings.HasPrefix(conf.BKEndpoint, "https://") {
    collectorEndpoint = fmt.Sprintf("%v/jaeger/v1/traces", conf.BKEndpoint)
} else {
    collectorEndpoint = fmt.Sprintf("http://%v/jaeger/v1/traces", conf.BKEndpoint)
}
cfg := jaegercfg.Configuration{
    Reporter: &jaegercfg.ReporterConfig{
        HTTPHeaders: map[string]string{
        	"x-bk-token": conf.Token,
        },
        CollectorEndpoint: collectorEndpoint,
    },
}
```

## 3. 使用场景

当前示例项目聚焦于 Jaeger Client 的 Traces 应用场景，集中在：

```go
func(hws *HelloWorldService) helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 从 HTTP 请求中提取追踪上下文
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header),
	)

	// 2. 创建 Span（如果存在上游上下文则继承）
	var span opentracing.Span
	if err == nil {
		span = opentracing.StartSpan("helloWorldHandler", ext.RPCServerOption(wireContext))
	} else {
		span = opentracing.StartSpan("helloWorldHandler")
	}
	defer span.Finish()

	// 3. 将 Span 放入上下文
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	// Traces（调用链）- 自定义 Span
	hws.traces_custom_span_demo(ctx)
	// Traces（调用链）- Span 事件
	hws.traces_span_event_demo(ctx)
	// Traces（调用链）- 模拟错误
	hws.tracesRandomErrorDemo(ctx)
}
```

### 3.1 获取 Tracer

在 Jaeger 初始化完成后，通过 OpenTracing 标准接口将其注册为全局追踪器：

```go
opentracing.SetGlobalTracer(tracer)
```

当应用程序需要获取追踪器实例时，通过 OpenTracing 的统一访问点获取全局追踪器：​

```go
opentracing.GlobalTracer()
```

### 3.2 创建 Span

当操作未关联到现有追踪上下文时，可直接创建新的根级 Span：​

```go
span = opentracing.StartSpan("helloWorldHandler")
defer span.Finish()
```

当操作处于特定调用链中，应基于上下文创建关联 Span：

```go
span, _ := opentracing.StartSpanFromContext(ctx, "traces_custom_span_demo")
defer span.Finish()
```

在客户端发送 HTTP 请求时，将分布式追踪上下文注入到请求头：

```go
span := tracer.StartSpan("Caller/queryHelloWorld")
defer span.Finish()
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
err = tracer.Inject(
    span.Context(),
    opentracing.HTTPHeaders,
    opentracing.HTTPHeadersCarrier(req.Header),
)
```

在服务端接收 HTTP 请求时，从请求头中提取分布式追踪上下文，并创建关联的 Span：
```go
wireContext, err := opentracing.GlobalTracer().Extract(
    opentracing.HTTPHeaders,
    opentracing.HTTPHeadersCarrier(r.Header),
)
span = opentracing.StartSpan("helloWorldHandler", ext.RPCServerOption(wireContext))
```

### 3.3 设置属性

通过 Span 的 SetTag 方法设置自定义追踪属性：​

```go
span.SetTag("helloworld.kind", 2)
span.SetTag("helloworld.step", "traces_custom_span_demo")
```

### 3.4 设置事件

通过 Span 的 LogKV 方法设置自定义事件属性：​

```go
span.LogKV("helloworld.kind", 3)
span.LogKV("helloworld.step", "traces_span_event_demo")
```

### 3.5 设置异常事件

通过 Span 记录异常信息：

```go
import "github.com/opentracing/opentracing-go/ext"
ext.Error.Set(span, true)
span.LogFields(
    otlog.String("event", "error"), // 记录一个 error 事件
    otlog.String("error.message", err.Error()),
    otlog.String("error.type", fmt.Sprintf("%T", err)),
)
```

## 4. 快速体验

### 4.1 运行样例

#### 4.1.1 运行

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="{{access_config.token}}" \
-e SERVICE_NAME="{{service_name}}" \
-e OTLP_ENDPOINT="{{access_config.otlp.http_endpoint}}" \
-e ENABLE_TRACES="{{access_config.otlp.enable_traces}}" jaeger-client-demo-go:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

#### 4.1.2 运行参数说明

{{JAEGER_DEMO_RUN_PARAMETERS}}

### 4.2 查看数据

#### 4.2.1 Traces 检索

{{VIEW_TRACES_DATA}}

## 5. 了解更多

{{LEARN_MORE}}