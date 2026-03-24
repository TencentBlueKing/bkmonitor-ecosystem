# Go（OpenTelemetry SDK）接入

{{OVERVIEW}}

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
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/go-examples/helloworld
docker build -t helloworld-go:latest .
```

## 2. 快速接入

### 2.1 创建应用

{{APPLICATION_SET_UP}}

### 2.2 开箱即用 SDK 接入示例

OpenTelemetry 提供标准化的框架和工具包，用于创建和管理 Traces、Metrics、Logs 数据。

示例项目提供集成 OpenTelemetry Go SDK 并将观测数据发送到 bk-collector 的方式，可以参考 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/helloworld/service/otlp/otlp.go" target="_blank">service/otlp/otlp.go</a> 进行接入。

### 2.3 关键配置

🌟 请仔细阅读本章节，以确保观测数据能准确上报到 APM。

#### 2.3.1 上报地址 & 应用 Token

{{MUST_CONFIG_EXPORTER}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/helloworld/service/otlp/otlp.go" target="_blank">service/otlp/otlp.go newTracerExporter</a> 提供了创建样例：

```go
// setUpTraces
func (s *Service) setUpTraces(ctx context.Context, res *resource.Resource) error {
	tracerExporter, err := s.newTracerExporter(ctx)
	if err != nil {
		return err
	}
	s.tracerProvider = newTracerProvider(res, tracerExporter)
	otel.SetTextMapPropagator(newPropagator())
	return nil
}

// newHttpTracerExporter Initialize a new HTTP tracer exporter.
func newHttpTracerExporter(ctx context.Context, endpoint string, headers map[string]string) (*otlptrace.Exporter, error) {
	return otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithHeaders(headers),
	)
}

func (s *Service) newTracerExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	return newHttpTracerExporter(
		ctx,
		// ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
		// 格式为 ip:port 或 domain:port，不要带 schema
		s.config.Endpoint,
		// ❗❗【非常重要】请传入应用 Token
		map[string]string{"x-bk-token": s.config.Token},
	)
}
```

指标、日志的配置方式和上述一致，请参考 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/helloworld/service/otlp/otlp.go" target="_blank">service/otlp/otlp.go</a> 中的 `newMeterExporter`、`newLoggerExporter` 函数。

`x-bk-token` 也可以通过「环境变量」的方式进行配置：

```shell
export OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo"
```

配置优先级：SDK > 环境变量，更多请参考 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#header-configuration" target="_blank">Header Configuration</a>。

#### 2.3.2 服务信息

{{MUST_CONFIG_RESOURCES}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/helloworld/service/otlp/otlp.go" target="_blank">service/otlp/otlp.go newResource</a> 提供了创建样例：

```go
func (s *Service) newResource() (*resource.Resource, error) {
	extraRes, err := resource.New(
		...
		resource.WithAttributes(
			// ❗❗【非常重要】应用服务唯一标识
			semconv.ServiceNameKey.String(s.config.ServiceName),
		),
	)
	// resource.Default() 提供了部分 SDK 默认属性
	res, err := resource.Merge(resource.Default(), extraRes)
	return res, nil
}
```

## 3. 使用场景

示例项目整理常见的使用场景，集中在：

```go
func HelloWorld(w http.ResponseWriter, req *http.Request) {
	ctx, span := tracer.Start(req.Context(), "Handle/HelloWorld")
	defer span.End()

	// Logs（日志）
	logsDemo(ctx, req)

	country := choiceCountry()
	logger.InfoContext(ctx, fmt.Sprintf("get country -> %s", country))

	// Metrics（指标） - Counter 类型
	metricsCounterDemo(ctx, country)
	// Metrics（指标） - Histograms 类型
	metricsHistogramDemo(ctx)

	// Traces（调用链）- 自定义 Span
	tracesCustomSpanDemo(ctx)
	// Traces（调用链）- Span 事件
	tracesSpanEventDemo(ctx)
	// Traces（调用链）- 模拟错误
	if err := tracesRandomErrorDemo(ctx, span); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	greeting := generateGreeting(country)
	w.Write([]byte(greeting))
}
```

可以结合代码和下方说明进行使用：<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/go-examples/helloworld/service/http/helloworld.go" target="_blank">service/http/helloworld.go</a>。

### 3.1 Traces

#### 3.1.1 创建 Resource

Resource 代表观测数据所属的资源实体。

例如运行在 Kubernetes 上的容器所生成的观测数据，具有进程名称、Pod 名称等资源实体信息。

可以通过 `resource.Detector` 自动检测正在运行的进程、所在操作系统等资源属性信息：

```go
func (s *Service) newResource() (*resource.Resource, error) {
	extraRes, err := resource.New(
		s.ctx,
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			// ❗❗【非常重要】应用服务唯一标识
			semconv.ServiceNameKey.String(s.config.ServiceName),
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

* <a href="https://opentelemetry.io/docs/languages/go/resources/" target="_blank">Resources</a>

#### 3.1.2 创建 Span

Span 是 Traces 的构建块，代表一个工作或操作单元。

Span 通过 `otel.Tracer` 进行创建，<a href="https://opentelemetry.io/docs/specs/otel/trace/api/" target="_blank">`Tracer`</a> 是一个用于创建和管理 Span 的对象。它提供了 API 接口，开发人员可以用它来在应用程序代码中生成和记录 Span。

**后续样例提及的 `tracer` 创建方式如下：**

```go
import (
    "go.opentelemetry.io/otel"
)

const Name = "helloworld"
var tracer = otel.Tracer(Name)
````

需要将 `ctx` 以参数传入，以便获取父 Span 已设置的上下文，示例代码如下：

```go
import (
    "go.opentelemetry.io/otel"
)

const Name = "helloworld"
var tracer = otel.Tracer(Name)

// tracesCustomSpanDemo Traces（调用链）- 增加自定义 Span
func tracesCustomSpanDemo(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "CustomSpanDemo/doSomething")
	defer span.End()

	doSomething(50)
}
```

* <a href="https://opentelemetry.io/docs/languages/go/instrumentation/#creating-spans" target="_blank">Creating Spans</a>

#### 3.1.3 设置属性

Attributes（属性）是 Span 元数据，以 Key-Value 形式存在。

在 Span 设置属性，对问题定位、过滤、聚合非常有帮助。

```go
// 增加 Span 自定义属性
span.SetAttributes(
    attribute.Int("helloworld.kind", 1),
    attribute.String("helloworld.step", "tracesCustomSpanDemo"),
)
```

* <a href="https://opentelemetry.io/docs/languages/go/instrumentation/#span-attributes" target="_blank">Span Attributes</a>

#### 3.1.4 设置事件

Event（事件）是一段人类可读信息，用于记录 Span 生命周期内发生的事情。

```go
// tracesSpanEventDemo Traces（调用链）- Span 事件
func tracesSpanEventDemo(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "SpanEventDemo/doSomething")
	defer span.End()

	span.AddEvent("Before doSomething")
	doSomething(50)
	span.AddEvent("After doSomething")
}
```

* <a href="https://opentelemetry.io/docs/languages/go/instrumentation/#events" target="_blank">Span Events</a>

#### 3.1.5 记录错误

当一个 Span 出现错误，可以对其进行错误记录。

```go
func tracesRandomErrorDemo(ctx context.Context, span trace.Span) error {
	if err := randErr(0.1); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}
```

* <a href="https://opentelemetry.io/docs/languages/go/instrumentation/#record-errors" target="_blank">Record errors</a>

#### 3.1.6 设置状态

当一个 Span 未能成功，可以通过设置状态进行显式标记。

```go
span.SetStatus(codes.Error, err.Error())
```
* <a href="https://opentelemetry.io/docs/languages/go/instrumentation/#set-span-status" target="_blank">Set span status</a>

#### 3.1.7 在当前 Span 上设置自定义属性

在部分场景下，Span 可能在框架入口、中间件等位置便被创建，如果你希望在当前的 Span 设置属性，而不是新创建一个 Span，可以通过以下方式进行：

```go
import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// setCustomSpanAttributes Traces（调用链）- 在当前 Span 上设置自定义属性
func tracesSetCustomSpanAttributes(ctx context.Context) {
	currentSpan := trace.SpanFromContext(ctx)
	currentSpan.SetAttributes(attribute.String("ApiName", "ApiRequest"), attribute.Int("actId", 12345))
}
```

### 3.2 Metrics

#### 3.2.1 创建 Meter

<a href="https://opentelemetry.io/docs/specs/otel/metrics/api/" target="_blank">`Meter`</a> 是一个负责创建 Instruments 的对象。它提供了 API 接口，允许开发人员在代码中定义和记录 Metrics。

后续样例提及的 `meter` 创建方式如下：

```go
import (
    "go.opentelemetry.io/otel"
)

const Name = "helloworld"
var meter = otel.Meter(Name)
```

#### 3.2.2 Counters

Counters（计数器）用于记录非负递增值。

例如，可以通过以下方式上报请求总数：

```go
//【建议】初始化指标再使用，而不是在业务逻辑里初始化
func init() {
    requestsTotal, err = meter.Int64Counter("requests_total", metric.WithDescription("Total number of HTTP requests"))
}

// metricsCounterDemo Metrics（指标）- 使用 Counter 类型指标
func metricsCounterDemo(ctx context.Context, country string) {
	requestsTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("country", country)))
}
```
* <a href="https://opentelemetry.io/docs/languages/go/instrumentation/#using-counters" target="_blank">Using Counters</a>

#### 3.2.3 Histograms

Histograms（直方图）用于记录数值分布情况。

例如，可以通过以下方式上报某段逻辑的处理耗时：

```go
func init() {
    taskExecuteDurationSeconds, err = meter.Float64Histogram(
        "task_execute_duration_seconds",
        metric.WithDescription("Task execute duration in seconds"),
        metric.WithExplicitBucketBoundaries(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0),
    )
}

// metricsHistogramDemo Metrics（指标）- 使用 Histogram 类型指标
func metricsHistogramDemo(ctx context.Context) {
	begin := time.Now()
	doSomething(100)
	cost := time.Since(begin)
	taskExecuteDurationSeconds.Record(ctx, cost.Seconds())
}
```

* <a href="https://opentelemetry.io/docs/languages/go/instrumentation/#using-histograms" target="_blank">Using Histograms</a>

#### 3.2.4 Gauges

Gauges（仪表）用于记录瞬时值。

例如，可以通过以下方式，上报当前内存使用率：

```go
// metricsGaugeDemo Metrics（指标）- 使用 GaugeDemo 类型指标
func metricsGaugeDemo() error {
	memoryUsage, err := meter.Float64ObservableGauge("memory_usage", metric.WithDescription("Memory usage"))
	if err != nil {
		return err
	}

	_, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		o.ObserveFloat64(memoryUsage, 0.1+rng.Float64()*0.2)
		return nil
	}, memoryUsage)
	if err != nil {
		return err
	}

	return nil
}
```

* <a href="https://opentelemetry.io/docs/languages/go/instrumentation/#using-gauges" target="_blank">Using Gauges</a>

### 3.3 Logs

#### 3.3.1 记录日志

```go
import (
	"log/slog"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

var logger = otelslog.NewLogger("helloworld")

// logsDemo Logs（日志）打印日志
func logsDemo(ctx context.Context, req *http.Request) {
	// 上报日志
	logger.InfoContext(ctx, fmt.Sprintf("received request: %s %s", req.Method, req.URL))

	// 添加自定义属性
	attrs := []slog.Attr{
		slog.String("method", req.Method), slog.String("k1", "v1"), slog.Int("k2", 123),
	}
	logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		fmt.Sprintf("report log with attrs, received request: %s %s", req.Method, req.URL),
		attrs...,
	)
}
```

## 4. 快速体验

### 4.1 运行样例

#### 4.1.1 运行

❗❗【非常重要】Go SDK 的场景 `OTLP_ENDPOINT` 无需 `http://` 前缀，SDK 会默认补充，否则上报会失败。

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="{{access_config.token}}" \
-e SERVICE_NAME="{{service_name}}" \
-e OTLP_ENDPOINT="{{access_config.otlp.http_endpoint_without_schema}}" \
-e ENABLE_TRACES="{{access_config.otlp.enable_traces}}" \
-e ENABLE_METRICS="{{access_config.otlp.enable_metrics}}" \
-e ENABLE_LOGS="{{access_config.profiling.enabled}}" helloworld-go:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

#### 4.1.2 运行参数说明

| 参数                   | 值（根据所填写接入信息生成）                                          | 说明                                                                                                                                                       |
|----------------------|:--------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| `TOKEN`              | `"{{access_config.token}}"`                             | 【必须】APM 应用 `Token`。                                                                                                                                      |
| `SERVICE_NAME`       | `"{{service_name}}"`                                    | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分。                                                                                                                          |
| `OTLP_ENDPOINT`      | `"{{access_config.otlp.http_endpoint_without_schema}}"` | 【必须】OT 数据上报地址，支持以下协议：<br />  `gRPC`：`{{access_config.otlp.endpoint}}`<br /> `HTTP`：`{{access_config.otlp.http_endpoint_without_schema}}`（demo 使用该协议演示上报） |
| `PROFILING_ENDPOINT` | `"{{access_config.profiling.endpoint}}"`                | 【可选】Profiling 数据上报地址。                                                                                                                                    |
| `ENABLE_TRACES`      | `{{access_config.otlp.enable_traces}}`                  | 是否启用调用链上报。                                                                                                                                               |
| `ENABLE_METRICS`     | `{{access_config.otlp.enable_metrics}}`                 | 是否启用指标上报。                                                                                                                                                |
| `ENABLE_LOGS`        | `{{access_config.otlp.enable_logs}}`                    | 是否启用日志上报。                                                                                                                                                |
| `ENABLE_PROFILING`   | `{{access_config.profiling.endpoint}}`                  | 是否启用性能分析上报。                                                                                                                                              |

* *<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/" target="_blank">OTLP Exporter Configuration</a>*

### 4.2 查看数据

#### 4.2.1 Traces 检索

{{VIEW_TRACES_DATA}}

#### 4.2.2 指标检索

{{VIEW_CUSTOM_METRICS_DATA}}

#### 4.2.3 日志检索

{{VIEW_LOG_DATA}}

## 5. 了解更多

{{LEARN_MORE}}
