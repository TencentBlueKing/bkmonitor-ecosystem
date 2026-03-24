# Java（OpenTelemetry SDK）接入

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
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/java-examples/helloworld
docker build -t helloworld-java .
```

## 2. 快速接入

### 2.1 创建应用

{{APPLICATION_SET_UP}}

### 2.2 开箱即用 SDK 接入示例

OpenTelemetry 提供标准化的框架和工具包，用于创建和管理 Traces、Metrics、Logs 数据。

示例项目提供集成 OpenTelemetry Java SDK 并将观测数据发送到 bk-collector 的方式，可以参考 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/java-examples/helloworld/src/main/java/com/tencent/bkm/demo/helloworld/service/impl/otlp/OtlpService.java" target="_blank">service/impl/otlp/OtlpService.java</a> 进行接入。

### 2.3 关键配置

🌟 请仔细阅读本章节，以确保观测数据能准确上报到 APM。

#### 2.3.1 上报地址 & 应用 Token

{{MUST_CONFIG_EXPORTER}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/java-examples/helloworld/src/main/java/com/tencent/bkm/demo/helloworld/service/impl/otlp/OtlpService.java" target="_blank">service/impl/otlp/OtlpService.java getTracerProvider</a>  提供了创建样例：

```java
private SdkTracerProvider getTracerProvider(Resource resource) {
    SpanExporter exporter = this.getSpanExporter();
    return SdkTracerProvider.builder()
            .setResource(resource)
            .addSpanProcessor(
                    BatchSpanProcessor.builder(exporter)
                            .setScheduleDelay(EXPORTER_DEFAULT_SCHEDULE_DELAY)
                            .build())
            .build();
}

private SpanExporter getSpanExporter() {
    return OtlpHttpSpanExporter.builder()
            //❗️❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
            .setEndpoint(config.getEndpoint() + "/v1/traces")
            .setTimeout(EXPORTER_DEFAULT_TIMEOUT)
            // ❗️❗【非常重要】配置为应用 Token
            .addHeader("x-bk-token", this.config.getToken())
            .build();
}
```

#### 2.3.2 服务信息

{{MUST_CONFIG_RESOURCES}}

示例项目在 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/java-examples/helloworld/src/main/java/com/tencent/bkm/demo/helloworld/service/impl/otlp/OtlpService.java" target="_blank">service/impl/otlp/OtlpService.java getResource</a> 提供了创建样例：

```java
private Resource getResource() {
    Resource extraResource = Resource.builder()
            //❗❗【非常重要】应用服务唯一标识
            .put(AttributeKey.stringKey("service.name"), this.config.getServiceName())
            .build();
    // getDefault 提供了部分 SDK 默认属性
    return Resource.getDefault()
            .merge(extraResource)
            // ...其他 Resource
}
```

## 3. 使用场景

示例项目整理常见的使用场景，集中在：

```java
public String handleHelloWorld(HttpExchange exchange) throws Exception {
    Span span = this.tracer.spanBuilder("Handle/HelloWorld").startSpan();
    try (Scope ignored = span.makeCurrent()) {
        // Logs（日志）
        this.logsDemo(exchange);

        String country = choiceCountry();
        logger.info("get country -> {}", country);

        // Metrics（指标） - Counter 类型
        this.metricsCounterDemo(country);
        // Metrics（指标） - Histograms 类型
        this.metricsHistogramDemo();

        // Traces（调用链）- 自定义 Span
        this.tracesCustomSpanDemo();
        // Traces（调用链）- Span 事件
        this.tracesSpanEventDemo();
        // Traces（调用链）- 模拟错误
        tracesRandomErrorDemo();

        return generateGreeting(country);
    } catch (Exception e) {
        span.recordException(e);
        throw e;
    } finally {
        span.end();
    }
}
```

可以结合代码和下方说明进行使用：<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/java-examples/helloworld/src/main/java/com/tencent/bkm/demo/helloworld/service/impl/http/HelloWorldHttpHandler.java" target="_blank">service/impl/http/HelloWorldHttpHandler.java</a>。

### 3.1 Traces

#### 3.1.1 创建 Resource

Resource 代表观测数据所属的资源实体。

例如运行在 Kubernetes 上的容器所生成的观测数据，具有进程名称、Pod 名称等资源实体信息。

可以通过 <a href="https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/main/instrumentation/resources/library" target="_blank">opentelemetry-resources</a> 自动检测正在运行的进程、所在操作系统等资源属性信息。

增加依赖：

```groovy
implementation("io.opentelemetry.instrumentation:opentelemetry-resources:2.7.0-alpha")
```

初始化 `Resources`

```java
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.instrumentation.resources.ContainerResource;
import io.opentelemetry.instrumentation.resources.HostResource;
import io.opentelemetry.instrumentation.resources.OsResource;
import io.opentelemetry.instrumentation.resources.ProcessResource;
import io.opentelemetry.sdk.resources.Resource;

private Resource getResource() {
    Resource extraResource = Resource.builder()
            //❗❗【非常重要】应用服务唯一标识
            .put(AttributeKey.stringKey("service.name"), this.config.getServiceName())
            .build();
    // getDefault 提供了部分 SDK 默认属性
    return Resource.getDefault()
            .merge(extraResource)
            .merge(ProcessResource.get())
            .merge(ContainerResource.get())
            .merge(OsResource.get())
            .merge(HostResource.get());
}
```

* <a href="https://opentelemetry.io/docs/languages/java/sdk/#resource" target="_blank">Resources</a>

#### 3.1.2 创建 Span

Span 是 Traces 的构建块，代表一个工作或操作单元。

Span 通过 `GlobalOpenTelemetry.getTracer("helloworld").spanBuilder` 进行创建：

```java
private void tracesCustomSpanDemo() {
    Span span = GlobalOpenTelemetry.getTracer("helloworld").spanBuilder("CustomSpanDemo/doSomething").startSpan();
    try (Scope ignored = span.makeCurrent()) {
        doSomething(50);
    } finally {
        span.end();
    }
}
```

* <a href="https://opentelemetry.io/docs/languages/java/instrumentation/#create-spans" target="_blank">Creating Spans</a>

#### 3.1.3 设置属性

Attributes（属性）是 Span 元数据，以 Key-Value 形式存在。

在 Span 设置属性，对问题定位、过滤、聚合非常有帮助。

```java
// 增加 Span 自定义属性
span.setAttribute(AttributeKey.longKey("helloworld.kind"), 1L);
span.setAttribute(AttributeKey.stringKey("helloworld.step"), "tracesCustomSpanDemo");
```

* <a href="https://opentelemetry.io/docs/languages/java/instrumentation/#span-attributes" target="_blank">Span Attributes</a>

#### 3.1.4 设置事件

Event（事件）是一段人类可读信息，用于记录 Span 生命周期内发生的事情。

```java
private void tracesSpanEventDemo() {
    Span span = tracer.spanBuilder("tracesSpanEventDemo/doSomething").startSpan();
    try (Scope ignored = span.makeCurrent()) {
        Attributes evnetAttributes = Attributes.of(
                AttributeKey.longKey("helloworld.kind"), 2L,
                AttributeKey.stringKey("helloworld.step"), "tracesSpanEventDemo"
        );
        span.addEvent("Before doSomething", evnetAttributes);
        doSomething(50);
        span.addEvent("After doSomething", evnetAttributes);
    } finally {
        span.end();
    }
}
```

* <a href="https://opentelemetry.io/docs/languages/java/instrumentation/#create-spans-with-events" target="_blank">Span Events</a>

#### 3.1.5 记录异常

当一个 Span 出现异常，可以对其进行异常记录。

```java
private void tracesRandomErrorDemo() throws Exception {
    try {
        throwRandomError(0.1F);
    } catch (Exception e) {
        // 获取当前 Span
        Span span = Span.current();
        // 记录异常事件
        // Refer: https://opentelemetry.io/docs/languages/java/instrumentation/#record-exceptions-in-spans
        span.recordException(e);
        throw e;
    }
}
```

* <a href="https://opentelemetry.io/docs/languages/java/instrumentation/#record-exceptions-in-spans" target="_blank">Record exceptions in spans</a>

#### 3.1.6 设置状态

当一个 Span 未能成功，可以通过设置状态进行显式标记。

```java
span.setStatus(StatusCode.ERROR, e.getMessage());
```

* <a href="https://opentelemetry.io/docs/languages/java/instrumentation/#set-span-status" target="_blank">Set span status</a>

### 3.2 Metrics

#### 3.2.1 Counters

Counters（计数器）用于记录非负递增值。

例如，可以通过以下方式上报请求总数：

```java
//【建议】初始化指标再使用，而不是在业务逻辑里初始化
this.requestsTotal = this.meter.counterBuilder("requests_total")
        .setDescription("Total number of HTTP requests")
        .setUnit("requests")
        .build();

private void metricsCounterDemo(String country) {
    this.requestsTotal.add(1, Attributes.of(AttributeKey.stringKey("country"), country));
}
```

* <a href="https://opentelemetry.io/docs/languages/java/instrumentation/#using-counters" target="_blank">Using Counters</a>

#### 3.2.2 Histograms

Histograms（直方图）用于记录数值分布情况。

例如，可以通过以下方式上报某段逻辑的处理耗时：

```java
this.taskExecuteDurationSeconds = this.meter.histogramBuilder("task_execute_duration_seconds")
        .setDescription("Task execute duration in seconds")
        .setExplicitBucketBoundariesAdvice(List.of(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0))
        .setUnit("seconds")
        .build();

private void metricsHistogramDemo() {
    long begin = System.nanoTime();
    doSomething(100);
    long end = System.nanoTime();
    double durationInSeconds = (end - begin) / 1_000_000_000.0;
    // 记录耗时
    taskExecuteDurationSeconds.record(durationInSeconds);
}
```

* <a href="https://opentelemetry.io/docs/languages/java/instrumentation/#using-histograms" target="_blank">Using Histograms</a>

#### 3.2.3 Gauges

Gauges（仪表）用于记录瞬时值。

例如，可以通过以下方式，上报当前内存使用率：

```java
private void metricsGaugeDemo() {
    this.meter.gaugeBuilder("memory_usage")
            .setDescription("Memory usage")
            .buildWithCallback(
                    result -> {
                        Random random = new Random();
                        result.record(0.1 + random.nextDouble() * 0.2);
                    }
            );
}
```

* <a href="https://opentelemetry.io/docs/languages/java/instrumentation/#using-observable-async-gauges" target="_blank">Using Observable (Async) Gauges</a>

### 3.3 Logs

Logs 与 Traces / Metrics 不同，没有提供 OpenTelemetry API，设计原理可参考：<a href="https://opentelemetry.io/docs/specs/otel/logs/" target="_blank">OpenTelemetry Logging</a>。

Logs 采用和现有的日志框架（例如 SLF4j、JUL、Logback、Log4j）进行结合，通过 <a href="https://opentelemetry.io/docs/languages/java/instrumentation/#log-appenders" target="_blank">Log appenders</a> 桥接到 OpenTelemetry 生态。

#### 3.3.1 Log4j 接入
> 示例项目已提供 `Log4j` 可运行的接入案例，可以跳过本小节，通过运行实例代码的方式体验接入效果。

增加依赖

```groovy
implementation("io.opentelemetry.instrumentation:opentelemetry-log4j-appender-2.17:2.8.0-alpha")
```

增加 `src/main/resources/log4j2.xml`

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Configuration status="INFO" packages="io.opentelemetry.instrumentation.log4j.appender.v2_17">
    <Appenders>
        <Console name="Console" target="SYSTEM_OUT">
            <PatternLayout
                pattern="%d{HH:mm:ss.SSS} [%t] %-5level %logger{36} trace_id: %X{trace_id} span_id: %X{span_id} trace_flags: %X{trace_flags} - %msg%n"/>
        </Console>
        <OpenTelemetry name="OpenTelemetryAppender" captureMapMessageAttributes="true" captureExperimentalAttributes="true"/>
    </Appenders>
    <Loggers>
        <Root level="info">
            <AppenderRef ref="OpenTelemetryAppender"/>
            <AppenderRef ref="Console"/>
        </Root>
    </Loggers>
</Configuration>
```

在 OpenTelemetry 接入实例中，增加 `LogsAppender` 配置：

```java
private void setUpLogsAppender() {
    if (this.config.isEnableLogs()) {
        io.opentelemetry.instrumentation.log4j.appender.v2_17.OpenTelemetryAppender.install(this.openTelemetrySdk);
    }
}
```

#### 3.3.2 记录日志

```java
import org.apache.logging.log4j.Logger;

private static final Logger logger = LogManager.getLogger(HelloWorldHttpHandler.class);

private void logsDemo(HttpExchange exchange) {
    logger.info("received request: {} {}", exchange.getRequestMethod(), exchange.getRequestURI());
}
```

## 4. 快速体验

### 4.1 运行样例

#### 4.1.1 运行

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="xxxx" \
-e SERVICE_NAME="helloworld" \
-e OTLP_ENDPOINT="{{access_config.otlp.http_endpoint}}" \
-e PROFILING_ENDPOINT="{{access_config.profiling.endpoint}}" \
-e ENABLE_PROFILING="true" \
-e ENABLE_TRACES="true" \
-e ENABLE_METRICS="true" \
-e ENABLE_LOGS="true" helloworld-java
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

#### 4.1.2 运行参数说明

{{DEMO_RUN_PARAMETERS}}

### 4.2 查看数据

#### 4.2.1 Traces 检索

{{VIEW_TRACES_DATA}}

#### 4.2.2 指标检索

{{VIEW_CUSTOM_METRICS_DATA}}

#### 4.2.3 日志检索

{{VIEW_LOG_DATA}}

## 5. 了解更多

{{LEARN_MORE}}