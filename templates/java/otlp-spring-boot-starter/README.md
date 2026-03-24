# Java（Spring Boot OpenTelemetry 无侵入）接入

本指南将帮助您通过 Zero-code Instrumentation 接入蓝鲸应用性能监控，以 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/java-examples/spring-boot-starter" target="_blank">入门项目-spring-boot-starter</a> 为例，介绍调用链、指标、日志数据接入及 SDK 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速入门 OpenTelemetry，接入并体验蓝鲸应用性能监控相关功能。

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
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/java-examples/spring-boot-starter
docker build -t spring-boot-starter-java:latest .
```

## 2. 快速接入

OpenTelemetry Spring Boot starter 利用 <a href="https://docs.spring.io/spring-boot/reference/using/auto-configuration.html#using.auto-configuration" target="_blank">Spring Boot Auto-configuration</a> 的特性实现 Bean 依赖注入，自动配置与 OpenTelemetry 相关的 Bean 和组件，适用于：

- 避免埋点带来的代码侵入性。

- 不希望或无法使用 <a href="https://opentelemetry.io/docs/zero-code/java/agent/" target="_blank">Java Agent</a> 的场景。

本节将介绍如何自动附加埋点到 SpringBoot 项目上。我们也提供了开发样例 —— <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/java-examples/spring-boot-starter" target="_blank">spring-boot-starter-demo</a> 演示这一过程。

### 2.1 创建应用

{{APPLICATION_SET_UP}}

### 2.2 环境依赖

OpenTelemetry Spring Boot starter 可与 Spring Boot 2.6+ 和 Spring Boot 3.1+ 配合使用，示例项目使用的是 Spring Boot 3.3.4 的版本，依赖 JDK 的版本至少为 17。

增加以下依赖：

- `opentelemetry-instrumentation-bom`：确保所有 OpenTelemetry 依赖的版本对齐。

- `opentelemetry-spring-boot-starter`：OpenTelemetry starter 依赖，用于对 SpringBoot 项目进行自动埋点。

<a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/java-examples/spring-boot-starter/build.gradle.kts" target="_blank">Gradle（build.gradle.kts）示例：</a>

```kotlin
import org.springframework.boot.gradle.plugin.SpringBootPlugin

plugins {
  id("java")
  id("org.springframework.boot") version "3.3.4"
}

dependencies {
  implementation(platform(SpringBootPlugin.BOM_COORDINATES))
  implementation(platform("io.opentelemetry.instrumentation:opentelemetry-instrumentation-bom:2.9.0"))
  implementation("org.springframework.boot:spring-boot-starter-web")
  implementation("io.opentelemetry.instrumentation:opentelemetry-spring-boot-starter")
}

```

### 2.3 增加项目配置

<a href="https://opentelemetry.io/docs/zero-code/java/spring-boot-starter/sdk-configuration/" target="_blank">Spring Boot starter 配置</a> 有两种方式：环境变量（推荐）、配置文件（`application.properties` 或 `application.yaml`），优先级：环境变量 > 配置文件。

以下仅列举出最小化配置项，具体的配置值请参考下文「2.4 关键配置」。

环境变量配置示例：

```shell
export OTEL_SERVICE_NAME="your-service-name"
export OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo"
export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"
```

配置 `application.properties` 示例：

```ini
otel.service.name=your-service-name
otel.exporter.otlp.headers=x-bk-token=todo
otel.exporter.otlp.endpoint=http://localhost:4318
```

### 2.4 关键配置

#### 2.4.1 环境变量配置
{{AUTOMATIC_RUN_PARAMETERS}}

#### 2.4.2 服务信息

{{MUST_CONFIG_RESOURCES}}

## 3. 使用场景

如果需要在程序中上报更多数据，可以获取 `Tracer`、`Meter` 对象进行补充埋点，具体请参考 <a href="{{REFER_JAVA_OTLP_URL}}#3-使用场景" target="_blank">Java（OpenTelemetry SDK）接入「3. 使用场景」</a> 部分。

获取 `Tracer`、`Meter` 对象：

```java
@Service
public class TravelService {
    private final Tracer tracer;
    private final Meter meter;

    public TravelService(OpenTelemetry openTelemetry) {
        this.tracer = openTelemetry.getTracer(getClass().getName());
        this.meter = openTelemetry.getMeter(getClass().getName());
    };
}
```

## 4. 快速体验

### 4.1 运行样例

运行前注意事项：

- 运行之前请记得执行 `docker build` 命令，参考本文 1.3 节。

- 如果是本地开发测试，请确保您已运行快速验证 demo 数据上报逻辑，快速开始 👉 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/common/ob-all-in-one" target="_blank">ob-all-in-one</a>。

复制以下命令参数在你的终端运行：

```shell
docker run \
-e OTEL_SERVICE_NAME="spring-boot-starter" \
-e OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo" \
-e OTEL_EXPORTER_OTLP_PROTOCOL="http/protobuf" \
-e OTEL_EXPORTER_OTLP_ENDPOINT="{{access_config.otlp.http_endpoint}}" spring-boot-starter-java:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

如果您运行命令是要接入蓝鲸 APM 平台，那么请务必注意以下事项：

- 【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。

- 【必须】`OTEL_EXPORTER_OTLP_ENDPOINT`：数据上报地址，请根据页面指引提供的接入地址进行填写。

- 【必须】`OTEL_EXPORTER_OTLP_PROTOCOL`：如果使用 gRPC 上报，请确保 `OTEL_EXPORTER_OTLP_ENDPOINT` 同步改为 gRPC 上报地址。

### 4.2 查看数据

#### 4.2.1 Traces 检索

{{VIEW_TRACES_DATA}}

#### 4.2.2 指标检索

{{VIEW_CUSTOM_METRICS_DATA}}

#### 4.2.3 日志检索

{{VIEW_LOG_DATA}}

## 5. 了解更多

{{LEARN_MORE}}