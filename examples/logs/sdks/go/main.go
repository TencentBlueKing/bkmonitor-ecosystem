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
