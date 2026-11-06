# Go-日志（SDK）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="{{docs.logs.learn.sdk_logs_report}}" target="_blank">自定义日志 OpenTelemetry SDK 上报</a>

* <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs（OTel 日志）</a>：OpenTelemetry 中用于描述离散日志事件的信号类型。

* <a href="https://opentelemetry.io/docs/specs/otel/logs/data-model/" target="_blank">Logs Data Model（OTel 日志数据模型）</a>：定义 `resourceLogs`、`scopeLogs`、`logRecords`、`body`、`attributes`、`severityNumber` 等字段含义。

* <a href="https://opentelemetry.io/docs/specs/otlp/#otlphttp" target="_blank">OTLP/HTTP（OpenTelemetry HTTP 上报协议）</a>：定义通过 HTTP 上报 OTel 数据的协议方式，Go SDK 默认通过 `/v1/logs` 上报日志。

### 1.2 接入样例获取

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/logs/sdks/go
```

- *[1] Git 用于获取样例，Docker（或兼容容器工具）用于构建和运行容器样例，仅供快速验证。*
- *[2] 数据上报不依赖 Git 或 Docker，业务可参考后文代码或上报协议，自行组织上报方式。*

## 2. 快速接入

### 2.1 创建应用

参考 <a href="{{docs.logs.learn.sdk_logs_report}}" target="_blank">自定义日志 OpenTelemetry SDK 上报</a> 创建一个自定义日志数据源，关注创建后提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。

* `API_URL`：国内站点默认是「 {{access_config.otlp.http_endpoint}}/v1/logs 」，其他环境、跨云场景请根据页面接入指引填写。

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `API_URL` | 是 | ❗❗【非常重要】日志上报接口地址（`Access URL`），请根据页面接入指引填写；如果页面提供的是 OTLP HTTP Endpoint 根地址，请在末尾追加 `/v1/logs`。 |
| `TOKEN` | 是 | ❗❗【非常重要】日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。 |

#### 2.2.1 关键配置

使用 OpenTelemetry Go SDK 进行日志上报时，需正确配置以下核心参数：

```go
import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// getEnv 读取环境变量，若未设置则返回默认值。
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// newHttpExporter 创建 OTLP HTTP 日志导出器。
// endpointURL 为完整上报地址（例如 http://127.0.0.1:4318/v1/logs），
// SDK 通过 WithEndpointURL 自动解析协议和路径。
func newHttpExporter(endpointURL, token string) (*otlploghttp.Exporter, error) {
	return otlploghttp.New(
		context.Background(),
		otlploghttp.WithEndpointURL(endpointURL),
		otlploghttp.WithHeaders(map[string]string{"X-BK-TOKEN": token}),
	)
}

// ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
token := getEnv("TOKEN", "fixme")
// ❗❗【非常重要】上报地址，国内站点默认是「 {{access_config.otlp.http_endpoint}}/v1/logs 」，
// 其他环境、跨云场景请根据页面接入指引填写。
apiURL := getEnv("API_URL", "{{access_config.otlp.http_endpoint}}/v1/logs")

// 创建 OTLP HTTP Log Exporter，注入认证 Header
exporter, err := newHttpExporter(apiURL, token)
if err != nil {
	// 处理错误
}

// 创建资源属性（服务名、环境等）
res := resource.NewSchemaless(
	attribute.String("service.name", "custom-log-sdk-demo"),
	attribute.String("deployment.environment.name", "local"),
)

// 创建 LoggerProvider，关联资源属性
provider := sdklog.NewLoggerProvider(
	sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	sdklog.WithResource(res),
)
global.SetLoggerProvider(provider)

// 获取 OpenTelemetry Logger
otelLogger := provider.Logger("otel")
```

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/logs/sdks/go" target="_blank">bkmonitor-ecosystem/examples/logs/sdks/go</a> 中找到。

通过 docker build 构建名为 logs-sdk-go 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、API_URL 传递配置参数，实现周期上报日志：

```bash
docker build -t logs-sdk-go .

docker run -e TOKEN="fixme" \
 -e API_URL="{{access_config.otlp.http_endpoint}}/v1/logs" \
 logs-sdk-go
```

运行输出：

```bash
2026/11/06 12:00:00 Starting log reporter (exporter=http) ...
2026/11/06 12:00:01 Sent: 12:00:01.000000 INFO - log from opentelemetry sdk (count=1)
2026/11/06 12:00:02 Sent: 12:00:02.000000 DEBUG - log from opentelemetry sdk (count=2)
2026/11/06 12:00:03 Sent: 12:00:03.000000 WARN - log from opentelemetry sdk (count=3)
2026/11/06 12:00:04 Sent: 12:00:04.000000 ERROR - log from opentelemetry sdk (count=4)
...
```

### 2.4 样例代码

该样例使用 OpenTelemetry Go SDK，通过 OTLP/HTTP 协议上报日志。代码集成 `go.opentelemetry.io/otel` 和 `go.opentelemetry.io/otel/sdk/log`，通过 `SdkLoggerProvider` 对 OTel LogRecord 进行构造与批量上报。

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otelLog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// getEnv 读取环境变量，若未设置则返回默认值。
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// newHttpExporter 创建 OTLP HTTP 日志导出器。
func newHttpExporter(endpointURL, token string) (*otlploghttp.Exporter, error) {
	return otlploghttp.New(
		context.Background(),
		otlploghttp.WithEndpointURL(endpointURL),
		otlploghttp.WithHeaders(map[string]string{"X-BK-TOKEN": token}),
	)
}

// newGrpcExporter 创建 OTLP gRPC 日志导出器。
func newGrpcExporter(endpoint, token string) (*otlploggrpc.Exporter, error) {
	return otlploggrpc.New(
		context.Background(),
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithHeaders(map[string]string{"X-BK-TOKEN": token}),
	)
}

// consoleExporter 实现 sdklog.Exporter 接口，将日志记录输出到控制台。
type consoleExporter struct{}

func (e *consoleExporter) Export(ctx context.Context, records []sdklog.Record) error {
	for _, r := range records {
		timestamp := r.Timestamp().Format("2006/01/02 15:04:05.000000")
		severityText := r.SeverityText()
		body := r.Body().AsString()
		fmt.Printf("%s %s - %s\n", timestamp, severityText, body)
	}
	return nil
}

func (e *consoleExporter) Shutdown(ctx context.Context) error   { return nil }
func (e *consoleExporter) ForceFlush(ctx context.Context) error  { return nil }

// newConsoleExporter 创建仅输出到控制台的导出器。
func newConsoleExporter() sdklog.Exporter {
	return &consoleExporter{}
}

// logLevel 定义日志级别，包含 OTLP Severity 和文本描述。
type logLevel struct {
	severity otelLog.Severity
	text     string
}

// logLevels 预定义 DEBUG/INFO/WARN/ERROR 四种日志级别。
var logLevels = []logLevel{
	{severity: 5, text: "DEBUG"},
	{severity: 9, text: "INFO"},
	{severity: 13, text: "WARN"},
	{severity: 17, text: "ERROR"},
}

func main() {
	// 命令行参数
	output := flag.String("exporter", "http", "output exporter type: stdout/http/grpc")
	flag.Parse()

	// ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
	token := getEnv("TOKEN", "fixme")
	// ❗❗【非常重要】上报地址，国内站点默认是「 {{access_config.otlp.http_endpoint}}/v1/logs 」，
	// 其他环境、跨云场景请根据页面接入指引填写。
	apiURL := getEnv("API_URL", "{{access_config.otlp.http_endpoint}}/v1/logs")

	// 创建导出器
	var exporter sdklog.Exporter
	var err error
	switch *output {
	case "grpc":
		exporter, err = newGrpcExporter(apiURL, token)
	case "http":
		exporter, err = newHttpExporter(apiURL, token)
	default:
		exporter = newConsoleExporter()
	}
	if err != nil {
		log.Fatalf("failed to initialize log exporter: %v", err)
	}

	// 创建资源
	res := resource.NewSchemaless(
		attribute.String("service.name", "custom-log-sdk-demo"),
		attribute.String("deployment.environment.name", "local"),
	)

	// 创建 LoggerProvider
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(provider)

	// 信号处理（支持 SIGINT 和 SIGTERM）
	ctx, cancel := newOSSignalContext()
	defer cancel()

	// 获取 OTel Logger
	logger := provider.Logger("otel")

	// 输出启动信息
	log.Printf("Starting log reporter (exporter=%s) ...\n", *output)

	// 主循环：每秒上报一条随机级别的日志
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		var count int
		for {
			select {
			case <-ticker.C:
				count++
				lvl := logLevels[rng.Intn(len(logLevels))]
				now := time.Now()

				// 构建日志记录
				record := otelLog.Record{}
				record.SetTimestamp(now)
				record.SetObservedTimestamp(now)
				record.SetSeverity(lvl.severity)
				record.SetSeverityText(lvl.text)
				record.SetBody(otelLog.StringValue(fmt.Sprintf("log from opentelemetry sdk (count=%d)", count)))
				record.AddAttributes(
					otelLog.Int("count", count),
					otelLog.String("level", lvl.text),
				)

				// 通过 OpenTelemetry 发送日志
				logger.Emit(ctx, record)

				// 在终端同步打印发送的日志内容
				if *output != "stdout" {
					log.Printf("Sent: %s %s - %s\n", now.Format("2006/01/02 15:04:05.000000"), lvl.text, record.Body().AsString())
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	log.Println("Received shutdown signal, exiting...")

	// 优雅关闭 LoggerProvider
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := provider.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("failed to stop the log provider: %v", err)
	}
}

// newOSSignalContext 创建监听 SIGINT 和 SIGTERM 的上下文，支持优雅退出。
func newOSSignalContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, func() {
		signal.Stop(c)
		cancel()
	}
}
```

## 3. 了解更多

进一步了解以下内容：

* 进行 <a href="{{docs.logs.learn_search}}" target="_blank">日志检索</a>。

* 了解 <a href="{{docs.logs.container_custom_report}}" target="_blank">容器日志自定义上报使用文档</a>。

* 了解 <a href="{{docs.logs.container_collector_install}}" target="_blank">容器日志采集器安装</a>。
