# Go-日志（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/http/README.md" target="_blank">自定义日志 HTTP 上报</a>

* <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs（OTel 日志）</a>：OpenTelemetry 中用于描述离散日志事件的信号类型。

* <a href="https://opentelemetry.io/docs/specs/otel/logs/data-model/" target="_blank">Logs Data Model（OTel 日志数据模型）</a>：定义 `resourceLogs`、`scopeLogs`、`logRecords`、`body`、`attributes`、`severityNumber` 等字段含义。

* <a href="https://opentelemetry.io/docs/specs/otlp/#otlphttp" target="_blank">OTLP/HTTP（OpenTelemetry HTTP 上报协议）</a>：定义通过 HTTP 上报 OTel 数据的协议方式，本示例使用 `/v1/logs` 上报日志。

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/logs/http/go
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/http/README.md" target="_blank">自定义日志 HTTP 上报</a> 创建一个上报协议为 `JSON` 的自定义日志，关注创建后提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。

* `API_URL`：国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，其他环境、跨云场景请根据页面接入指引填写。

### 2.2 样例运行参数

运行参数说明：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `API_URL` | 是 | ❗❗【非常重要】日志上报接口地址（`Access URL`），请根据页面接入指引填写；如果页面提供的是 OTLP HTTP Endpoint 根地址，请在末尾追加 `/v1/logs`。 |
| `TOKEN` | 是 | ❗❗【非常重要】日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。 |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/master/examples/logs/http/go" target="_blank">bkmonitor-ecosystem/examples/logs/http/go</a> 中找到。

通过 docker build 构建名为 logs-http-go 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、API_URL 传递配置参数，实现周期上报日志：

```bash
docker build -t logs-http-go .

docker run -e TOKEN="fixme" \
 -e API_URL="http://127.0.0.1:4318/v1/logs" \
 logs-http-go
```

运行输出：

```bash
2026/07/29 02:53:11 Starting log reporter (press Ctrl+C to stop)
2026/07/29 02:53:11 Sending log level: WARN (13)
2026/07/29 02:53:11 response.status_code=200, body={}
2026/07/29 02:53:11 Sending log level: DEBUG (5)
2026/07/29 02:53:11 response.status_code=200, body={}
2026/07/29 02:53:11 Sending log level: INFO (9)
2026/07/29 02:53:11 response.status_code=200, body={}
2026/07/29 02:53:12 Sending log level: ERROR (17)
2026/07/29 02:53:12 response.status_code=200, body={}
...
```

### 2.4 样例代码

上报代码示例：

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// ==================== 环境变量辅助函数 ====================

// getEnv 从环境变量读取值，若为空则返回默认值
func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// ==================== 默认配置 ====================
var (
	// ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
	token = getEnv("TOKEN", "fixme")
	// ❗❗【非常重要】上报地址，国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，
	// 其他环境、跨云场景请根据页面接入指引填写
	apiURL = getEnv("API_URL", "http://127.0.0.1:4318/v1/logs")
	// 上报间隔（秒）
	interval = 5
)

// ==================== 数据类型定义 ====================

// LogLevel 定义日志级别结构，包含 OTLP 协议中的严重程度编号和文本描述
type LogLevel struct {
	SeverityNumber int    `json:"severityNumber"`
	SeverityText   string `json:"severityText"`
	Message        string `json:"message"`
}

// logLevels 预定义的日志级别列表，覆盖 DEBUG/INFO/WARN/ERROR 四种级别
var logLevels = []LogLevel{
	{SeverityNumber: 5, SeverityText: "DEBUG", Message: "debug log from go http"},
	{SeverityNumber: 9, SeverityText: "INFO", Message: "info log from go http"},
	{SeverityNumber: 13, SeverityText: "WARN", Message: "warn log from go http"},
	{SeverityNumber: 17, SeverityText: "ERROR", Message: "error log from go http"},
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// ==================== 辅助函数 ====================

// getCurrentNanoTimestamp 返回当前 UTC 时间的纳秒级 Unix 时间戳字符串
func getCurrentNanoTimestamp() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getRandomLevel 随机返回一个日志级别
func getRandomLevel() LogLevel {
	return logLevels[rng.Intn(len(logLevels))]
}

// ==================== OTLP 日志上报逻辑 ====================

// buildPayload 构建 OTLP 格式的日志上报负载，包含随机选择的日志级别
func buildPayload() map[string]interface{} {
	currentNano := getCurrentNanoTimestamp()
	level := getRandomLevel()
	return map[string]interface{}{
		"resourceLogs": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"attributes": []interface{}{
						map[string]interface{}{
							"key": "service.name",
							"value": map[string]interface{}{
								"stringValue": "custom-log-demo",
							},
						},
						map[string]interface{}{
							"key": "deployment.environment.name",
							"value": map[string]interface{}{
								"stringValue": "local",
							},
						},
					},
				},
				"scopeLogs": []interface{}{
					map[string]interface{}{
						"scope": map[string]interface{}{
							"name": "java-http-demo",
						},
						"logRecords": []interface{}{
							map[string]interface{}{
								"timeUnixNano":         currentNano,
								"observedTimeUnixNano": currentNano,
								"severityNumber":       level.SeverityNumber,
								"severityText":         level.SeverityText,
								"body": map[string]interface{}{
									"stringValue": level.Message,
								},
								"attributes": []interface{}{
									map[string]interface{}{
										"key": "demo.source",
										"value": map[string]interface{}{
											"stringValue": "golang",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// doPost 发送 HTTP POST 请求上报日志
func doPost(client *http.Client, payload map[string]interface{}) {
	// 提取日志记录信息
	resourceLogs := payload["resourceLogs"].([]interface{})
	firstResourceLog := resourceLogs[0].(map[string]interface{})
	scopeLogs := firstResourceLog["scopeLogs"].([]interface{})
	firstScopeLog := scopeLogs[0].(map[string]interface{})
	logRecords := firstScopeLog["logRecords"].([]interface{})
	logRecord := logRecords[0].(map[string]interface{})

	log.Printf("Sending log level: %s (%v)",
		logRecord["severityText"], logRecord["severityNumber"])

	jsonData, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", apiURL,
		bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-bk-token", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to post request: %v", err)
		return
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	log.Printf("response.status_code=%d, body=%s",
		resp.StatusCode, buf.String())
}

// ==================== 主函数 ====================
func main() {
	client := &http.Client{Timeout: 10 * time.Second}
	log.Println("Starting log reporter (press Ctrl+C to stop)...")

	// 捕获 SIGINT（Ctrl+C）和 SIGTERM 实现优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{}, 1)
	go func() {
		<-sigCh
		log.Println("Received signal, exiting...")
		close(done)
	}()

	for {
		select {
		case <-done:
			return
		default:
			payload := buildPayload()
			doPost(client, payload)
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
}
```

## 3. 了解更多

进一步了解以下内容：

* 进行 <a href="#" target="_blank">日志检索</a>。

* 了解 <a href="#" target="_blank">容器日志自定义上报使用文档</a>。

* 了解 <a href="#" target="_blank">容器日志采集器安装</a>。