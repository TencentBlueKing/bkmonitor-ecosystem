# 服务快速接入指引（Python）

本示例仅演示如何将 <a href="https://opentelemetry.io/docs/concepts/signals/traces/" target="_blank">Traces</a>、<a href="https://opentelemetry.io/docs/concepts/signals/metrics/" target="_blank">Metrics</a>、<a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs</a>、<a href="https://grafana.com/docs/pyroscope/latest/introduction/profiling/" target="_blank">Profiling</a> 四类观测数据接入蓝鲸应用性能监控。

## 1. 环境要求

在开始之前，请确保您已经安装了以下软件：
* Git
* Docker 或者其他平替的容器工具。


## 2. 初始化示例 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/python-examples/helloworld
docker build -t helloworld-python:latest .
```


## 3. 运行示例 demo

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="xxx" \
-e SERVICE_NAME="helloworld" \
-e OTLP_ENDPOINT="http://127.0.0.1:4318" \
-e PROFILING_ENDPOINT="http://127.0.0.1:4318/pyroscope" \
-e ENABLE_TRACES="true" \
-e ENABLE_METRICS="true" \
-e ENABLE_LOGS="true" \
-e ENABLE_PROFILING="true" helloworld-python:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

运行参数说明：

| 参数                 | 值（根据所填写接入信息生成）             | 说明                                                         |
| -------------------- | :--------------------------------------- | ------------------------------------------------------------ |
| `TOKEN`              | `"{{access_config.token}}"`              | 【必须】APM 应用 `Token`。                                     |
| `SERVICE_NAME`       | `"{{service_name}}"`                     | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分。 |
| `OTLP_ENDPOINT`      | `"{{access_config.otlp.http_endpoint}}"` | 【必须】OT 数据上报地址，支持以下协议：<br />  `gRPC`：`{{access_config.otlp.endpoint}}`<br /> `HTTP`：`{{access_config.otlp.http_endpoint}}`（demo 使用该协议演示上报） |
| `PROFILING_ENDPOINT` | `"{{access_config.profiling.endpoint}}"` | 【可选】Profiling 数据上报地址。                               |
| `ENABLE_TRACES`      | `{{access_config.otlp.enable_traces}}`   | 是否启用调用链上报。                                           |
| `ENABLE_METRICS`     | `{{access_config.otlp.enable_metrics}}`  | 是否启用指标上报。                                             |
| `ENABLE_LOGS`        | `{{access_config.otlp.enable_logs}}`     | 是否启用日志上报。                                             |
| `ENABLE_PROFILING`   | `{{access_config.profiling.endpoint}}`   | 是否启用性能分析上报。                                         |

* *<a href="https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/" target="_blank">OTLP Exporter Configuration</a>*