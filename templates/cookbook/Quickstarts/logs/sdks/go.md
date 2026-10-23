# Go-日志（SDK）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="./README.md" target="_blank">自定义日志 OpenTelemetry SDK 上报</a>：介绍日志数据源的创建流程和查看数据入口。

* <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs（OTel 日志）</a>：OpenTelemetry 中记录离散事件的信号类型。

* <a href="https://opentelemetry.io/docs/specs/otel/logs/data-model/" target="_blank">Logs Data Model（OTel 日志数据模型）</a>：定义日志记录、资源属性等字段的含义。

* <a href="https://opentelemetry.io/docs/specs/otlp/#otlphttp" target="_blank">OTLP/HTTP</a>：通过 HTTP 传输 OpenTelemetry 数据的协议，本样例使用 `/v1/logs` 上报日志。

### 1.2 开发环境要求

在开始之前，请确保已经安装：

* Git。

* Docker 或兼容的容器工具。Dockerfile 使用 Go 1.26 构建，无需在宿主机安装 Go；直接运行 `go run .` 则需要 Go 1.26 或更新版本。

### 1.3 初始化 demo

克隆仓库并进入样例目录：

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/logs/sdks/go
```

命令用法见 <a href="https://git-scm.com/docs/git-clone" target="_blank">Git clone 官方文档</a>。

## 2. 快速接入

### 2.1 创建应用

参考 <a href="./README.md" target="_blank">自定义日志 OpenTelemetry SDK 上报</a> 创建上报协议为 `OTLP/HTTP (Protobuf)` 的日志数据源，并记录页面提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时通过 `x-bk-token` 请求头传递。

* `API_URL`：完整的日志上报地址，包含 `http://` 或 `https://`，路径以 `/v1/logs` 结尾。如果页面只提供 OTLP/HTTP Endpoint 根地址，请追加 `/v1/logs`。

确认运行样例的主机能够访问上报地址。Token 应通过环境变量传入，不要写入源码或提交到仓库。

### 2.2 样例运行参数

运行参数说明：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `API_URL` | 是 | ❗❗【非常重要】当前数据源的完整 OTLP/HTTP 日志上报地址；如果页面只给出 Endpoint 根地址，请追加 `/v1/logs`。 |
| `TOKEN` | 是 | ❗❗【非常重要】当前数据源的 Token，通过 `x-bk-token` 请求头发送。 |

请使用接收 OTLP/HTTP Protobuf 日志的地址，不能使用仅接收普通 JSON 日志的接口。

#### 2.2.1 关键配置

样例从环境变量读取并校验 `API_URL` 和 `TOKEN`。创建 Exporter 时，使用完整上报地址和当前数据源的 Token：

```go
exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(endpoint),
		otlploghttp.WithHeaders(map[string]string{"x-bk-token": token}),
	)
```

关于配置方法，请参考 <a href="https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp" target="_blank">OTLP/HTTP 日志 Exporter 官方 API 文档</a>。完整的配置校验及错误处理见下文样例代码。

### 2.3 运行样例

完整样例位于 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/logs/sdks/go" target="_blank">bkmonitor-ecosystem/examples/logs/sdks/go</a>。构建 Docker 镜像，并使用页面提供的完整上报地址和 Token 运行容器；下面的尖括号内容是占位符，运行前请替换：

```bash
docker build -t logs-sdk-go .

docker run --rm \
  -e API_URL="<页面提供的完整 OTLP/HTTP 日志上报地址>" \
  -e TOKEN="<日志数据源 Token>" \
  logs-sdk-go
```

命令用法见 <a href="https://docs.docker.com/reference/cli/docker/buildx/build/" target="_blank">Docker build 官方文档</a>和 <a href="https://docs.docker.com/reference/cli/docker/container/run/" target="_blank">Docker run 官方文档</a>。

样例上报一条 INFO 级别日志后退出。SDK 未报告发送错误时，终端输出：

```text
日志已发送，请到对应索引集检索
```

这不代表日志已经入库。进入「日志采集 → 采集详情 → 查看数据」，在当前数据源对应的索引集中检索 `Go SDK 日志上报示例`。入库和索引刷新可能需要一段时间。如果没有查到日志，请确认地址、Token、网络、索引集和检索时间范围；终端出现 `发送日志` 错误时，先检查地址、鉴权和网络配置。

也可以在安装 Go 1.26 或更新版本的宿主机上设置相同的环境变量，进入样例目录执行 `go run .`。

### 2.4 样例代码

样例先校验配置，再创建 OTLP/HTTP Exporter 和 `LoggerProvider`。`otelslog` 将 Go 标准库 `slog` 的消息和属性转换为 OpenTelemetry 日志记录。退出前调用 `Shutdown`，等待批处理器发送日志：

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

`service.name` 和 `deployment.environment.name` 描述产生日志的服务；`demo.source` 描述本条日志事件。关于桥接器和 Exporter 的接口，请参考 <a href="https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog" target="_blank">otelslog 官方 API 文档</a>和 <a href="https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp" target="_blank">OTLP/HTTP 日志 Exporter 官方 API 文档</a>。

## 3. 了解更多

* 参考 <a href="./README.md" target="_blank">自定义日志 OpenTelemetry SDK 上报</a> 的「查看数据」和「常见问题」部分，了解索引集选择与排查步骤。

* 关于 Go 日志桥接器和 Exporter，请参考 <a href="https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog" target="_blank">otelslog 官方 API 文档</a>和 <a href="https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp" target="_blank">OTLP/HTTP 日志 Exporter 官方 API 文档</a>。
