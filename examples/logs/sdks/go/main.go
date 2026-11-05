// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 提供了上报 OTLP 日志数据到蓝鲸监控平台的 SDK 示例应用。
package main

import (
	"context"
	"flag"
	"fmt"
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

// getEnv 读取环境变量
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// 日志级别定义
type logLevel struct {
	number otelLog.Severity
	text   string
}

var logLevels = []logLevel{
	{5, "DEBUG"},
	{9, "INFO"},
	{13, "WARN"},
	{17, "ERROR"},
}

// newHttpExporter 创建 OTLP HTTP 导出器
func newHttpExporter(endpoint, token string, insecure bool) (*otlploghttp.Exporter, error) {
	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithHeaders(map[string]string{"X-BK-TOKEN": token}),
	}
	if insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}
	return otlploghttp.New(context.Background(), opts...)
}

// newGrpcExporter 创建 OTLP gRPC 导出器
func newGrpcExporter(endpoint, token string, insecure bool) (*otlploggrpc.Exporter, error) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithHeaders(map[string]string{"X-BK-TOKEN": token}),
	}
	if insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	return otlploggrpc.New(context.Background(), opts...)
}

// consoleExporter 自定义控制台导出器，输出纯文本格式
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

func (e *consoleExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (e *consoleExporter) ForceFlush(ctx context.Context) error {
	return nil
}

// newConsoleExporter 创建控制台导出器
func newConsoleExporter() sdklog.Exporter {
	return &consoleExporter{}
}

func main() {
	// 命令行参数
	output := flag.String("exporter", "stdout", "output exporter type: stdout/http/grpc")
	interval := flag.Duration("interval", 1*time.Second, "log sending interval")
	insecure := flag.Bool("insecure", false, "use insecure connection (only for http/grpc)")
	flag.Parse()

	// ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
	token := getEnv("TOKEN", "fixme")
	// ❗❗【非常重要】上报地址，国内站点默认是「 {{access_config.otlp.http_endpoint}}/v1/logs 」，
// 其他环境、跨云场景请根据页面接入指引填写
	apiURL := getEnv("API_URL", "fixme")

	// 创建导出器
	var exporter sdklog.Exporter
	var err error
	switch *output {
	case "grpc":
		exporter, err = newGrpcExporter(apiURL, token, *insecure)
	case "http":
		exporter, err = newHttpExporter(apiURL, token, *insecure)
	default:
		exporter = newConsoleExporter()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize log exporter: %v\n", err)
		os.Exit(1)
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

	// 获取 OTel Logger
	logger := provider.Logger("otel")

	// 信号处理
	ctx, cancel := newOSSignalContext()
	defer cancel()

	// 输出启动信息
	startTime := time.Now().Format("2006/01/02 15:04:05.000000")
	fmt.Printf("%s Starting log reporter (exporter=%s) ...\n", startTime, *output)

	// 主循环
	go func() {
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()

		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		var count int
		for {
			select {
			case <-ticker.C:
				count++
				// 随机选择日志级别
				lvl := logLevels[rng.Intn(len(logLevels))]

				// 构建日志记录
				now := time.Now()
				record := otelLog.Record{}
				record.SetTimestamp(now)
				record.SetObservedTimestamp(now)
				record.SetSeverity(lvl.number)
				record.SetSeverityText(lvl.text)
				record.SetBody(otelLog.StringValue(fmt.Sprintf("log from opentelemetry sdk (count=%d)", count)))
				record.AddAttributes(
					otelLog.Int("count", count),
					otelLog.String("level", lvl.text),
				)

				// 发送日志
				logger.Emit(ctx, record)
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	fmt.Println("Received shutdown signal, exiting...")

	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := provider.Shutdown(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to stop the log provider: %v\n", err)
		os.Exit(1)
	}
}

// newOSSignalContext 创建监听 SIGINT 和 SIGTERM 的上下文
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
