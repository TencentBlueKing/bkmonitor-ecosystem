# Go-日志（SDK）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/sdks/README.md" target="_blank">自定义日志 OpenTelemetry SDK 上报</a>

* <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs（OTel 日志）</a>：OpenTelemetry 中用于描述离散日志事件的信号类型。

* <a href="https://opentelemetry.io/docs/specs/otel/logs/data-model/" target="_blank">Logs Data Model（OTel 日志数据模型）</a>：定义 `resourceLogs`、`scopeLogs`、`logRecords`、`body`、`attributes`、`severityNumber` 等字段含义。

* <a href="https://opentelemetry.io/docs/specs/otlp/#otlphttp" target="_blank">OTLP/HTTP（OpenTelemetry HTTP 上报协议）</a>：定义通过 HTTP 上报 OTel 数据的协议方式，Go SDK 通过 `/v1/logs` 上报日志。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/logs/sdks/go
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/sdks/README.md" target="_blank">自定义日志 OpenTelemetry SDK 上报</a> 创建一个自定义日志数据源，关注创建后提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。

* `API_URL`：完整的日志上报地址。示例地址为「 http://127.0.0.1:4318/v1/logs 」，实际地址请根据页面接入指引填写，并确保运行容器的环境能够访问。

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
// ❗❗【非常重要】填写页面提供的完整 OTLP/HTTP 日志地址，路径以 /v1/logs 结尾。
endpoint := strings.TrimSpace(os.Getenv("API_URL"))
// ❗❗【非常重要】填写当前日志数据源的 Token，Exporter 将其放在 x-bk-token 请求头中。
token := strings.TrimSpace(os.Getenv("TOKEN"))

// 创建 OTLP HTTP Log Exporter，注入认证 Header
exporter, err := otlploghttp.New(ctx,
	otlploghttp.WithEndpointURL(endpoint),
	otlploghttp.WithHeaders(map[string]string{"x-bk-token": token}),
)
if err != nil {
	return fmt.Errorf("创建 OTLP 日志 Exporter: %w", err)
}

// 创建 LoggerProvider，关联资源属性（服务名、环境等）
provider := sdklog.NewLoggerProvider(
	sdklog.WithResource(resource.NewWithAttributes("",
		attribute.String("service.name", "logs-sdk-go"),
		attribute.String("deployment.environment.name", "local"),
	)),
	sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
)

// 将 Go 标准库 slog 桥接到 OTel LoggerProvider
logger := otelslog.NewLogger("logs-sdk-go", otelslog.WithLoggerProvider(provider))
```

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/master/examples/logs/sdks/go" target="_blank">bkmonitor-ecosystem/examples/logs/sdks/go</a> 中找到。

通过 docker build 构建名为 logs-sdk-go 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、API_URL 传递配置参数。下面的尖括号内容是占位符，运行前请替换：

```bash
docker build -t logs-sdk-go .

docker run --rm -e TOKEN="<日志数据源 Token>" \
 -e API_URL="<页面提供的完整 OTLP/HTTP 日志上报地址>" \
 logs-sdk-go
```

运行输出：

```bash
日志已发送，请到对应索引集检索
```

样例上报一条 INFO 级别日志后退出，可在对应索引集中检索 `Go SDK 日志上报示例`。

### 2.4 样例代码

该样例使用 OpenTelemetry Go SDK，通过 OTLP/HTTP 协议上报日志。代码通过 `otelslog` 将 Go 标准库 `slog` 桥接到 OTel LoggerProvider，并在退出前调用 `Shutdown`，等待批处理器发送日志。

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func run(ctx context.Context) error {
	// ❗❗【非常重要】填写页面提供的完整 OTLP/HTTP 日志地址，路径以 /v1/logs 结尾。
	endpoint := strings.TrimSpace(os.Getenv("API_URL"))
	// ❗❗【非常重要】填写当前日志数据源的 Token，Exporter 将其放在 x-bk-token 请求头中。
	token := strings.TrimSpace(os.Getenv("TOKEN"))
	if endpoint == "" || token == "" {
		return fmt.Errorf("请设置 API_URL 和 TOKEN")
	}
	u, err := url.Parse(endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" ||
		u.User != nil || u.RawQuery != "" || u.Fragment != "" || !strings.HasSuffix(u.Path, "/v1/logs") {
		return fmt.Errorf("API_URL 必须是以 /v1/logs 结尾的完整 HTTP(S) 地址，且不能包含查询参数或片段")
	}

	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(endpoint),
		otlploghttp.WithHeaders(map[string]string{"x-bk-token": token}),
	)
	if err != nil {
		return fmt.Errorf("创建 OTLP 日志 Exporter: %w", err)
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource.NewWithAttributes("",
			attribute.String("service.name", "logs-sdk-go"),
			attribute.String("deployment.environment.name", "local"),
		)),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
	logger := otelslog.NewLogger("logs-sdk-go", otelslog.WithLoggerProvider(provider))
	logger.InfoContext(ctx, "Go SDK 日志上报示例", "demo.source", "go-sdk")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := provider.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("发送日志: %w", err)
	}
	return nil
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("日志已发送，请到对应索引集检索")
}
```

## 3. 了解更多

进一步了解以下内容：

* 参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/sdks/README.md" target="_blank">自定义日志 OpenTelemetry SDK 上报</a> 的「查看数据」和「常见问题」部分。

* 了解 <a href="https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog" target="_blank">otelslog 官方 API 文档</a>。

* 了解 <a href="https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp" target="_blank">OTLP/HTTP 日志 Exporter 官方 API 文档</a>。