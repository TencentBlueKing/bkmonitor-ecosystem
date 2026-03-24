# Profiling-Java（Pyroscope SDK）接入

{{PROFILING_OVERVIEW}}

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

## 2. 快速体验

### 2.1 运行样例

{{PROFILING_RUN_PARAMETERS}}

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="{{access_config.token}}" \
-e SERVICE_NAME="{{service_name}}" \
-e PROFILING_ENDPOINT="{{access_config.profiling.endpoint}}" \
-e ENABLE_PROFILING="{{access_config.profiling.enabled}}" helloworld-java
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

### 2.2 查看数据

等待片刻，便可在「服务详情-Profiling」看到应用数据。

![](image/img.png)

## 3. 快速接入

### 3.1 Pyroscope SDK

{{MUST_CONFIG_PROFILING}}

在项目中引入模块依赖：

```groovy
implementation("io.pyroscope:agent:0.14.0")
```

示例项目提供集成 Pyroscope Java SDK 并将性能数据发送到 bk-collector 的方式，可以参考 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/java-examples/helloworld/src/main/java/com/tencent/bkm/demo/helloworld/service/impl/profiling/ProfilingService.java" target="_blank">service/impl/profiling/ProfilingService.java</a> 进行接入：

```java
this.pyroscopeConfig = new io.pyroscope.javaagent.config.Config.Builder()
        //❗❗【非常重要】请传入应用 Token
        .setAuthToken(config.getToken())
        //❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
        .setServerAddress(config.getProfilingEndpoint())
        //❗❗【非常重要】应用服务唯一标识
        .setApplicationName(this.config.getServiceName())
        .setProfilingEvent(EventType.ITIMER)
        .setFormat(Format.JFR)
        .build();
```

参考官方文档以获得更多信息：<a href="https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/java/" target="_blank">Configure the client to send profiles - Java</a>

### 3.2 关联 Traces 数据

Pyroscope 支持同 OpenTelemetry 集成，将 Traces 和 Profiling 数据链接起来，从而实现分析具体跨度（Span）的资源使用情况的目的。

在开始之前，可以阅读文档 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/java-examples/helloworld/README.md" target="_blank">Java（OpenTelemetry SDK）接入</a> 了解 OpenTelemetry。

在项目中引入依赖：

```groovy
implementation("io.pyroscope:otel:0.10.1.4")
```

配置 `TraceProvider`

```java
import io.opentelemetry.sdk.trace.SdkTracerProviderBuilder;
import io.otel.pyroscope.PyroscopeOtelConfiguration;
import io.opentelemetry.sdk.OpenTelemetrySdk;

SdkTracerProviderBuilder providerBuilder = SdkTracerProvider.builder();
PyroscopeOtelConfiguration pyroscopeTelemetryConfig = new PyroscopeOtelConfiguration.Builder()
        .setAddSpanName(true)
        .setAppName("my-opentelemetry-proj-java")
        .setPyroscopeEndpoint("{HTTP_PUSH_URL}/pyroscope")
        .setRootSpanOnly(false)
        .build();
providerBuilder.addSpanProcessor(new PyroscopeOtelSpanProcessor(pyroscopeTelemetryConfig));
OpenTelemetrySdk.builder().setTracerProvider(providerBuilder.build()).buildAndRegisterGlobal();
```

参考官方文档以获得更多信息：<a href="https://grafana.com/docs/pyroscope/latest/configure-client/trace-span-profiles/java-span-profiles/" target="_blank">Span profiles with Traces to profiles for Java</a>

## 4. 了解更多

{{LEARN_MORE}}
