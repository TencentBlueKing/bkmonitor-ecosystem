# Java（Spring Boot OpenTelemetry 无侵入）接入

本指南将帮助您通过 Zero-code Instrumentation 接入蓝鲸应用性能监控，以 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/java-examples/spring-boot-starter" target="_blank">入门项目-spring-boot-starter</a> 为例，介绍调用链、指标、日志数据接入及 SDK 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速入门 OpenTelemetry，接入并体验蓝鲸应用性能监控相关功能。

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
cd bkmonitor-ecosystem/examples/java-examples/spring-boot-starter
docker build -t spring-boot-starter-java:latest .
```

## 2. 快速接入

OpenTelemetry Spring Boot starter 利用 <a href="https://docs.spring.io/spring-boot/reference/using/auto-configuration.html#using.auto-configuration" target="_blank">Spring Boot Auto-configuration</a> 的特性实现 Bean 依赖注入，自动配置与 OpenTelemetry 相关的 Bean 和组件，适用于：

- 避免埋点带来的代码侵入性。

- 不希望或无法使用 <a href="https://opentelemetry.io/docs/zero-code/java/agent/" target="_blank">Java Agent</a> 的场景。

本节将介绍如何自动附加埋点到 SpringBoot 项目上。我们也提供了开发样例 —— <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/java-examples/spring-boot-starter" target="_blank">spring-boot-starter-demo</a> 演示这一过程。

### 2.1 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

### 2.2 环境依赖

OpenTelemetry Spring Boot starter 可与 Spring Boot 2.6+ 和 Spring Boot 3.1+ 配合使用，示例项目使用的是 Spring Boot 3.3.4 的版本，依赖 JDK 的版本至少为 17。

增加以下依赖：

- `opentelemetry-instrumentation-bom`：确保所有 OpenTelemetry 依赖的版本对齐。

- `opentelemetry-spring-boot-starter`：OpenTelemetry starter 依赖，用于对 SpringBoot 项目进行自动埋点。

<a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/java-examples/spring-boot-starter/build.gradle.kts" target="_blank">Gradle（build.gradle.kts）示例：</a>

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
| 环境变量名称                                             | 推荐值                        | 说明                                                                                                                                                                                                                                                                                                                                                                                               |
|----------------------------------------------------|----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `OTEL_SERVICE_NAME`                                | `"${服务名称，请根据右侧说明填写}"`   | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分。优先级比资源属性的设置高，更多信息请参考<a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name" target="_blank">服务名配置</a>。                                                                                                                                                                                                                     |
| `OTEL_EXPORTER_OTLP_HEADERS`                         | `"x-bk-token=todo"` | 【必须】Exporter 导出数据时附加额外的 Headers，多个 key-value 以逗号分隔，例如：`key1=value1,key2=value2`。</br>【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。 |
| `OTEL_EXPORTER_OTLP_PROTOCOL`                      | `"http/protobuf"`          | 【必须】指定<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#protocol-configuration" target="_blank">上报协议</a>，上报协议改变时，上报地址也需要手动修改。</br>【推荐】`protobuf/http`：使用 HTTP 协议上报。</br>【可选】`grpc`：使用 gRPC 上报，如果使用该方式，请确保 `OTEL_EXPORTER_OTLP_ENDPOINT` 也同步改为 gRPC 上报地址。                                                                                                          |
| `OTEL_EXPORTER_OTLP_ENDPOINT`                      | `"http://127.0.0.1:4318"`                         | 【必须】数据<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_endpoint" target="_blank">上报地址</a>，请根据页面指引提供的接入地址进行填写。支持以下协议：<br />`gRPC`：`http://127.0.0.1:4317`<br />`HTTP`：`http://127.0.0.1:4318`（demo 使用该协议演示上报）<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。                                                                                                                                                                                                                   |
| `OTEL_TRACES_EXPORTER`                             | `"otlp"`                   | 【可选】指定用于 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_traces_exporter" target="_blank">Traces Exporter</a>，值为 `"console,otlp"` 时，可以同时在控制台输出。                                                                                                                                                                                                                  |
| `OTEL_METRICS_EXPORTER`                            | `"otlp"`                   | 【可选】指定用于 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_metrics_exporter" target="_blank">Metrics Exporter</a>，值为 `"console,otlp"` 时，可以同时在控制台输出。                                                                                                                                                                                                                |
| `OTEL_LOGS_EXPORTER`                               | `"otlp"`                   | 【可选】指定用于 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_logs_exporter" target="_blank">Logs Exporter</a>，值为 `"console,otlp"` 时，可以同时在控制台输出。 |
| `OTEL_RESOURCE_ATTRIBUTES`                         | `""` | 【可选】<a href="https://opentelemetry.io/docs/concepts/resources/" target="_blank">Resource</a> 代表观测数据所属的资源实体，并通过资源属性进行描述。<a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_resource_attributes" target="_blank">Resource Attributes</a> 设置，多个 key-value 以逗号分隔，例如：`key1=value1,key2=value2`。参考下一小节 `服务信息`。 |

#### 2.4.2 服务信息

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

## 3. 使用场景

如果需要在程序中上报更多数据，可以获取 `Tracer`、`Meter` 对象进行补充埋点，具体请参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/java/otlp/README.md#3-使用场景" target="_blank">Java（OpenTelemetry SDK）接入「3. 使用场景」</a> 部分。

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

- 如果是本地开发测试，请确保您已运行快速验证 demo 数据上报逻辑，快速开始 👉 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/common/ob-all-in-one" target="_blank">ob-all-in-one</a>。

复制以下命令参数在你的终端运行：

```shell
docker run \
-e OTEL_SERVICE_NAME="spring-boot-starter" \
-e OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo" \
-e OTEL_EXPORTER_OTLP_PROTOCOL="http/protobuf" \
-e OTEL_EXPORTER_OTLP_ENDPOINT="http://127.0.0.1:4318" spring-boot-starter-java:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

如果您运行命令是要接入蓝鲸 APM 平台，那么请务必注意以下事项：

- 【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。

- 【必须】`OTEL_EXPORTER_OTLP_ENDPOINT`：数据上报地址，请根据页面指引提供的接入地址进行填写。

- 【必须】`OTEL_EXPORTER_OTLP_PROTOCOL`：如果使用 gRPC 上报，请确保 `OTEL_EXPORTER_OTLP_ENDPOINT` 同步改为 gRPC 上报地址。

### 4.2 查看数据

#### 4.2.1 Traces 检索

Tracing 检索功能主要用于对分布式系统中的请求链路进行跟踪和分析，请参考<a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/explore_traces.md" target="_blank">「应用性能监控 APM/调用链追踪」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/traces.png)

#### 4.2.2 指标检索

自定义指标功能旨在帮助用户针对特定应用及其服务进行深度性能指标监控，请参考<a href="#" target="_blank">「应用性能监控 APM/自定义指标」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/custom-metrics.png)

#### 4.2.3 日志检索

日志功能主要用于查看和分析对应服务（应用程序）运行过程中产生的各类日志信息，请参考<a href="#" target="_blank">「应用性能监控 APM/日志分析」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/logs.png)

## 5. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>