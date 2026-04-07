# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

# pylint: disable=line-too-long

from . import base


class EcosystemRepositoryName(metaclass=base.FieldMeta):
    class Meta:
        name = "ECOSYSTEM_REPOSITORY_NAME"
        scope = base.ScopeType.OPEN.value
        value = "bkmonitor-ecosystem"


class EcosystemRepositoryUrl(metaclass=base.FieldMeta):
    class Meta:
        name = "ECOSYSTEM_REPOSITORY_URL"
        scope = base.ScopeType.OPEN.value
        value = "https://github.com/TencentBlueKing/bkmonitor-ecosystem"


class EcosystemCodeRootUrl(metaclass=base.FieldMeta):
    class Meta:
        name = "ECOSYSTEM_CODE_ROOT_URL"
        scope = base.ScopeType.OPEN.value
        value = "https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main"


class ReferPythonOtlpUrl(metaclass=base.FieldMeta):
    class Meta:
        name = "REFER_PYTHON_OTLP_URL"
        scope = base.ScopeType.OPEN.value
        value = "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/python/otlp/README.md"


class ReferJavaOtlpUrl(metaclass=base.FieldMeta):
    class Meta:
        name = "REFER_JAVA_OTLP_URL"
        scope = base.ScopeType.OPEN.value
        value = "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/java/otlp/README.md"


class ReferGolangTrpcOteamUrl(metaclass=base.FieldMeta):
    class Meta:
        name = "REFER_GOLANG_TRPC_OTEAM_URL"
        scope = base.ScopeType.OPEN.value
        value = ""


class TermIntro(metaclass=base.FieldMeta):
    class Meta:
        name = "TERM_INTRO"
        scope = base.ScopeType.OPEN.value
        value = """* Traces：<a href="https://opentelemetry.io/docs/concepts/signals/traces/" target="_blank">调用链</a>，表示请求在应用程序的执行路径。
* Metrics：<a href="https://opentelemetry.io/docs/concepts/signals/metrics/" target="_blank">指标</a>，表示对运行服务的测量。
* Logs: <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">日志</a>，表示对事件的记录。
* Profiling: <a href="https://grafana.com/docs/pyroscope/latest/introduction/profiling/" target="_blank">性能分析</a>，表示对应用程序运行时资源的持续测量。
* Telemetry Data：观测数据，指代 Traces、Metrics、Logs、Profiling 等。
* APM：蓝鲸观测平台应用性能监控，提供四类观测数据开箱即用的观测能力。
* <a href="https://github.com/TencentBlueKing/bkmonitor-datalink/tree/main/pkg/collector" target="_blank">bk-collector</a>：腾讯蓝鲸的 APM 服务端组件，负责接收 Prometheus、OpenTelemetry、Jaeger、Skywalking 等主流开源组件的观测数据，并对数据进行清洗转发到观测平台链路。"""


class ReferHelloWorldUrl(metaclass=base.FieldMeta):
    class Meta:
        name = "REFER_HELLO_WORLD_URL"
        scope = base.ScopeType.OPEN.value
        value = "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/examples/helloworld.md"


class Overview(metaclass=base.FieldMeta):
    class Meta:
        name = "OVERVIEW"
        scope = base.ScopeType.OPEN.value
        value = """本指南将帮助您使用 OpenTelemetry SDK 接入蓝鲸应用性能监控，以 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/examples/helloworld.md" target="_blank">入门项目-HelloWorld</a> 为例，介绍调用链、指标、日志数据接入及 SDK 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速入门 OpenTelemetry，接入并体验蓝鲸应用性能监控相关功能。"""


class QuickStartOverview(metaclass=base.FieldMeta):
    class Meta:
        name = "QUICK_START_OVERVIEW"
        scope = base.ScopeType.OPEN.value
        value = """本示例仅演示如何将 <a href="https://opentelemetry.io/docs/concepts/signals/traces/" target="_blank">Traces</a>、<a href="https://opentelemetry.io/docs/concepts/signals/metrics/" target="_blank">Metrics</a>、<a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs</a>、<a href="https://grafana.com/docs/pyroscope/latest/introduction/profiling/" target="_blank">Profiling</a> 四类观测数据接入蓝鲸应用性能监控。"""


class ViewData(metaclass=base.FieldMeta):
    class Meta:
        name = "VIEW_DATA"
        scope = base.ScopeType.OPEN.value
        value = """> TODO"""


class ViewTracesData(metaclass=base.FieldMeta):
    class Meta:
        name = "VIEW_TRACES_DATA"
        scope = base.ScopeType.OPEN.value
        value = """Tracing 检索功能主要用于对分布式系统中的请求链路进行跟踪和分析，请参考[「应用性能监控 APM/调用链追踪」](https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/explore_traces.md) 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/traces.png)"""


class ViewCustomMetricsData(metaclass=base.FieldMeta):
    class Meta:
        name = "VIEW_CUSTOM_METRICS_DATA"
        scope = base.ScopeType.OPEN.value
        value = """自定义指标功能旨在帮助用户针对特定应用及其服务进行深度性能指标监控，请参考[「应用性能监控 APM/自定义指标」](#) 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/custom-metrics.png)"""


class ViewLogData(metaclass=base.FieldMeta):
    class Meta:
        name = "VIEW_LOG_DATA"
        scope = base.ScopeType.OPEN.value
        value = """日志功能主要用于查看和分析对应服务（应用程序）运行过程中产生的各类日志信息，请参考[「应用性能监控 APM/日志分析」](#) 进一步了解相关功能。
![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/logs.png)"""


class LearnMore(metaclass=base.FieldMeta):
    class Meta:
        name = "LEARN_MORE"
        scope = base.ScopeType.OPEN.value
        value = """* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>"""


class MustConfigResources(metaclass=base.FieldMeta):
    class Meta:
        name = "MUST_CONFIG_RESOURCES"
        scope = base.ScopeType.OPEN.value
        value = """请在 <a href="https://opentelemetry.io/docs/concepts/resources/" target="_blank">Resources</a> 添加以下属性，蓝鲸观测平台通过这些属性，将数据关联到具体的应用、资源实体：

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
```"""


class ApplicationSetUp(metaclass=base.FieldMeta):
    class Meta:
        name = "APPLICATION_SET_UP"
        scope = base.ScopeType.OPEN.value
        value = """参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/1-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `OTLP_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。"""


class ProfilingApplicationSetUp(metaclass=base.FieldMeta):
    class Meta:
        name = "PROFILING_APPLICATION_SET_UP"
        scope = base.ScopeType.OPEN.value
        value = """参考 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/scene-apm/apm_monitor_overview.md" target="_blank">APM 接入流程</a> 创建一个应用，接入指引会基于应用生成相应的上报配置，如下：

![](https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/common/images/profiling-application-setup.png)

关注接入指引提供的两个配置项：

- `TOKEN`：上报唯一凭证。

- `PROFILING_ENDPOINT`：数据上报地址。

有任何问题可企微联系 `BK助手` 协助处理。"""


class MustConfigExporter(metaclass=base.FieldMeta):
    class Meta:
        name = "MUST_CONFIG_EXPORTER"
        scope = base.ScopeType.OPEN.value
        value = """请在创建 <a href="https://opentelemetry.io/docs/specs/otel/protocol/exporter/" target="_blank">Exporter</a> 时准确传入以下信息：

| 参数         | 说明                            |
|------------|-------------------------------|
| `endpoint` | 【必须】数据上报地址，请根据页面指引提供的接入地址进行填写。 |
| `x-bk-token`| 【必须】APM 应用 Token，作为 headers 传递。 |"""


class MustConfigProfiling(metaclass=base.FieldMeta):
    class Meta:
        name = "MUST_CONFIG_PROFILING"
        scope = base.ScopeType.OPEN.value
        value = """<a href="https://grafana.com/docs/pyroscope/latest/" target="_blank">Pyroscope</a> 是 Grafana 旗下用于聚合连续分析数据的开源软件项目。

请在创建 `PyroscopeConfig` 时，准确传入以下信息：

| 属性                | 说明                                            |
|-------------------|-----------------------------------------------|
| `AuthToken`       | 【必须】APM 应用 `Token`                            |
| `ApplicationName` | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分                |
| `ServerAddress`   | 【必须】Profiling 数据上报地址，请根据页面指引提供的 HTTP 接入地址进行填写 |"""


class DemoRunParameters(metaclass=base.FieldMeta):
    class Meta:
        name = "DEMO_RUN_PARAMETERS"
        scope = base.ScopeType.OPEN.value
        value = """运行参数说明：

| 参数                   | 推荐值                                | 说明                                        |
|----------------------|--------------------------------------|-------------------------------------------|
| `TOKEN`              | `""`                                 | APM 应用 `Token`。                            |
| `SERVICE_NAME`       | `"helloworld"`                       | 服务唯一标识，一个应用可以有多个服务，通过该属性区分。                |
| `OTLP_ENDPOINT`      | `"http://127.0.0.1:4318"` | OT 数据上报地址，请根据页面指引提供的接入地址进行填写，支持以下协议：<br />`gRPC`：`http://127.0.0.1:4317`<br />`HTTP`：`http://127.0.0.1:4318`（demo 使用该协议演示上报）<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。        |
| `PROFILING_ENDPOINT` | `"http://127.0.0.1:4318/pyroscope"`  | Profiling 数据上报地址，请根据页面指引提供的 HTTP 接入地址进行填写。<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。 |
| `ENABLE_TRACES`      | `false`                              | 是否启用调用链上报。                                 |
| `ENABLE_METRICS`     | `false`                              | 是否启用指标上报。                                  |
| `ENABLE_LOGS`        | `false`                              | 是否启用日志上报。                                  |
| `ENABLE_PROFILING`   | `false`                            | 是否启用性能分析上报。                                |"""


class SWDemoRunParameters(metaclass=base.FieldMeta):
    class Meta:
        name = "SW_DEMO_RUN_PARAMETERS"
        scope = base.ScopeType.OPEN.value
        value = """运行参数说明：

| 参数                       | 推荐值                               | 说明                                        |
|---------------------------|--------------------------------------|-------------------------------------------|
| `SW_AGENT_NAME`           | `"helloworld-java-sw"`               | 【必须】服务唯一标识，用于表示提供相同功能/逻辑的逻辑组。                |
| `SW_AGENT_COLLECTOR_BACKEND_SERVICES`      | `"127.0.0.1:4317"`  | 【必须】数据上报地址，请根据页面指引提供的接入地址进行填写。（demo 使用该协议演示上报）<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。        |
| `SW_AGENT_AUTHENTICATION` | `"x-bk-token"`                                 | 【必须】上报数据时需要的认证信息。</br>【非常重要】x-bk-token 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。。                            |
"""


class JaegerDemoRunParameters(metaclass=base.FieldMeta):
    class Meta:
        name = "JAEGER_DEMO_RUN_PARAMETERS"
        scope = base.ScopeType.OPEN.value
        value = """运行参数说明：

| 参数                   | 推荐值                                | 说明                                        |
|----------------------|--------------------------------------|-------------------------------------------|
| `TOKEN`              | `""`                                 | APM 应用 `Token`。                            |
| `SERVICE_NAME`       | `"jaeger-client-demo-go"`                       | 服务唯一标识，一个应用可以有多个服务，通过该属性区分。                |
| `OTLP_ENDPOINT`      | `"http://127.0.0.1:4318"` | OT 数据上报地址，请根据页面指引提供的接入地址进行填写，支持以下协议：<br />`HTTP`：`http://127.0.0.1:4318`（demo 使用该协议演示上报）<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。        |
| `ENABLE_TRACES`      | `false`                              | 是否启用调用链上报。                                 |"""


class AccessConfig(metaclass=base.FieldMeta):
    class Meta:
        name = "access_config"
        scope = base.ScopeType.OPEN.value
        value = {
            "token": "xxx",
            "otlp": {
                "enable_traces": "true",
                "enable_metrics": "true",
                "enable_logs": "true",
                "grpc_endpoint": "127.0.0.1:4317",
                "endpoint": "http://127.0.0.1:4317",
                "http_endpoint": "http://127.0.0.1:4318",
                "http_endpoint_without_schema": "127.0.0.1:4318",
            },
            "profiling": {
                "endpoint": "http://127.0.0.1:4318/pyroscope",
                "enabled": "true",
                "enable_memory_profiling": "true",
            },
            "sw": {
                "endpoint": "127.0.0.1:4317",
            },
            "custom": {
                "http": "http://127.0.0.1:10205/v2/push/",
                "sdk": "127.0.0.1:4318",
            },
        }


class ServiceName(metaclass=base.FieldMeta):
    class Meta:
        name = "service_name"
        scope = base.ScopeType.OPEN.value
        value = "helloworld"


class ProfilingOverview(metaclass=base.FieldMeta):
    class Meta:
        name = "PROFILING_OVERVIEW"
        scope = base.ScopeType.OPEN.value
        value = """本指南将帮助您使用 Pyroscope SDK 接入蓝鲸应用性能监控，以 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/common/examples/helloworld.md" target="_blank">入门项目-HelloWorld</a> 为例，介绍性能分析数据接入及 SDK 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速接入并体验蓝鲸应用性能监控相关功能。"""


class ProfilingRunParameters(metaclass=base.FieldMeta):
    class Meta:
        name = "PROFILING_RUN_PARAMETERS"
        scope = base.ScopeType.OPEN.value
        value = """运行参数说明：

| 参数                   | 推荐值                                | 说明                                        |
|----------------------|------------------------------------|-------------------------------------------|
| `TOKEN`              | `""`                               | APM 应用 `Token`。                            |
| `PROFILING_ENDPOINT` | `"http://127.0.0.1:4318/pyroscope"` | Profiling 数据上报地址，请根据页面指引提供的 HTTP 接入地址进行填写。<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。 |
| `SERVICE_NAME`       | `"helloworld"`                     | 服务唯一标识，一个应用可以有多个服务，通过该属性区分。                |
| `ENABLE_PROFILING`   | `false`                            | 是否启用性能分析上报。                                |

💡 为保证数据能上报到平台，`TOKEN`、`PROFILING_ENDPOINT` 请务必根据应用接入指引提供的实际值填写。"""


class QuickStartRunParameters(metaclass=base.FieldMeta):
    class Meta:
        name = "QUICK_START_RUN_PARAMETERS"
        scope = base.ScopeType.OPEN.value
        value = """| 参数                 | 值（根据所填写接入信息生成）             | 说明                                                         |
| -------------------- | :--------------------------------------- | ------------------------------------------------------------ |
| `TOKEN`              | `"{{access_config.token}}"`              | 【必须】APM 应用 `Token`。                                     |
| `SERVICE_NAME`       | `"{{service_name}}"`                     | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分。 |
| `OTLP_ENDPOINT`      | `"{{access_config.otlp.http_endpoint}}"` | 【必须】OT 数据上报地址，支持以下协议：<br />  `gRPC`：`{{access_config.otlp.endpoint}}`<br /> `HTTP`：`{{access_config.otlp.http_endpoint}}`（demo 使用该协议演示上报） |
| `PROFILING_ENDPOINT` | `"{{access_config.profiling.endpoint}}"` | 【可选】Profiling 数据上报地址。                               |
| `ENABLE_TRACES`      | `{{access_config.otlp.enable_traces}}`   | 是否启用调用链上报。                                           |
| `ENABLE_METRICS`     | `{{access_config.otlp.enable_metrics}}`  | 是否启用指标上报。                                             |
| `ENABLE_LOGS`        | `{{access_config.otlp.enable_logs}}`     | 是否启用日志上报。                                             |
| `ENABLE_PROFILING`   | `{{access_config.profiling.endpoint}}`   | 是否启用性能分析上报。                                         |

* *<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/" target="_blank">OTLP Exporter Configuration</a>*"""


class AutomaticRunParameters(metaclass=base.FieldMeta):
    class Meta:
        name = "AUTOMATIC_RUN_PARAMETERS"
        scope = base.ScopeType.OPEN.value
        value = """| 环境变量名称                                             | 推荐值                        | 说明                                                                                                                                                                                                                                                                                                                                                                                               |
|----------------------------------------------------|----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `OTEL_SERVICE_NAME`                                | `"${服务名称，请根据右侧说明填写}"`   | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分。优先级比资源属性的设置高，更多信息请参考<a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name" target="_blank">服务名配置</a>。                                                                                                                                                                                                                     |
| `OTEL_EXPORTER_OTLP_HEADERS`                         | `"x-bk-token=todo"` | 【必须】Exporter 导出数据时附加额外的 Headers，多个 key-value 以逗号分隔，例如：`key1=value1,key2=value2`。</br>【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。 |
| `OTEL_EXPORTER_OTLP_PROTOCOL`                      | `"http/protobuf"`          | 【必须】指定<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#protocol-configuration" target="_blank">上报协议</a>，上报协议改变时，上报地址也需要手动修改。</br>【推荐】`protobuf/http`：使用 HTTP 协议上报。</br>【可选】`grpc`：使用 gRPC 上报，如果使用该方式，请确保 `OTEL_EXPORTER_OTLP_ENDPOINT` 也同步改为 gRPC 上报地址。                                                                                                          |
| `OTEL_EXPORTER_OTLP_ENDPOINT`                      | `"http://127.0.0.1:4318"`                         | 【必须】数据<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_endpoint" target="_blank">上报地址</a>，请根据页面指引提供的接入地址进行填写。支持以下协议：<br />`gRPC`：`http://127.0.0.1:4317`<br />`HTTP`：`http://127.0.0.1:4318`（demo 使用该协议演示上报）<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。                                                                                                                                                                                                                   |
| `OTEL_TRACES_EXPORTER`                             | `"otlp"`                   | 【可选】指定用于 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_traces_exporter" target="_blank">Traces Exporter</a>，值为 `"console,otlp"` 时，可以同时在控制台输出。                                                                                                                                                                                                                  |
| `OTEL_METRICS_EXPORTER`                            | `"otlp"`                   | 【可选】指定用于 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_metrics_exporter" target="_blank">Metrics Exporter</a>，值为 `"console,otlp"` 时，可以同时在控制台输出。                                                                                                                                                                                                                |
| `OTEL_LOGS_EXPORTER`                               | `"otlp"`                   | 【可选】指定用于 <a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_logs_exporter" target="_blank">Logs Exporter</a>，值为 `"console,otlp"` 时，可以同时在控制台输出。 |
| `OTEL_RESOURCE_ATTRIBUTES`                         | `""` | 【可选】<a href="https://opentelemetry.io/docs/concepts/resources/" target="_blank">Resource</a> 代表观测数据所属的资源实体，并通过资源属性进行描述。<a href="https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_resource_attributes" target="_blank">Resource Attributes</a> 设置，多个 key-value 以逗号分隔，例如：`key1=value1,key2=value2`。参考下一小节 `服务信息`。 |"""


class GoTrpcOpenTips(metaclass=base.FieldMeta):
    class Meta:
        name = "GO_TRPC_OPEN_TIPS"
        scope = base.ScopeType.OPEN.value
        value = """**本篇文档面向开源版 tRPC-Go，如果您使用的是 <a href="https://github.com/trpc-group/trpc-go" target="_blank">trpc-go/trpc-go</a>（内部版 tRPC-Go【大部分场景下都是内部版】），请参考 [Go（tRPC 云观 Oteam SDK）接入](#) 。**"""


class BlueApps4TraceRunParameters(metaclass=base.FieldMeta):
    class Meta:
        name = "BLUE_APPS4_TRACE_RUN_PARAMETERS"
        scope = base.ScopeType.OPEN.value
        value = """访问任意API路径, 用于上报 trace 数据

在 `蓝鲸监控平台(bkmonitor)` 的**数据检索**页，在侧边栏的空间管理找到自己的应用空间，并选择自己的应用，查看 trace 数据，如下图："""


class BlueApps4MetricsRunParameters(metaclass=base.FieldMeta):
    class Meta:
        name = "BLUE_APPS4_METRICS_RUN_PARAMETERS"
        scope = base.ScopeType.OPEN.value
        value = """访问 metrics 相关路径，用于上报 metrics 数据

在 `蓝鲸监控平台(bkmonitor)` 的**仪表盘**页，在侧边栏的**空间管理**找到自己的应用空间，并选择自己的应用，在 General 栏找仪表盘 —— `bksaas/framework-python` 查看上报的 metrics 数据，如下图："""


class BlueApps5MetricsRunParameters(metaclass=base.FieldMeta):
    class Meta:
        name = "BLUE_APPS5_METRICS_RUN_PARAMETERS"
        scope = base.ScopeType.OPEN.value
        value = """在 `蓝鲸监控平台(bkmonitor)` 的**仪表盘**页，在侧边栏的**空间管理**找到自己的应用空间，并选择自己的应用，在 General 栏找仪表盘 —— `bksaas/framework-python` 查看上报的 metrics 数据，如下图：
"""


class BlueAppsMetricsDataRunParameters(metaclass=base.FieldMeta):
    class Meta:
        name = "BLUE_APPS5_METRICS_DATA_RUN_PARAMETERS"
        scope = base.ScopeType.OPEN.value
        value = """在 `蓝鲸监控平台(bkmonitor)` 的**数据检索**页，在侧边栏的**空间管理**找到自己的应用空间，并选择自己的应用，在侧边栏找到**指标检索**，在指标检索页的搜索框查找自己定义的业务指标，如下图：
"""


class DocsAccessConfig(metaclass=base.FieldMeta):
    class Meta:
        name = "docs"
        scope = base.ScopeType.OPEN.value
        value = {
            "metrics": {
                "What": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Term/metrics/what.md",
                "Types": "{{COOKBOOK_METRICS_TYPES}}",
                "http": {
                    "readme": {
                        "faq_different_protocols": "#",
                        "HTTP_Custom_Report": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/http/README.md",
                        "metrics_http_readme": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/http/README.md",
                        "Http_Curl": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/http/curl.md",
                        "Http_Python": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/http/python.md",
                        "Http_C": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/http/cpp.md",
                        "Http_Java": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/http/java.md",
                        "Http_Go": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/http/go.md",
                    }
                },
                "learn": {
                    "Index_search": "#",
                    "Use_indicators": "#",
                    "configure_dashboard": "https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/dashboard.md",
                    "alarms": "https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/alarm-configurations/rules.md",
                    "SDK_Python": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/sdks/python.md",
                    "SDK_C": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/sdks/cpp.md",
                    "SDK_Java": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/sdks/java.md",
                    "SDK_Go": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/sdks/go.md",
                    "metrics_sdks_readme": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/sdks/README.md",
                },
                "sdks": {"sdk_summary": "#"},
            },
            "events": {
                "Monitor_collector_instal": "https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/collectors/install.md",
                "readme": {
                    "faq_no_data": "#",
                },
                "http": {
                    "report_curl": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/events/http/curl.md",
                    "report_python": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/events/http/python.md",
                    "report_bkmonitorbeat": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/events/http/bkmonitorbeat.md",
                    "Http_Preadme": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/events/http/README.md",
                    "report_java": "https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/events/http/java.md",
                },
                "report_access": "#",
                "Host_events": "#",
                "Container_events": "#",
            },
        }
