# Profiling-Go（Pyroscope SDK）接入

本指南将帮助您使用 Pyroscope SDK 接入蓝鲸应用性能监控，以 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/common/examples/helloworld.md" target="_blank">入门项目-HelloWorld</a> 为例，介绍性能分析数据接入及 SDK 使用场景。

入门项目功能齐全且可在开发环境运行，可以通过该项目快速接入并体验蓝鲸应用性能监控相关功能。

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
* Go 1.21 或更高版本

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/go-examples/helloworld
docker build -t helloworld-go:latest .
```

## 2. 快速体验

### 2.1 运行样例

运行参数说明：

| 参数                   | 推荐值                                | 说明                                        |
|----------------------|------------------------------------|-------------------------------------------|
| `TOKEN`              | `""`                               | APM 应用 `Token`。                            |
| `PROFILING_ENDPOINT` | `"http://127.0.0.1:4318/pyroscope"` | Profiling 数据上报地址，请根据页面指引提供的 HTTP 接入地址进行填写。<br />推荐值为「国内站点」，其他环境、跨云场景请根据页面服务接入指引填写。 |
| `SERVICE_NAME`       | `"helloworld"`                     | 服务唯一标识，一个应用可以有多个服务，通过该属性区分。                |
| `ENABLE_PROFILING`   | `false`                            | 是否启用性能分析上报。                                |

💡 为保证数据能上报到平台，`TOKEN`、`PROFILING_ENDPOINT` 请务必根据应用接入指引提供的实际值填写。

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="xxx" \
-e SERVICE_NAME="helloworld" \
-e PROFILING_ENDPOINT="http://127.0.0.1:4318/pyroscope" \
-e ENABLE_PROFILING="true" helloworld-go:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

### 2.2 查看数据

等待片刻，便可在「服务详情-Profiling」看到应用数据。

![](image/img.png)

## 3. 快速接入

### 3.1 Pyroscope SDK

<a href="https://grafana.com/docs/pyroscope/latest/" target="_blank">Pyroscope</a> 是 Grafana 旗下用于聚合连续分析数据的开源软件项目。

请在创建 `PyroscopeConfig` 时，准确传入以下信息：

| 属性                | 说明                                            |
|-------------------|-----------------------------------------------|
| `AuthToken`       | 【必须】APM 应用 `Token`                            |
| `ApplicationName` | 【必须】服务唯一标识，一个应用可以有多个服务，通过该属性区分                |
| `ServerAddress`   | 【必须】Profiling 数据上报地址，请根据页面指引提供的 HTTP 接入地址进行填写 |

在项目中引入模块依赖：

```shell
go get github.com/grafana/pyroscope-go@v1.1.2
```

示例项目提供集成 Pyroscope Go SDK 并将性能数据发送到 bk-collector 的方式，可以参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/helloworld/service/profiling/profiling.go" target="_blank">service/profiling/profiling.go</a> 进行接入:

```go
profiler, _ = pyroscope.Start(
    pyroscope.Config{
        //❗❗【非常重要】请传入应用 Token
        AuthToken: s.config.Token,
        //❗❗【非常重要】应用服务唯一标识
        ApplicationName: s.config.ServiceName,
        //❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
        ServerAddress: s.config.Addr,
        Logger:        pyroscope.StandardLogger,
        ProfileTypes: []pyroscope.ProfileType{
            pyroscope.ProfileCPU
        }
    }
)
```

参考官方文档以获得更多信息：<a href="https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/go_push/" target="_blank">Configure the client to send profiles - Go</a>

### 3.2 关联 Traces 数据

Pyroscope 支持同 OpenTelemetry 集成，将 Traces 和 Profiling 数据链接起来，从而实现分析具体跨度（Span）的资源使用情况的目的。

在开始之前，可以阅读文档 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/helloworld/README.md" target="_blank">Go（OpenTelemetry SDK）接入</a> 了解 OpenTelemetry。

在项目中引入依赖：

```shell
# Make sure you also upgrade pyroscope server to version 0.14.0 or higher.
go get github.com/grafana/otel-profiling-go@v0.5.1
```

示例项目在 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/go-examples/helloworld/service/otlp/otlp.go" target="_blank">service/otlp/otlp.go setUpTraces</a> 提供了创建样例：

```go
// setUpTraces
func (s *Service) setUpTraces(ctx context.Context, res *resource.Resource) error {
	if !s.config.EnableTraces {
		return nil
	}

	tracerExporter, err := s.newTracerExporter(ctx)
	if err != nil {
		return err
	}
	s.tracerProvider = newTracerProvider(res, tracerExporter)
	s.wg.Add(1)

	if s.config.EnableProfiling {
		// 关键代码，注入 otelpyroscope TracerProvider
		otel.SetTracerProvider(otelpyroscope.NewTracerProvider(s.tracerProvider))
	} else {
		otel.SetTracerProvider(s.tracerProvider)
	}

	otel.SetTextMapPropagator(newPropagator())
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Printf("[otel] error: %v", err)
	}))

	return nil
}
```

如果是<a href="" target="_blank">Go（tRPC 云观 Oteam SDK）接入</a>，可在 `pyroscope.Start` 位置添加如下代码（需要 SDK 版本在 `0.6.3` 及以上）

```go
// import 部分
// "go.opentelemetry.io/otel"
// otelpyroscope "github.com/grafana/otel-profiling-go"
// oteltrpctrace "opentelemetry/opentelemetry-go-ecosystem/instrumentation/oteltrpc/traces"

tracerProvider := otel.GetTracerProvider()
otel.SetTracerProvider(otelpyroscope.NewTracerProvider(tracerProvider))
oteltrpctrace.SetDefaultTracer(otel.Tracer(""))

// profiler, _ := pyroscope.Start(
// 	pyroscope.Config{
// 		... ...
// 	})
```

参考官方文档以获得更多信息：<a href="https://grafana.com/docs/pyroscope/latest/configure-client/trace-span-profiles/go-span-profiles/" target="_blank">Span profiles with Traces to profiles for Go</a>

## 4. 了解更多

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem" target="_blank">各语言、框架接入代码样例</a>