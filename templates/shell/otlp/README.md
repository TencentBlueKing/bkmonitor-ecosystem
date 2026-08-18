# 调用链（HTTP）上报

## 1. 概述

调用链 HTTP 上报使用 OpenTelemetry Protocol（OTLP）将 Span 发送到蓝鲸应用性能监控。本文从一条可运行的
`curl` 请求开始，说明 OTLP/HTTP 请求结构、Span 核心字段和父子关系。

直接发送 HTTP 请求适合协议联调和轻量接入。生产服务优先使用 OpenTelemetry SDK，由 SDK 处理上下文传播、
采样、批量发送和失败重试。

## 2. 准备开始

### 2.1 创建应用

{{APPLICATION_SET_UP}}

### 2.2 数据协议

OTLP/HTTP 调用链上报使用 HTTP POST 请求。默认路径为 `/v1/traces`，请求体是
`ExportTraceServiceRequest` 的 JSON Protobuf 表示。

关键上报配置：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `TOKEN` | 是 | ❗❗【非常重要】APM 应用 Token，上报时必须通过 `x-bk-token` Header 传递。 |
| `SERVICE_NAME` | 是 | ❗❗【非常重要】服务唯一标识，写入 Resource 的 `service.name`；一个 APM 应用可以包含多个服务。 |
| `OTLP_ENDPOINT` | 是 | ❗❗【非常重要】OTLP HTTP 根地址。国内站点默认是「{{access_config.otlp.http_endpoint}}」，其他环境、跨云场景请根据页面接入指引填写；上报调用链时在末尾追加 `/v1/traces`。 |

#### 2.2.1 发送一条调用链

以下请求发送两个 Span：服务端 Span 表示结账接口，客户端 Span 表示它调用支付服务。两个 Span 使用相同的
`traceId`，客户端 Span 通过 `parentSpanId` 指向服务端 Span。

```shell
#!/bin/bash
# ❗❗【非常重要】认证令牌，请替换为页面提供的 APM 应用 Token。
TOKEN="fixme"

# ❗❗【非常重要】服务唯一标识，最终写入 Resource 的 service.name。
SERVICE_NAME="http-trace-demo"

# ❗❗【非常重要】OTLP HTTP 根地址。国内站点默认是「{{access_config.otlp.http_endpoint}}」，其他环境、跨云场景请根据页面接入指引填写。
OTLP_ENDPOINT="{{access_config.otlp.http_endpoint}}"

TRACE_ID="$(openssl rand -hex 16)"
SERVER_SPAN_ID="$(openssl rand -hex 8)"
CLIENT_SPAN_ID="$(openssl rand -hex 8)"
END_TIME_UNIX_NANO="$(($(date +%s) * 1000000000))"
START_TIME_UNIX_NANO="$((END_TIME_UNIX_NANO - 2000000000))"
CLIENT_START_TIME_UNIX_NANO="$((END_TIME_UNIX_NANO - 1500000000))"
CLIENT_END_TIME_UNIX_NANO="$((END_TIME_UNIX_NANO - 500000000))"

REPORT_DATA=$(cat <<EOF
{
  "resourceSpans": [
    {
      "resource": {
        "attributes": [
          { "key": "service.name", "value": { "stringValue": "${SERVICE_NAME}" } },
          { "key": "service.version", "value": { "stringValue": "1.0.0" } },
          { "key": "deployment.environment.name", "value": { "stringValue": "local" } }
        ]
      },
      "scopeSpans": [
        {
          "scope": {
            "name": "curl-demo",
            "version": "1.0.0"
          },
          "spans": [
            {
              "traceId": "${TRACE_ID}",
              "spanId": "${SERVER_SPAN_ID}",
              "flags": 1,
              "name": "POST /checkout",
              "kind": 2,
              "startTimeUnixNano": "${START_TIME_UNIX_NANO}",
              "endTimeUnixNano": "${END_TIME_UNIX_NANO}",
              "attributes": [
                { "key": "http.request.method", "value": { "stringValue": "POST" } },
                { "key": "url.scheme", "value": { "stringValue": "https" } },
                { "key": "url.path", "value": { "stringValue": "/checkout" } },
                { "key": "http.route", "value": { "stringValue": "/checkout" } },
                { "key": "server.address", "value": { "stringValue": "shop.example.com" } },
                { "key": "server.port", "value": { "intValue": "443" } },
                { "key": "http.response.status_code", "value": { "intValue": "500" } },
                { "key": "error.type", "value": { "stringValue": "500" } }
              ],
              "status": {
                "code": 2
              }
            },
            {
              "traceId": "${TRACE_ID}",
              "spanId": "${CLIENT_SPAN_ID}",
              "parentSpanId": "${SERVER_SPAN_ID}",
              "flags": 1,
              "name": "POST",
              "kind": 3,
              "startTimeUnixNano": "${CLIENT_START_TIME_UNIX_NANO}",
              "endTimeUnixNano": "${CLIENT_END_TIME_UNIX_NANO}",
              "attributes": [
                { "key": "http.request.method", "value": { "stringValue": "POST" } },
                { "key": "url.full", "value": { "stringValue": "https://payment.example.com/authorize" } },
                { "key": "server.address", "value": { "stringValue": "payment.example.com" } },
                { "key": "server.port", "value": { "intValue": "443" } },
                { "key": "http.response.status_code", "value": { "intValue": "200" } }
              ]
            }
          ]
        }
      ]
    }
  ]
}
EOF
)

curl -sS -X POST "${OTLP_ENDPOINT%/}/v1/traces" \
  -H "x-bk-token: ${TOKEN}" \
  --json "${REPORT_DATA}"
```

完整接收时返回：

```json
{}
```

HTTP `200` 的响应也可能包含 `partialSuccess`。`rejectedSpans` 大于 `0` 表示部分 Span 被拒绝，此时根据
`errorMessage` 修正请求，不要重试原请求。

#### 2.2.2 请求结构

`ExportTraceServiceRequest`：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `resourceSpans[]` | `array<ResourceSpans>` | 是 | 按资源分组的 Span 集合。 |

`ResourceSpans`：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `resourceSpans[].resource` | `Resource` | 否 | 产生当前 Span 集合的服务、容器、Pod 或进程。 |

`Resource`：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `resource.attributes[]` | `array<KeyValue>` | 否 | 资源属性。❗❗【非常重要】必须设置 `service.name`，用于区分应用下的服务。 |

`ScopeSpans`：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `scopeSpans[].scope` | `InstrumentationScope` | 否 | 生成 Span 的埋点库或模块。 |
| `scopeSpans[].spans[]` | `array<Span>` | 是 | 当前 scope 生成的 Span 列表。 |

`InstrumentationScope`：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `scope.name` | `string` | 否 | 埋点库或模块名称。 |
| `scope.version` | `string` | 否 | 埋点库或模块版本。 |

#### 2.2.3 Span 核心字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `traceId` | `string` | 是 | `16` 字节 Trace ID，OTLP JSON 使用 `32` 位十六进制字符串；全零值无效。 |
| `spanId` | `string` | 是 | `8` 字节 Span ID，OTLP JSON 使用 `16` 位十六进制字符串；同一 Trace 内保持唯一，全零值无效。 |
| `parentSpanId` | `string` | 否 | 父 Span ID。根 Span 省略该字段，子 Span 填写父 Span 的 `spanId`。 |
| `flags` | `integer` | 否 | 位标记。低 `8` 位对应 W3C Trace Context flags，值 `1` 表示 sampled。 |
| `name` | `string` | 是 | 操作名称，必须是非空字符串；同一操作使用稳定名称。 |
| `kind` | `integer` | 否 | Span 类型，见 `kind` 取值；建议显式设置。 |
| `startTimeUnixNano` | `string` | 是 | 开始时间，Unix 纳秒时间戳。OTLP JSON 中 `fixed64` 使用十进制字符串。 |
| `endTimeUnixNano` | `string` | 是 | 结束时间，Unix 纳秒时间戳，必须大于或等于 `startTimeUnixNano`。 |
| `attributes[]` | `array<KeyValue>` | 否 | 当前操作的属性，同一列表内的 key 必须唯一。 |
| `events[]` | `array<Event>` | 否 | Span 生命周期内发生的带时间戳事件。 |
| `status` | `Status` | 否 | 操作最终状态。省略时等价于 `status.code=0`，见 `status.code` 取值。 |

`Event`：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `events[].timeUnixNano` | `string` | 是 | Event 发生时间，Unix 纳秒时间戳。 |
| `events[].name` | `string` | 是 | Event 名称，必须是非空字符串。 |
| `events[].attributes[]` | `array<KeyValue>` | 否 | Event 属性，同一列表内的 key 必须唯一。 |

`kind` 取值：

| 值 | 名称 | 适用场景 |
| --- | --- | --- |
| `0` | `SPAN_KIND_UNSPECIFIED` | 未指定。不要主动使用；接收端可能按 `INTERNAL` 处理。 |
| `1` | `SPAN_KIND_INTERNAL` | 应用内部操作，不跨进程边界。 |
| `2` | `SPAN_KIND_SERVER` | 服务端处理入站远程请求。 |
| `3` | `SPAN_KIND_CLIENT` | 客户端发起出站远程请求。 |
| `4` | `SPAN_KIND_PRODUCER` | 生产者发送消息或调度异步任务。 |
| `5` | `SPAN_KIND_CONSUMER` | 消费者接收并处理消息或异步任务。 |

`Status`：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `status.code` | `integer` | 否 | 最终状态码，默认值为 `0`。 |
| `status.message` | `string` | 否 | 面向开发者的错误说明，仅在 `status.code=2` 时使用。 |

`status.code` 取值：

| 值 | 名称 | 含义 |
| --- | --- | --- |
| `0` | `STATUS_CODE_UNSET` | 未设置，不等同于显式成功。没有错误时通常保留该值。 |
| `1` | `STATUS_CODE_OK` | 开发者或运维人员已确认操作成功。一般埋点库不主动设置。 |
| `2` | `STATUS_CODE_ERROR` | 操作发生错误。 |

#### 2.2.4 Resources 与 Attributes

Resources 描述“谁产生 Span”，Attributes 描述“当前 Span 做了什么”。稳定的服务、环境、Pod 信息放在
`resource.attributes`；请求方法、路由、状态码和错误类型放在 `span.attributes`。

Attribute 由唯一的 key 和任意值组成，支持字符串、布尔值、数值、数组、对象和字节数据。

常用 Resource 属性：

* `service.name`：❗❗【非常重要】服务唯一标识，一个 APM 应用可以包含多个服务。
* `service.version`：服务版本。
* `deployment.environment.name`：部署环境，例如 `prod`、`staging`。
* `k8s.pod.name`：Kubernetes Pod 名称。

HTTP Span 应遵循 OpenTelemetry HTTP 语义约定。服务端 Span 的 `kind` 使用 `2`，客户端 Span 的 `kind`
使用 `3`；属性使用 `http.request.method`、`http.route`、`url.path`、`server.address` 和
`http.response.status_code` 等标准名称。

* <a href="https://opentelemetry.io/docs/concepts/resources/" target="_blank">OpenTelemetry Resources</a>
* <a href="https://opentelemetry.io/docs/specs/semconv/http/http-spans/" target="_blank">OpenTelemetry HTTP Span semantic conventions</a>

## 3. 了解更多

{{LEARN_MORE}}
