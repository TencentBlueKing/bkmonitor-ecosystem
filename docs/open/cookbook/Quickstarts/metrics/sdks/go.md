# go-指标（Prometheus）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Term/metrics/what.md" target="_blank">什么是指标</a>

* <a href="{{COOKBOOK_METRICS_TYPES}}" target="_blank">指标类型</a>

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/metrics/sdks/go
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/sdks/README.md" target="_blank">自定义指标 Prometheus SDK 上报</a> 创建一个上报协议为 `Prometheus` 的自定义指标，关注创建后提供的配置项：

* `TOKEN`：数据源 Token，后续需要在上报数据时使用。

同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image-1.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数     | 类型                | 描述                         |
| ------------ | ------------------- | ---------------------------- |
|`TOKEN`         |String      | ❗❗【非常重要】 自定义指标数据源 `Token`。 |
|`API_URL`       |String      | ❗❗【非常重要】 数据上报接口地址（`Access URL`），国内站点请填写「 127.0.0.1:4318 」，其他环境、跨云场景请根据页面接入指引填写。|
|`INTERVAL`      |Integer  　 |数据上报间隔，默认值为 60 秒。  ​​ |
|`METRICS_PORT`  |Integer  　 |指标暴露端口，默认 2323。|

#### 2.2.1 关键配置

蓝鲸监控支持原生 Prometheus 协议，如果业务已接入 Prometheus SDK，使用 `push_to_gateway` 方法，修改上报地址为 `API_URL`，增加注入 `X-BK-TOKEN` 的 handler。

采用了 `pusher.Push()` 方法，通过配置好的安全客户端，将 `registry` 中收集的所有指标一次性发送到指定的 Pushgateway。

```go
func pushMetrics() error {
    if apiURL == "" {
        return fmt.Errorf("API_URL未配置")
    }

    // ❗️❗️【非常重要】 创建使用X-BK-TOKEN认证的HTTP客户端
    client := &http.Client{
        Transport: &xbkTokenTransport{token: token},
        Timeout:   30 * time.Second,
    }

    pusher := push.New(apiURL, job).
        Gatherer(registry).
        Grouping("instance", instance).
        Client(client)

    return pusher.Push()
}

// 推送指标
    if err := pushMetrics(); err != nil {
        log.Printf("❌ 推送失败: %v", err)
    } else {
        log.Printf("✅ 推送成功")
    }
```

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/metrics/sdks/go" target="_blank">bkmonitor-ecosystem/examples/metrics/sdks/go</a> 中找到。

PUSH 上报（metric 服务主动上报到端点）：

```bash
docker build -t metrics-sdk-go .

docker run \
-e JOB="default_monitor_job" \
-e INSTANCE="127.0.0.1" \
-e API_URL="127.0.0.1:4318" \
-e TOKEN="xxx" \
-e INTERVAL=60 metrics-sdk-go
```

### 2.4 使用示例

#### 2.4.1 Counter

用于记录累计值（如 API 调用总量、错误次数），只能递增。 可用于统计接口请求量、错误率（结合 rate / increase 等函数计算）。

例如，可以通过以下方式上报请求总数：

```go
import "github.com/prometheus/client_golang/prometheus"

// ===== 计数器相关 =====
// Metrics（指标）- 使用 Counter 类型指标
// Refer：https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Counter
var apiCounter = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "api_calls_total",
        Help: "API调用总次数",
    },
    []string{"api_name", "status_code"},
)

func generateCounterMetrics() {
    methods := []string{"GET", "POST", "PUT", "DELETE"}
    endpoints := []string{"/api/users", "/api/orders", "/api/products"}
    statusCodes := []string{"200", "400", "500"}

    method := methods[rand.Intn(len(methods))]
    endpoint := endpoints[rand.Intn(len(endpoints))]
    status := statusCodes[rand.Intn(len(statusCodes))]

    apiCounter.WithLabelValues(endpoint, status).Inc()
    log.Printf("📊 Counter指标 | %s %s | 状态: %s", method, endpoint, status)
}
```

<a href="https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Counter" target="_blank">Prometheus Golang SDK - Counter</a>。

#### 2.4.2 Gauge

用于记录瞬时值（可任意增减），如实时资源状态、队列长度、活跃连接数等。

```go
// ===== 仪表盘相关 =====
// Metrics（指标）- 使用 Gauge 类型指标
// Refer：https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Gauge
var cpuGauge = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "cpu_usage_percent",
        Help: "CPU使用率百分比",
    },
    []string{"host_name"},
)

func generateGaugeMetrics() {
    hosts := []string{"web-server-01", "db-server-01", "app-server-01"}
    host := hosts[rand.Intn(len(hosts))]
    usage := 10.0 + rand.Float64()*80.0 // 10%-90%之间的随机值

    cpuGauge.WithLabelValues(host).Set(usage)
    log.Printf("📈 Gauge指标 | %s | 使用率: %.1f%%", host, usage)
}
```

<a href="https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Gauge" target="_blank">Prometheus Golang SDK - Gauge</a>。

#### 2.4.3 Histogram

用于记录数值分布情况（如任务耗时、响应大小），通过预定的桶（bucket）统计观测值落入各区间的频率，并自动生成 _sum （总和）、_count （总数）等衍生指标。适用于分析耗时分布、计算分位数（P90/P95）等场景。

```go
// ===== 直方图相关 =====
// Metrics（指标）- 使用 Histogram 类型指标
// Refer: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Histogram
var requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "请求耗时分布",
        Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
    },
    []string{"service"},
)

func generateHistogramMetrics() {
    services := []string{"user-service", "order-service", "payment-service"}
    service := services[rand.Intn(len(services))]
    duration := 0.01 + rand.Float64()*4.99 // 0.01-5.0秒的延迟

    requestDuration.WithLabelValues(service).Observe(duration)
    log.Printf("⏱️  Histogram指标 | %s | 延迟: %.3fs", service, duration)
}
```

<a href="https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Histogram" target="_blank">Prometheus Golang SDK - Histogram</a>。

#### 2.4.4 Summary

用于在客户端直接计算分位数（如 P95/P99 请求耗时），适用于需高精度分位数且无需跨实例聚合场景。

```go
// ===== 摘要相关 =====
// Metrics（指标）- 使用 Summary 类型指标
// Refer: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Summary
var processingTime = prometheus.NewSummaryVec(
    prometheus.SummaryOpts{
        Name:       "data_processing_seconds",
        Help:       "任务处理时间摘要",
        Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
    },
    []string{"operation"},
)

func generateSummaryMetrics() {
    operations := []string{"data_validation", "payment_processing", "email_sending"}
    operation := operations[rand.Intn(len(operations))]
    processTime := 0.005 + rand.Float64()*0.995 // 0.005-1.0秒的处理时间

    processingTime.WithLabelValues(operation).Observe(processTime)
    log.Printf("⚡ Summary指标 | %s | 耗时: %.3fs", operation, processTime)
}
```

<a href="https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Summary" target="_blank">Prometheus Golang SDK - Summary</a>。

### 2.5 样例代码

该样例使用 Prometheus_client 库实现四种指标类型（`Counter`、`Gauge`、`Histogram`、`Summary`）上报：

```go
package main

import (
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/client_golang/prometheus/push"
)

// ==================== 配置信息 ====================
var (
    // ❗️❗️【非常重要】请填写为申请到的自定义指标认证令牌（`Token`）。
    token    = getEnv("TOKEN", "")
    // ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写。
    apiURL   = getEnv("API_URL", "")
    job      = getEnv("JOB", "default_monitor_job")  // 任务名称
    instance = getEnv("INSTANCE", "127.0.0.1")      // 实例名称
    port     = getEnv("PORT", "2323")      //  默认2323端口暴露/metrics端点
    interval = getEnvAsInt("INTERVAL", 60)  // 上报间隔，默认60秒

    registry = prometheus.NewRegistry() // 创建注册表
)

// ==================== 指标类型定义 ====================

// Counter类型 - API调用统计
// Refer：https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Counter
var apiCounter = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "api_calls_total",
        Help: "API调用总次数",
    },
    []string{"api_name", "status_code"},
)

// Gauge类型 - CPU使用率监控
// Refer：https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Gauge
var cpuGauge = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "cpu_usage_percent",
        Help: "CPU使用率百分比",
    },
    []string{"host_name"},
)

// Histogram类型 - 请求耗时分布
// Refer: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Histogram
var requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "请求耗时分布",
        Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
    },
    []string{"service"},
)

// Summary类型 - 处理时间摘要
// Refer: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Summary
var processingTime = prometheus.NewSummaryVec(
    prometheus.SummaryOpts{
        Name:       "data_processing_seconds",
        Help:       "任务处理时间摘要",
        Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
    },
    []string{"operation"},
)

// ==================== 自定义HTTP（X-BK-TOKEN认证） ====================
type xbkTokenTransport struct {
    token string
}

func (t *xbkTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // 克隆请求以避免修改原始请求
    reqClone := req.Clone(req.Context())
    if t.token != "" {
        reqClone.Header.Set("X-BK-TOKEN", t.token) // ❗️❗️【非常重要】注入 `TOKEN`。
    }
    return http.DefaultTransport.RoundTrip(reqClone)
}

// ==================== 环境变量读取函数 ====================
func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value, exists := os.LookupEnv(key); exists {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

// ==================== 指标演示函数 ====================
func generateCounterMetrics() {
    methods := []string{"GET", "POST", "PUT", "DELETE"}
    endpoints := []string{"/api/users", "/api/orders", "/api/products"}
    statusCodes := []string{"200", "400", "500"}

    method := methods[rand.Intn(len(methods))]
    endpoint := endpoints[rand.Intn(len(endpoints))]
    status := statusCodes[rand.Intn(len(statusCodes))]

    apiCounter.WithLabelValues(endpoint, status).Inc()
    log.Printf("📊 Counter指标 | %s %s | 状态: %s", method, endpoint, status)
}

func generateGaugeMetrics() {
    hosts := []string{"web-server-01", "db-server-01", "app-server-01"}
    host := hosts[rand.Intn(len(hosts))]
    usage := 10.0 + rand.Float64()*80.0 // 10%-90%之间的随机值

    cpuGauge.WithLabelValues(host).Set(usage)
    log.Printf("📈 Gauge指标 | %s | 使用率: %.1f%%", host, usage)
}

func generateHistogramMetrics() {
    services := []string{"user-service", "order-service", "payment-service"}
    service := services[rand.Intn(len(services))]
    duration := 0.01 + rand.Float64()*4.99 // 0.01-5.0秒的延迟

    requestDuration.WithLabelValues(service).Observe(duration)
    log.Printf("⏱️  Histogram指标 | %s | 延迟: %.3fs", service, duration)
}

func generateSummaryMetrics() {
    operations := []string{"data_validation", "payment_processing", "email_sending"}
    operation := operations[rand.Intn(len(operations))]
    processTime := 0.005 + rand.Float64()*0.995 // 0.005-1.0秒的处理时间

    processingTime.WithLabelValues(operation).Observe(processTime)
    log.Printf("⚡ Summary指标 | %s | 耗时: %.3fs", operation, processTime)
}

// ==================== 安全的指标推送函数 ====================
func pushMetrics() error {
    if apiURL == "" {
        return fmt.Errorf("API_URL未配置")
    }

    // ❗️❗️【非常重要】 创建使用X-BK-TOKEN认证的HTTP客户端
    client := &http.Client{
        Transport: &xbkTokenTransport{token: token},
        Timeout:   30 * time.Second,
    }

    pusher := push.New(apiURL, job).
        Gatherer(registry).
        Grouping("instance", instance).
        Client(client)

    return pusher.Push()
}

// ==================== 初始化函数 ====================
func init() {
    rand.Seed(time.Now().UnixNano())

    // 注册所有指标到注册表
    registry.MustRegister(apiCounter)
    registry.MustRegister(cpuGauge)
    registry.MustRegister(requestDuration)
    registry.MustRegister(processingTime)
}

// ==================== 主函数 ====================
func main() {
    log.Println("🚀 启动Prometheus指标上报服务")
    log.Printf("🔧 配置信息:")
    log.Printf("  实例: %s", instance)
    log.Printf("  任务: %s", job)
    log.Printf("  目标: %s", apiURL)
    log.Printf("  认证: %s", func() string {
        if token != "" { return "已配置" }
        return "未配置"
    }())
    log.Printf("  间隔: %d秒", interval)
    log.Printf("  端口: %s", port)
    log.Println("")

    // 启动Pull模式HTTP服务器
    go func() {
        http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
        http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`{"status": "healthy"}`))
        })

        addr := ":" + port
        log.Printf("🌐 Pull模式启动: http://0.0.0.0%s/metrics", addr)
        if err := http.ListenAndServe(addr, nil); err != nil {
            log.Printf("⚠️  HTTP服务器启动失败: %v", err)
        }
    }()

    // 主循环 - 指标生成和推送
    ticker := time.NewTicker(time.Duration(interval) * time.Second)
    defer ticker.Stop()

    counter := 0
    for {
        <-ticker.C
        counter++

        log.Printf("\n--- 第%d次上报开始 ---", counter)
        startTime := time.Now()

        // 生成所有指标数据
        generateCounterMetrics()
        generateGaugeMetrics()
        generateHistogramMetrics()
        generateSummaryMetrics()

        // 推送指标
        if err := pushMetrics(); err != nil {
            log.Printf("❌ 推送失败: %v", err)
        } else {
            log.Printf("✅ 推送成功")
        }

        elapsed := time.Since(startTime).Seconds()
        log.Printf("--- 第%d次上报完成 | 耗时: %.2fs ---", counter, elapsed)
    }
}
```

### 2.6 PULL 模式

上文主要介绍将指标数据，**主动推送**到蓝鲸监控平台，也可以通过 HTTP 暴露指标， HTTP 端点暴露指标，供 Prometheus 服务器主动抓取。

样例代码同时兼容 PULL 和 PUSH，通过 `promhttp.HandlerFor` 使用注册表 `registry` 暴露指标：

```go
var port = getEnv("PORT", "2323")   //  默认2323端口暴露/metrics端点

// 启动Pull模式HTTP服务器
go func() {
    http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "healthy"}`))
    })

    addr := ":" + port
    log.Printf("🌐 Pull模式启动: http://0.0.0.0%s/metrics", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Printf("⚠️  HTTP服务器启动失败: %v", err)
    }
}()
...
```

运行样例：

```bash
docker build -t metrics-sdk-go .
docker run -p 2323:2323 --name sdk-pull-go metrics-sdk-go
```

获取指标：

```bash
curl http://127.0.0.1:2323/metrics
```

得到类似输出说明启动成功：

```go
# HELP api_calls_total API调用总次数
# TYPE api_calls_total counter
api_calls_total{api_name="/api/users",status_code="500"} 1
# HELP cpu_usage_percent CPU使用率百分比
# TYPE cpu_usage_percent gauge
cpu_usage_percent{host_name="app-server-01"} 54.14169664418541
# HELP data_processing_seconds 任务处理时间摘要
# TYPE data_processing_seconds summary
data_processing_seconds{operation="email_sending",quantile="0.5"} 0.4457927017752269
data_processing_seconds{operation="email_sending",quantile="0.9"} 0.4457927017752269
data_processing_seconds{operation="email_sending",quantile="0.99"} 0.4457927017752269
data_processing_seconds_sum{operation="email_sending"} 0.4457927017752269
data_processing_seconds_count{operation="email_sending"} 1
# HELP http_request_duration_seconds 请求耗时分布
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{service="payment-service",le="0.05"} 0
http_request_duration_seconds_bucket{service="payment-service",le="0.1"} 0
http_request_duration_seconds_bucket{service="payment-service",le="0.25"} 0
http_request_duration_seconds_bucket{service="payment-service",le="0.5"} 0
http_request_duration_seconds_bucket{service="payment-service",le="1"} 0
http_request_duration_seconds_bucket{service="payment-service",le="2"} 0
http_request_duration_seconds_bucket{service="payment-service",le="5"} 1
http_request_duration_seconds_bucket{service="payment-service",le="+Inf"} 1
http_request_duration_seconds_sum{service="payment-service"} 2.5345757848993262
http_request_duration_seconds_count{service="payment-service"} 1
```

## 3. 了解更多

* 进行 <a href="#" target="_blank">指标检索</a>。

* 了解 <a href="#" target="_blank">怎么使用监控指标</a>。

* 了解如何 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/dashboard.md" target="_blank">配置仪表盘</a>。

* 了解如何使用 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/alarm-configurations/rules.md" target="_blank">监控告警</a>。