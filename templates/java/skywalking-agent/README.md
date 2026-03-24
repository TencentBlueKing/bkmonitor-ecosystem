# Java（SkyWalking Java Agent）接入

本指南将帮助您使用 SkyWalking Agent 接入蓝鲸应用性能监控，以 <a href={{REFER_HELLO_WORLD_URL}} target="_blank">入门项目-HelloWorld</a> 为例，介绍调用链、指标、日志数据接入及 Agent 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速入门 SkyWalking，接入并体验蓝鲸应用性能监控相关功能。

## 1. 前置准备

### 1.1 术语介绍

{{TERM_INTRO}}

如果想了解 SkyWalking 对于上述相关术语的解释，请查看 <a href="https://skywalking.apache.org/docs/main/v10.1.0/en/concepts-and-designs/overview/" target="_blank">Overview</a>。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：
* Git
* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/java-examples/skywalking-agent
docker build -t helloworld-java-sw:latest .
```

## 2. 快速接入

SkyWalking Java Agent 是基于 <a href="https://docs.oracle.com/en/java/javase/17/docs/api/java.instrument/java/lang/instrument/package-summary.html" target="_blank">java.lang.instrument API</a> 和字节码增强技术制作而成的，通过添加 `-javaagent:/path/to/skywalking-agent.jar` 到 JVM 参数中，自动实现对应用程序的监控和追踪。适用于：避免埋点带来代码侵入性。

本节将介绍如何自动附加埋点到 SpringBoot 项目上。我们也提供了开发样例 —— <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/java-examples/skywalking-agent" target="_blank">skywalking-agent-demo</a> 演示这一过程。

### 2.1 创建应用

{{APPLICATION_SET_UP}}

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
collector.backend_service="{{access_config.sw.endpoint}}" \
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
-Dskywalking.collector.backend_service="{{access_config.sw.endpoint}}" \
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
export SW_AGENT_COLLECTOR_BACKEND_SERVICES="{{access_config.sw.endpoint}}"
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
{% raw %}
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
{% endraw %}
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
-e SW_AGENT_COLLECTOR_BACKEND_SERVICES="{{access_config.sw.endpoint}}" \
-e SW_AGENT_AUTHENTICATION="x-bk-token" helloworld-java-sw:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

#### 4.1.2 运行参数说明

{{SW_DEMO_RUN_PARAMETERS}}

### 4.2 查看数据

#### 4.2.1 Traces 检索

{{VIEW_TRACES_DATA}}

#### 4.2.2 指标检索

{{VIEW_CUSTOM_METRICS_DATA}}

## 5. 了解更多

{{LEARN_MORE}}