# Java（SkyWalking Java Agent）接入

本指南将帮助您使用 SkyWalking Agent 接入蓝鲸应用性能监控，以 <a href=https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/examples/helloworld.md target="_blank">入门项目-HelloWorld</a> 为例，介绍调用链、指标、日志数据接入及 Agent 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速入门 SkyWalking，接入并体验蓝鲸应用性能监控相关功能。

## 1. 前置准备

### 1.1 术语介绍

* Traces：<a href="https://opentelemetry.io/docs/concepts/signals/traces/" target="_blank">调用链</a>，表示请求在应用程序的执行路径。
* Metrics：<a href="https://opentelemetry.io/docs/concepts/signals/metrics/" target="_blank">指标</a>，表示对运行服务的测量。
* Logs: <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">日志</a>，表示对事件的记录。
* Profiling: <a href="https://grafana.com/docs/pyroscope/latest/introduction/profiling/" target="_blank">性能分析</a>，表示对应用程序运行时资源的持续测量。
* Telemetry Data：观测数据，指代 Traces、Metrics、Logs、Profiling 等。
* APM：蓝鲸观测平台应用性能监控，提供四类观测数据开箱即用的观测能力。
* <a href="https://github.com/TencentBlueKing/bkmonitor-datalink/tree/main/pkg/collector" target="_blank">bk-collector</a>：腾讯蓝鲸的 APM 服务端组件，负责接收 Prometheus、OpenTelemetry、Jaeger、Skywalking 等主流开源组件的观测数据，并对数据进行清洗转发到观测平台链路。

如果想了解 SkyWalking 对于上述相关术语的解释，请查看 <a href="https://skywalking.apache.org/docs/main/v10.1.0/en/concepts-and-designs/overview/" target="_blank">Overview</a>。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：
* Git
* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/java-examples/skywalking-agent
docker build -t helloworld-java-sw:latest .
```

## 2. 快速接入

SkyWalking Java Agent 是基于 <a href="https://docs.oracle.com/en/java/javase/17/docs/api/java.instrument/java/lang/instrument/package-summary.html" target="_blank">java.lang.instrument API</a> 和字节码增强技术制作而成的，通过添加 `-javaagent:/path/to/skywalking-agent.jar` 到 JVM 参数中，自动实现对应用程序的监控和追踪。适用于：避免埋点带来代码侵入性。

本节将介绍如何自动附加埋点到 SpringBoot 项目上。我们也提供了开发样例 —— <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/java-examples/skywalking-agent" target="_blank">skywalking-agent-demo</a> 演示这一过程。

### 2.1 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

### 2.2 环境依赖

SkyWalking 提供了不同版本的 SkyWalking Java Agent 和 plugins，需要根据应用程序依赖选择对应的版本。选择一个合适的 SkyWalking Java Agent：

<a href="https://skywalking.apache.org/docs/skywalking-java/v9.3.0/en/setup/service-agent/java-agent/readme/" target="_blank">Setup java agent 页面</a> 可以查看对应版本的兼容信息，选定版本后到 <a href="https://skywalking.apache.org/downloads/" target="_blank">SkyWalking 下载页面</a> 进行下载。

```shell
# 下载 SkyWalking Java Agent 以及对应的 asc 文件, sha512 文件和 KEYS 文件
curl -O -L https://dlcdn.apache.org/skywalking/java-agent/9.3.0/apache-skywalking-java-agent-9.3.0.tgz
curl -O -L https://downloads.apache.org/skywalking/java-agent/9.3.0/apache-skywalking-java-agent-9.3.0.tgz.asc
curl -O -L https://downloads.apache.org/skywalking/java-agent/9.3.0/apache-skywalking-java-agent-9.3.0.tgz.sha512
curl -O -L https://downloads.apache.org/skywalking/KEYS

# 校验文件并解压
gpg --import KEYS
gpg --verify apache-skywalking-java-agent-9.3.0.tgz.asc apache-skywalking-java-agent-9.3.0.tgz
sha512sum -c apache-skywalking-java-agent-9.3.0.tgz.sha512
tar -xzvf apache-skywalking-java-agent-9.3.0.tgz
```

### 2.3 关键配置 & 通过 Agent 运行项目

SkyWalking 有 <a href="https://skywalking.apache.org/docs/skywalking-java/v9.3.0/en/setup/service-agent/java-agent/setting-override/" target="_blank">四种配置方式</a>，优先级： `Agent Options > System.Properties(-D) > System environment variables > Config file`。下面给出上报到 APM 的关键配置项，<a href="https://skywalking.apache.org/docs/skywalking-java/v9.3.0/en/setup/service-agent/java-agent/configurations/" target="_blank">Table of Agent Configuration Properties</a> 包含 SkyWalking 可配置的属性列表，以下仅列出关键配置：

#### 2.3.1 Agent Options

在 Java 命令中的 Agent 路径后添加属性：

```shell
# 【必须】服务名，一个应用可以有多个服务，通过该属性区分。有关服务名配置的更多信息，请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name
java -javaagent:/path/to/skywalking-agent.jar=agent.service_name=helloworld-java-sw,\
# ❗❗【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。
agent.authentication=x-bk-token,\
# # ❗❗【非常重要】数据上报地址，请根据页面指引提供的 gRPC 接入地址进行填写。
collector.backend_service="127.0.0.1:4317" \
-jar your-app.jar
```

#### 2.3.2 System.Properties(-D)

在 Java 命令中添加这个系统属性：

```shell
java -javaagent:/path/to/skywalking-agent.jar \
# 【必须】服务名，一个应用可以有多个服务，通过该属性区分。有关服务名配置的更多信息，请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name
-Dskywalking.agent.service_name="helloworld-java-sw" \
# ❗❗【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。
-Dskywalking.agent.authentication="x-bk-token" \
# # ❗❗【非常重要】数据上报地址，请根据页面指引提供的 gRPC 接入地址进行填写。
-Dskywalking.collector.backend_service="127.0.0.1:4317" \
-jar your-app.jar
```
#### 2.3.3 System environment variables
环境变量配置：

```shell
# 【必须】服务名，一个应用可以有多个服务，通过该属性区分。有关服务名配置的更多信息，请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name
export SW_AGENT_NAME="helloworld-java-sw"
# ❗❗【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。
export SW_AGENT_AUTHENTICATION="x-bk-token"
# ❗❗【非常重要】数据上报地址，请根据页面指引提供的 gRPC 接入地址进行填写。
export SW_AGENT_COLLECTOR_BACKEND_SERVICES="127.0.0.1:4317"
java -javaagent:/path/to/skywalking-agent.jar -jar your-app.jar
```

## 3. 使用场景

SkyWalking Java Agent 已自动配置上报数据所需的服务，并对常用包的类进行字节码增强，即在无需修改代码的情况下，能上报常用包的观测数据。当你想增强监控能力时：

### 3.1 增加插件

<a href="https://skywalking.apache.org/docs/skywalking-java/v9.3.0/en/setup/service-agent/java-agent/optional-plugins/#optional-plugins" target="_blank">Optional Plugins</a> 里还存在一些没有默认引入的插件，可以复制到 `plugins` 文件夹以增强能力。

```shell
cp /path/to/skywalking-agent/optional-plugins/apm-resttemplate-6.x-plugin-9.3.0.jar /path/to/skywalking-agent/plugins/
```

### 3.2 引入 log4j2
log4j2 配置需要增加 `log4j2.xml` 文件和在 `build.gradle.kts` 文件引入依赖。

#### 3.2.1 增加 log4j2.xml

```xml
<Configuration status="WARN">
    <Appenders>
        <Console name="Console" target="SYSTEM_OUT">
            <PatternLayout pattern="%d %highlight{%p} %class{1.} %style{[%t]}{blue} [%traceId] %location %m %ex%n"/>
        </Console>
        <GRPCLogClientAppender name="grpc-log">
            <PatternLayout pattern="%d{HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
        </GRPCLogClientAppender>
    </Appenders>
    <Loggers>
        <Root level="info" includeLocation="false">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="grpc-log"/>
        </Root>
    </Loggers>
</Configuration>
```

#### 3.2.2 引入依赖

`build.gradle.kts` 文件示例：
```kotlin
import org.springframework.boot.gradle.plugin.SpringBootPlugin
plugins {
    id("java")
    id("org.springframework.boot") version "3.3.4"
    id("org.graalvm.buildtools.native") version "0.10.3"
}

dependencies {
    implementation("org.springframework.boot:spring-boot-starter-web"){
        exclude("org.springframework.boot", "spring-boot-starter-logging")
    }
    implementation("org.springframework.boot:spring-boot-starter-log4j2")
    implementation("org.apache.skywalking:apm-toolkit-log4j-2.x:9.3.0")
}
```

## 4. 快速体验

### 4.1 运行样例

#### 4.1.1 运行

复制以下命令参数在你的终端运行：

```shell
docker run \
-e SW_AGENT_NAME="helloworld-java-sw" \
-e SW_AGENT_COLLECTOR_BACKEND_SERVICES="127.0.0.1:4317" \
-e SW_AGENT_AUTHENTICATION="x-bk-token" helloworld-java-sw:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

#### 4.1.2 运行参数说明

运行参数说明：

| 参数                       | 推荐值                               | 说明                                        |
|---------------------------|--------------------------------------|-------------------------------------------|
| `SW_AGENT_NAME`           | `"helloworld-java-sw"`               | 【必须】服务唯一标识，用于表示提供相同功能/逻辑的逻辑组。                |
| `SW_AGENT_COLLECTOR_BACKEND_SERVICES`      | `"127.0.0.1:4317"`  | 【必须】数据上报地址，请根据页面指引提供的接入地址进行填写。（demo 使用该协议演示上报）<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。        |
| `SW_AGENT_AUTHENTICATION` | `"x-bk-token"`                                 | 【必须】上报数据时需要的认证信息。</br>【非常重要】x-bk-token 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。。                            |


### 4.2 查看数据

#### 4.2.1 Traces 检索

Tracing 检索功能主要用于对分布式系统中的请求链路进行跟踪和分析，请参考<a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/explore_traces.md" target="_blank">「应用性能监控 APM/调用链追踪」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/traces.png)

#### 4.2.2 指标检索

自定义指标功能旨在帮助用户针对特定应用及其服务进行深度性能指标监控，请参考<a href="#" target="_blank">「应用性能监控 APM/自定义指标」</a> 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/open/common/images/custom-metrics.png)

## 5. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>