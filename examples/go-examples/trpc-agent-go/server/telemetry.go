// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2026 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	ametric "trpc.group/trpc-go/trpc-agent-go/telemetry/metric"
	atrace "trpc.group/trpc-go/trpc-agent-go/telemetry/trace"
)

// setupTelemetry 在创建 Agent 之前安装全局 Trace、Metric 和 Log Provider。
//
// ❗❗【关键点】trpc-agent-go 的内置埋点（Runner / LlmAgent / model / tool 的 Span 与 GenAI 指标）
// 使用的是 OTel 全局 Provider：atrace.Start 会 otel.SetTracerProvider，ametric.InitMeterProvider
// 会注入框架全局 Meter。因此这里装好 Provider 之后，Agent 的调用链和 GenAI 指标便会自动上报，
// 业务代码无需手动埋点。由于 Agent 在 server 进程内执行，这些信号也在 server 端产生。
func setupTelemetry(ctx context.Context, cfg config) (func(context.Context) error, error) {
	// ❗❗【非常重要】指标导出器通过标准环境变量读取蓝鲸 APM 应用 Token。
	header := "x-bk-token=" + url.QueryEscape(cfg.token)
	if err := os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", header); err != nil {
		return nil, fmt.Errorf("set OTLP headers: %w", err)
	}
	if err := os.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", header); err != nil {
		return nil, fmt.Errorf("set metrics OTLP headers: %w", err)
	}
	if os.Getenv("OTEL_METRIC_EXPORT_INTERVAL") == "" {
		if err := os.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "30000"); err != nil {
			return nil, fmt.Errorf("set metric export interval: %w", err)
		}
	}

	resourceAttributes := k8sResourceAttributes()
	metricOptions := []ametric.Option{
		// ❗❗【非常重要】请使用 APM 接入指引提供的 HTTP OTLP 地址，格式为 host:port。
		ametric.WithEndpoint(cfg.otlpEndpoint),
		ametric.WithProtocol("http"),
		ametric.WithServiceName(cfg.serviceName),
		ametric.WithServiceNamespace(cfg.serviceNamespace),
		ametric.WithServiceVersion(cfg.serviceVersion),
	}
	if len(resourceAttributes) > 0 {
		metricOptions = append(metricOptions, ametric.WithResourceAttributes(resourceAttributes...))
	}
	meterProvider, err := ametric.NewMeterProvider(ctx, metricOptions...)
	if err != nil {
		return nil, fmt.Errorf("create meter provider: %w", err)
	}
	if err := ametric.InitMeterProvider(meterProvider); err != nil {
		_ = meterProvider.Shutdown(ctx)
		return nil, fmt.Errorf("initialize agent metrics: %w", err)
	}

	traceOptions := []atrace.Option{
		// ❗❗【非常重要】请使用 APM 接入指引提供的 HTTP OTLP 地址，格式为 host:port。
		atrace.WithEndpoint(cfg.otlpEndpoint),
		atrace.WithProtocol("http"),
		// ❗❗【非常重要】请将 APM 应用 Token 作为 x-bk-token Header 传入。
		atrace.WithHeaders(map[string]string{"x-bk-token": cfg.token}),
		atrace.WithServiceName(cfg.serviceName),
		atrace.WithServiceNamespace(cfg.serviceNamespace),
		atrace.WithServiceVersion(cfg.serviceVersion),
	}
	if len(resourceAttributes) > 0 {
		traceOptions = append(traceOptions, atrace.WithResourceAttributes(resourceAttributes...))
	}
	traceLifecycleCtx, cancelTraceLifecycle := context.WithCancel(ctx)
	shutdownTrace, err := atrace.Start(traceLifecycleCtx, traceOptions...)
	if err != nil {
		cancelTraceLifecycle()
		_ = meterProvider.Shutdown(ctx)
		return nil, fmt.Errorf("start agent traces: %w", err)
	}
	loggerProvider, err := newLoggerProvider(ctx, cfg, resourceAttributes)
	if err != nil {
		cancelTraceLifecycle()
		_ = shutdownTrace()
		_ = meterProvider.Shutdown(ctx)
		return nil, fmt.Errorf("create logger provider: %w", err)
	}
	global.SetLoggerProvider(loggerProvider)
	applicationLogger = newApplicationLogger(loggerProvider)

	return func(shutdownCtx context.Context) error {
		return shutdownProviders(
			shutdownCtx,
			cancelTraceLifecycle,
			shutdownTrace,
			meterProvider.Shutdown,
			loggerProvider.Shutdown,
		)
	}, nil
}

func newLoggerProvider(
	ctx context.Context,
	cfg config,
	resourceAttributes []attribute.KeyValue,
) (*sdklog.LoggerProvider, error) {
	// ❗❗【非常重要】请使用 APM 接入指引提供的 HTTP OTLP 地址和应用 Token。
	exporter, err := otlploghttp.New(
		ctx,
		otlploghttp.WithEndpoint(cfg.otlpEndpoint),
		otlploghttp.WithInsecure(),
		otlploghttp.WithHeaders(map[string]string{"x-bk-token": cfg.token}),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP log exporter: %w", err)
	}
	attributes := []attribute.KeyValue{
		semconv.ServiceName(cfg.serviceName),
		semconv.ServiceNamespace(cfg.serviceNamespace),
		semconv.ServiceVersion(cfg.serviceVersion),
	}
	attributes = append(attributes, resourceAttributes...)
	res, err := resource.New(
		ctx,
		resource.WithAttributes(attributes...),
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		_ = exporter.Shutdown(ctx)
		return nil, fmt.Errorf("create log resource: %w", err)
	}
	return sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	), nil
}

func shutdownProviders(
	ctx context.Context,
	cancelTrace context.CancelFunc,
	shutdownTrace func() error,
	shutdownMeter func(context.Context) error,
	shutdownLogger func(context.Context) error,
) error {
	type shutdownResult struct {
		provider string
		err      error
	}

	results := make(chan shutdownResult, 3)
	go func() {
		results <- shutdownResult{provider: "trace", err: shutdownTrace()}
	}()
	go func() {
		results <- shutdownResult{provider: "metric", err: shutdownMeter(ctx)}
	}()
	go func() {
		results <- shutdownResult{provider: "log", err: shutdownLogger(ctx)}
	}()

	var shutdownErr error
	deadline := ctx.Done()
	for completed := 0; completed < 3; {
		select {
		case result := <-results:
			completed++
			if result.err != nil {
				shutdownErr = errors.Join(
					shutdownErr,
					fmt.Errorf("shutdown %s provider: %w", result.provider, result.err),
				)
			}
		case <-deadline:
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("telemetry shutdown deadline: %w", ctx.Err()))
			cancelTrace()
			deadline = nil
		}
	}
	cancelTrace()
	return shutdownErr
}

func k8sResourceAttributes() []attribute.KeyValue {
	attributes := make([]attribute.KeyValue, 0, 3)
	for key, value := range map[string]string{
		"k8s.bcs.cluster.id": os.Getenv("K8S_BCS_CLUSTER_ID"),
		"k8s.namespace.name": os.Getenv("K8S_NAMESPACE"),
		"k8s.pod.name":       os.Getenv("K8S_POD_NAME"),
	} {
		if value != "" {
			attributes = append(attributes, attribute.String(key, value))
		}
	}
	return attributes
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes":
		return true
	case "0", "false", "no":
		return false
	default:
		return fallback
	}
}
