// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 提供了 Prometheus 指标上报 SDK 的示例实现
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
	token = getEnv("TOKEN", "")
	// ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写。
	apiURL   = getEnv("API_URL", "")
	job      = getEnv("JOB", "default_monitor_job") // 任务名称
	instance = getEnv("INSTANCE", "127.0.0.1")      // 实例名称
	port     = getEnv("PORT", "2323")               //  默认2323端口暴露/metrics端点
	interval = getEnvAsInt("INTERVAL", 60)          // 上报间隔，默认60秒

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
		if token != "" {
			return "已配置"
		}
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
