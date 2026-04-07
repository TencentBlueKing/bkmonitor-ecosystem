# Go（外部版 tRPC 云观 Oteam SDK）接入

本指南将帮助您使用云观 Oteam（OpenTelemetry Oteam）SDK 将**外部版** tRPC Go 项目接入蓝鲸应用性能监控。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速接入并体验蓝鲸应用性能监控相关功能。

**本篇文档面向开源版 tRPC-Go，如果您使用的是 <a href="https://github.com/trpc-group/trpc-go" target="_blank">trpc-go/trpc-go</a>（内部版 tRPC-Go【大部分场景下都是内部版】），请参考 <a href="#" target="_blank">Go（tRPC 云观 Oteam SDK）接入</a> 。**

## 1. 前置准备

### 1.1 术语介绍

* Traces：<a href="https://opentelemetry.io/docs/concepts/signals/traces/" target="_blank">调用链</a>，表示请求在应用程序的执行路径。
* Metrics：<a href="https://opentelemetry.io/docs/concepts/signals/metrics/" target="_blank">指标</a>，表示对运行服务的测量。
* Logs: <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">日志</a>，表示对事件的记录。
* Profiling: <a href="https://grafana.com/docs/pyroscope/latest/introduction/profiling/" target="_blank">性能分析</a>，表示对应用程序运行时资源的持续测量。
* Telemetry Data：观测数据，指代 Traces、Metrics、Logs、Profiling 等。
* APM：蓝鲸观测平台应用性能监控，提供四类观测数据开箱即用的观测能力。
* <a href="https://github.com/TencentBlueKing/bkmonitor-datalink/tree/main/pkg/collector" target="_blank">bk-collector</a>：腾讯蓝鲸的 APM 服务端组件，负责接收 Prometheus、OpenTelemetry、Jaeger、Skywalking 等主流开源组件的观测数据，并对数据进行清洗转发到观测平台链路。
* <a href="https://github.com/trpc-ecosystem/go-opentelemetry" target="_blank">OpenTelemetry Oteam</a>：OpenTelemetry Oteam 定位于云原生可观测性（监控，分布式追踪，日志） 标准制定和工具建设。

### 1.2 开发环境要求

在开始之前，请确保已经安装了以下软件：
* Git
* Docker 或者其他平替的容器工具。

请确保已了解 <a href="https://trpc.group/zh/docs/languages/go/" target="_blank">tRPC-Go</a> 开发流程。

### 1.3 运行 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/go-examples/trpc-otlp-oteam-open

# 初始化环境
make setup
make pb

# 💡token & otlp_endpoint & otlp_http_endpoint 需要准确配置才有数据上报，请继续阅读文档以了解环境变量用途和取值。
# 💡如需后台运行容器，请加 -d 作为参数
make dev \
  token="xxx" \
  otlp_endpoint="127.0.0.1:4317" \
  otlp_http_endpoint="http://127.0.0.1:4318"
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。

### 2.2 接入

#### 2.2.1 添加依赖

添加以下依赖到项目 `go.mod` 中：

```go
require (
    trpc.group/trpc-go/trpc-opentelemetry v1.0.2
    trpc.group/trpc-go/trpc-opentelemetry/oteltrpc v1.0.2
)
```

#### 2.2.2 匿名引入 tRPC 拦截器

```shell
import _ "trpc.group/trpc-go/trpc-opentelemetry/oteltrpc"
```

#### 2.2.3 修改 tRPC 项目配置文件

框架配置中，数据准确上报到平台的几个重要配置 ：

| 配置                                                                                      | 描述                               | 备注                                      |
|-----------------------------------------------------------------------------------------|----------------------------------|-----------------------------------------|
| `plugins.telemetry.opentelemetry.addr`                                                  | 数据上报地址，替换为前文提及的 `OTLP_ENDPOINT`。 | 请根据页面接入指引提供的 `gRPC` 接入地址进行填写。           |
| `plugins.telemetry.opentelemetry.logs.addr`                                             | 同上。                              | 同上。                                     |
| `plugins.telemetry.opentelemetry.metrics.prometheus_push.url`                           | 同上。                              | 请根据页面接入指引提供的 `HTTP` 地址进行填写。             |
| `plugins.telemetry.opentelemetry.tenant_id`                                             | 租户 ID，替换前文提及的应用 `TOKEN`。         |                                         |
| `plugins.telemetry.opentelemetry.metrics.prometheus_push.http_headers.X-BK-TOKEN`       | 同上。                              |                                         |
| `plugins.telemetry.opentelemetry.attributes[0].value`                                   | 数据上报 `service_name` 字段 *[1]*。    | 例如 `example.greeter`。                   |
| `plugins.telemetry.opentelemetry.metrics.prometheus_push.grouping.service_name`         | 同上。                              |                                         |
| `plugins.telemetry.opentelemetry.metrics.prometheus_push.grouping.instance` *[2]*       | 实例                               | Pod 的 IP 地址（容器部署）或本机 IP（物理机或虚拟机部署）      |
| `plugins.telemetry.opentelemetry.metrics.prometheus_push.grouping.namespace` *[2]*      | 物理环境                             | 请保持和配置文件中的 ${物理环境(global.namespace)} 一致 |
| `plugins.telemetry.opentelemetry.metrics.prometheus_push.grouping.env_name` *[2]*       | 用户环境                             | 请保持和配置文件中的 ${用户环境(global.env_name)} 一致  |
| `plugins.telemetry.opentelemetry.metrics.prometheus_push.grouping.container_name` *[2]* | Pod 的名称                          |                                         |

* [1] 只有准确填写 `sevice_name` 配置，服务才能关联到上报的「自定义指标」。
  * 规则：`${app}.${server}`
  * 例如 tRPC 配置文件中，`server.app=example` & `server.server=greeter`，那么 `sevice_name` 值为 `example.greeter`
* [2] `prometheus_push` 默认不会上报实例唯一标识，这会导致多实例情况下，指标聚合结果不准确，请尽可能将上文提到的实例字段补充上报。

添加 `opentelemetry` 到 `client.filter`、`server.filter`：

```yaml
server:
  filter:
    - opentelemetry

client:
  filter:
    - opentelemetry
```

示例配置（仅列举 SDK 相关配置，完整示例请参考：<a href="https://github.com/trpc-group/trpc-go/blob/main/docs/user_guide/framework_conf.md" target="_blank">trpc_go.yaml</a>）：

```yaml
global:  # 全局配置
  namespace: Development  # 环境类型，分正式 production 和非正式 development 两种类型
  env_name: test  # 环境名称，非正式环境下多环境的名称

server:  # 服务端配置
  app: example  # 业务的应用名
  server: greeter  # 进程服务名
  filter:  # 针对所有 service 处理函数前后的拦截器列表
    - opentelemetry

client:  # 客户端调用的后端配置
  filter:  # 针对所有后端调用函数前后的拦截器列表
    - opentelemetry

plugins:  # 插件配置
  telemetry: # 注意缩进层级关系
    opentelemetry:
      addr: "${otlp_endpoint}"  # ❗️❗️️【非常重要】上报地址，请根据页面接入指引提供的 `gRPC` 接入地址进行填写。
      tenant_id: "${token}"   # ❗️❗️️【非常重要】 替换为页面申请到的 Token
      sampler:
        fraction: 1.0 # 采样（0.0001代表每10000请求上报一次trace数据）
      attributes:
        - key: "service_name"
          value: example.greeter   # ❗️❗️️【非常重要】请使用 ${应用名(server.app)}.${服务名(server.server)}
#        #【可选】如何发现容器信息
#        #【推荐】将上报域名切换为集群内域名（bkm-collector.bkmonitor-operator），端口、上报路径与之前一致，
#        # 即可自动获取关联。
#        #【手动关联】手动补充以下全部集群信息字段，也可以进行关联：
#        # Pod 名称
#        - key: "k8s.pod.name"
#          value: "${K8S_POD_NAME}"
#        # Pod 名称
#        - key: "k8s.namespace.name"
#          value: "${K8S_NAMESPACE}"
#        # 集群 ID
#        - key: "k8s.bcs.cluster.id"
#          value: "BCS-K8S-00000"

      metrics:
        enable_register: false # 关闭etcd注册
        # 可选配置 (default false). 设置为 true 后，上报 metric 时会对被调接口名进行原样上报。
        # 非 restful 的 http 服务需要把 disable_rpc_method_mapping 设置为 true。
        # restful 服务则设置为 false 且需要使用 metric.RegisterMethodMapping 注册 path 与 pattern 映射关系，避免高基数问题。
        disable_rpc_method_mapping: true
        # ❗️【提醒】耗时分桶设置：可根据业务实际情况再微调下，另外分桶数最好也不要超过 15 个，分桶数越多，指标量级将等比例上涨。
        client_histogram_buckets: [ 0, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10 ]
        server_histogram_buckets: [ 0, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10 ]
        prometheus_push: # OpenTelemetry push 配置
          enabled: true # 是否开启 metric OpenTelemetry push，默认 false。
          url: "${otlp_http_endpoint}" # ❗️❗️️【非常重要】上报地址，请根据页面接入指引提供的 `HTTP` 地址进行填写。
          # ❗️【提醒】建议配置 30s，否则可能会影响调用分析数据展示。
          interval: 30s
          job: "reporter"
          grouping:
            service_name: example.greeter   # ❗️❗️️【非常重要】请使用 ${应用名(server.app)}.${服务名(server.server)}
            instance: "${K8S_POD_IP}" # ❗️❗️️【非常重要】Pod 的 IP 地址（容器部署）或本机 IP（物理机或虚拟机部署）
            namespace: "${namespace}" # ❗️❗️️【非常重要】请保持和配置文件中的 ${物理环境(global.namespace)} 一致
            env_name: "${env_name}" # ❗️❗️️【非常重要】请保持和配置文件中的 ${用户环境(global.env_name)} 一致
            container_name: "${K8S_POD_NAME}" # ❗️❗️️【非常重要】Pod 的名称
          http_headers:
            X-BK-TOKEN: "${token}"   # ❗️❗️️【非常重要】 替换为页面申请到的 Token
        # 错误码重定义：codes 可设置特定错误码的类型, 以便计算错误率/超时率/成功率和看板展示错误码描述。
        # 该配置在 Traces、Metrics 同时生效。
        codes: # 可选：
          - code: 500
            type: success
            description: success
            # ❗️提醒：service / method 互斥，优先级 service > method。
            # service: trpc.example.greeter.http
            method: /500
      traces: # 链路延迟采样相关配置
        disable_trace_body: false
        enable_deferred_sample: true # 开启延迟采样 在span结束后的导出采样, 额外上报出错的/高耗时的
        deferred_sample_error: true # 采样出错的
        deferred_sample_slow_duration: 1ms # 采样耗时大于指定值的
      logs:
        enabled: true # 远程日志开关，默认关闭
        level: "debug" # 日志级别，默认error
        trace_log_mode: "verbose"
        addr: "${otlp_endpoint}"   # ❗️❗️️【非常重要】上报地址，请根据页面接入指引提供的 `gRPC` 接入地址进行填写。

```

更多配置项说明，请参考：
* <a href="https://github.com/trpc-ecosystem/go-opentelemetry" target="_blank">OpenTelemetry Oteam Go SDK 指引</a>
* <a href="https://github.com/trpc-group/trpc-go/blob/main/docs/user_guide/framework_conf.md" target="_blank">tRPC-Go 框架配置</a>

## 3. 常见问题

### 3.1 protobuf panic

`google.golang.org/protobuf` 自 v1.26.0 版本开始，<a href="https://protobuf.dev/reference/go/faq/#fix-namespace-conflict" target="_blank">namespace conflict</a> 发生变化，由原来打印 Warning 变为直接 Panic。

由于 OpenTelemetry SDK 自 v0.20.0 开始引入了 v1.26.0, 多个同名 pb 文件时候触发这个特性：启动时直接 panic。

规避方式：增加 `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn` 的环境变量。

### 3.2 如何自动发现容器信息

#### 3.2.1 🌟【推荐】方案 1：通过集群内上报

将上报域名切换为集群内域名（bkm-collector.bkmonitor-operator），端口、上报路径与之前一致，即可自动获取关联。

#### 3.2.2 方案 2：手动关联

手动补充以下全部集群信息字段到 Span Resource，也可以进行关联：

| 字段                 | 描述             | 备注 |
| -------------------- | ---------------- | ---- |
| `k8s.bcs.cluster.id` | 集群 ID          | --   |
| `k8s.pod.name`       | Pod 名称         | --   |
| `k8s.namespace.name` | Pod 所在命名空间 | --   |

除了 `k8s.bcs.cluster.id` 外，可以在相应的 k8s 描述文件（Yaml）中，将 Pod 字段作为环境变量的值：

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

## 4. 使用场景

### 4.1 Traces

暂无。

### 4.2 Metrics

#### 4.2.1 指标 API

直接使用 tRPC-Go 指标 API 即可 👇。

* <a href="https://github.com/trpc-group/trpc-go/blob/main/metrics/README.zh_CN.md" target="_blank">tRPC-Go 指标监控</a>

#### 4.2.2 错误码重定义

可设置特定错误码的类型, 以便计算错误率/超时率/成功率和看板展示错误码描述。

该配置在 Traces、Metrics 同时生效。

```yaml
plugins:
  telemetry: # 注意缩进层级关系
    opentelemetry:
      metrics:
        # 默认值: 0:成功 success 21/101:超时timeout 其它：错误 exception
        codes: # 可选：
          - code: 21
            type: timeout
            description: server 超时
          - code: 101
            type: timeout
            description: client 超时
          # 下面为设置特定返回码的例子，业务可按需设置。
          - code: 100014
            # type 为 success 表示 100014 这个返回码 (无论主被调) 会被统计为成功。
            # 不区分主被调，如果担忧错误码冲突，可以设置 service 和 method 来限定生效的 service 和 method。
            type: success
            description: desc4 # 对这个返回码的具体描述
          - code: 100015
            type: exception # type 为 exception 表示 100015 是个异常的错误码。可在 description 里设置更详细的说明信息。
            description: desc5
            # ❗️提醒：service / method 互斥，优先级 service > method。
            service: # 不为空表示错误码特例仅匹配特定的 (无论主被调) service, 为空表示所有 service。
            method: # 不为空表示错误码特例仅匹配特定的 (无论主被调) method, 为空表示所有 method。
```

#### 4.2.3 HTTP 场景被调接口展示具体接口名

SDK 默认屏蔽 HTTP 被调接口以减少上报基数，请按如下指引按需开启。

1）场景 1 - 非 Restful 服务

接口路径不携带 `Id` / `?a=xxx&b=xx` 等不可枚举值，调整 `disable_rpc_method_mapping` 配置以原样上报被调接口名。

```yaml
plugins:
  telemetry: # 注意缩进层级关系
    opentelemetry:
      metrics:
        disable_rpc_method_mapping: true
```

2）场景 2 - Restful 服务

接口路径携带 `Id` / `?a=xxx&b=xx` 等不可枚举值，直接上报很可能引起时序高基数问题，导致 SDK 占用内存多、模调监控视图加载严重变慢。

需要在服务启动前，注册 `method regex` -> `pattern` 映射，将含有不可枚举值的高基数 `method` 转换为低基数的 `method pattern`。

```go
import (
    "trpc.group/trpc-go/trpc-opentelemetry/sdk/metric"
)

func registerMethodMapping() {
    // 将不可枚举部分替换成占位符。
    metric.RegisterMethodMapping("/v1/foobar/[a-z0-9A-Z]+", "/v1/foobar/{vid}")
}
```

#### 4.2.4 HTTP 场景主调指标减少被调接口上报基数

服务作为 client 请求其他 HTTP 服务，接口路径携带 `Id` / `?a=xxx&b=xx` 等不可枚举值，直接上报很可能引起时序高基数问题，导致 SDK 占用内存多、模调监控视图加载严重变慢。

tRPC 请求 HTTP 将请求路径设置在 `msg.ClientRPCName`，`msg.CalleeMethod` 允许自定义。

```go
package main

import (
  "time"
  "trpc.group/trpc-go/trpc-go"
  "trpc.group/trpc-go/trpc-go/client"
  http "trpc.group/trpc-go/trpc-go/http"
)

func periodicHTTPGet() {
  for {
    cli := http.NewClientProxy("trpc.example.greeter.http", client.WithTimeout(time.Second*2))
    // 关键代码，通过 client.WithCalleeMethod("xxxx") 将 path 映射为固定名称以降低基数。
    cli.Get(trpc.BackgroundContext(), "/trpc_info_test/guc_info/6666", nil, client.WithCalleeMethod("GetGucInfo"))
    cli.Get(trpc.BackgroundContext(), "/trpc_info_test/guc_info/6666/_update", nil, client.WithCalleeMethod("UpdateGucInfo"))
  }
}
```

### 4.3 Logs

直接使用 tRPC-Go 日志 API 即可 👇。

* <a href="https://github.com/trpc-group/trpc-go/blob/main/log/README.zh_CN.md" target="_blank">tRPC-Go 日志管理</a>

## 5. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>