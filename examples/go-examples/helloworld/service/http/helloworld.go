// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package http 提供了 HTTP 服务的 HelloWorld 接口实现
package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	Name = "helloworld"
)

var (
	tracer = otel.Tracer(Name)
	meter  = otel.Meter(Name)
	logger = otelslog.NewLogger(Name)
)

var countries = []string{
	"United States", "Canada", "United Kingdom", "Germany", "France", "Japan", "Australia", "China", "India", "Brazil",
}

var (
	errMySQLConnectTimeout = errors.New("mysql connect timeout")
	errUserNotFound        = errors.New("user not found")
	errNetworkUnreachable  = errors.New("network unreachable")
	errFileNotFound        = errors.New("file not found")

	customErrors = []error{
		errMySQLConnectTimeout, errUserNotFound, errNetworkUnreachable, errFileNotFound,
	}
)

var (
	requestsTotal              metric.Int64Counter
	taskExecuteDurationSeconds metric.Float64Histogram
	rpcClientHandledTotal      metric.Int64Counter
	rpcServerHandledTotal      metric.Int64Counter
	rpcClientHandledSeconds    metric.Float64Histogram
	rpcServerHandledSeconds    metric.Float64Histogram
)

func init() {
	var err error
	requestsTotal, err = meter.Int64Counter("requests_total", metric.WithDescription("Total number of HTTP requests"))
	if err != nil {
		panic(err)
	}
	taskExecuteDurationSeconds, err = meter.Float64Histogram(
		"task_execute_duration_seconds",
		metric.WithDescription("Task execute duration in seconds"),
		metric.WithExplicitBucketBoundaries(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0),
	)
	if err != nil {
		panic(err)
	}

	// Metrics（指标） - Gauge 类型
	if err = metricsGaugeDemo(); err != nil {
		panic(err)
	}

	rpcClientHandledTotal, err = meter.Int64Counter(
		"rpc_client_handled_total",
		metric.WithDescription("Total number of RPC client handled"),
	)
	if err != nil {
		panic(err)
	}
	rpcServerHandledTotal, err = meter.Int64Counter(
		"rpc_server_handled_total",
		metric.WithDescription("Total number of RPC server handled"),
	)
	if err != nil {
		panic(err)
	}

	rpcClientHandledSeconds, err = meter.Float64Histogram(
		"rpc_client_handled_seconds",
		metric.WithDescription("RPC client handled duration in seconds"),
		metric.WithExplicitBucketBoundaries(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0),
	)
	if err != nil {
		panic(err)
	}

	rpcServerHandledSeconds, err = meter.Float64Histogram(
		"rpc_server_handled_seconds",
		metric.WithDescription("RPC server handled duration in seconds"),
		metric.WithExplicitBucketBoundaries(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0),
	)
	if err != nil {
		panic(err)
	}
}

func queryHelloWorld(ctx context.Context, url string) {
	// 可选：定义 SpanKind
	// 部分 SpanKind 将影响 APM 服务分类：
	// SpanKindConsumer -> 后台任务。
	// SpanKindProducer -> 消息队列（且存在 attributes.messaging.system
	// refer：https://opentelemetry.io/docs/specs/semconv/messaging/messaging-spans/）。
	ctx, span := tracer.Start(ctx, "Caller/queryHelloWorld", trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	logger.InfoContext(ctx, fmt.Sprintf("[queryHelloWorld] send request"))
	res, err := client.Do(req)
	if err != nil {
		logger.ErrorContext(ctx, fmt.Sprintf("[queryHelloWorld] got error -> %v", err))
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.ErrorContext(ctx, fmt.Sprintf("[queryHelloWorld] got error -> %v", err))
		return
	}
	logger.InfoContext(ctx, fmt.Sprintf("[queryHelloWorld] received: %s", body))
}

// LoopQueryHelloWorld 定期循环调用 HelloWorld 服务
func LoopQueryHelloWorld(ctx context.Context, url string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			queryHelloWorld(ctx, url)
		case <-ctx.Done():
			return
		}
	}
}

// HelloWorld 处理 HTTP 请求并返回问候语
func HelloWorld(w http.ResponseWriter, req *http.Request) {
	ctx, span := tracer.Start(req.Context(), "Handle/HelloWorld")
	defer span.End()

	// Logs（日志）
	logsDemo(ctx, req)

	country := choiceCountry()
	logger.InfoContext(ctx, fmt.Sprintf("get country -> %s", country))

	// Metrics（指标） - Counter 类型
	metricsCounterDemo(ctx, country)
	// Metrics（指标） - Histograms 类型
	metricsHistogramDemo(ctx)
	// Metrics（指标） - 调用分析场景
	metricsRpcDemo(ctx, "server")
	metricsRpcDemo(ctx, "client")

	// Traces（调用链）- 自定义 Span
	tracesCustomSpanDemo(ctx)
	// Traces（调用链）- 在当前 Span 上设置自定义属性
	tracesSetCustomSpanAttributes(ctx)
	// Traces（调用链）- Span 事件
	tracesSpanEventDemo(ctx)
	// Traces（调用链）- 模拟错误
	if err := tracesRandomErrorDemo(ctx, span); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	greeting := generateGreeting(country)
	w.Write([]byte(greeting))
}

func doSomething(maxMs int) {
	r := 10 + rand.Intn(maxMs)
	time.Sleep(time.Duration(r) * time.Millisecond)
}

func choiceCountry() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(countries))
	return countries[randomIndex]
}

func choiceErr() error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(customErrors))
	return customErrors[randomIndex]
}

func generateGreeting(country string) string {
	return fmt.Sprintf("Hello World, %s!", country)
}

func randErr(errRate float64) error {
	if rand.Float64() < errRate {
		return choiceErr()
	}
	return nil
}

// logsDemo Logs（日志）打印日志
func logsDemo(ctx context.Context, req *http.Request) {
	// 上报日志
	logger.InfoContext(ctx, fmt.Sprintf("received request: %s %s", req.Method, req.URL))

	// 添加自定义属性
	attrs := []slog.Attr{
		slog.String("method", req.Method), slog.String("k1", "v1"), slog.Int("k2", 123),
	}
	logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		fmt.Sprintf("report log with attrs, received request: %s %s", req.Method, req.URL),
		attrs...,
	)
}

// metricsCounterDemo Metrics（指标）- 使用 Counter 类型指标
// Refer: https://opentelemetry.io/docs/languages/go/instrumentation/#using-counters
func metricsCounterDemo(ctx context.Context, country string) {
	requestsTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("country", country)))
}

// metricsGaugeDemo Metrics（指标）- 使用 Gauge 类型指标
// Refer:https://opentelemetry.io/docs/languages/go/instrumentation/#using-observable-async-gauges
func metricsGaugeDemo() error {
	memoryUsage, err := meter.Float64ObservableGauge("memory_usage", metric.WithDescription("Memory usage"))
	if err != nil {
		return err
	}

	_, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		o.ObserveFloat64(memoryUsage, 0.1+rng.Float64()*0.2)
		return nil
	}, memoryUsage)
	if err != nil {
		return err
	}

	return nil
}

// metricsHistogramDemo Metrics（指标）- 使用 Histogram 类型指标
// Refer: https://opentelemetry.io/docs/languages/go/instrumentation/#using-histograms
func metricsHistogramDemo(ctx context.Context) {
	begin := time.Now()
	doSomething(100)
	cost := time.Since(begin)
	taskExecuteDurationSeconds.Record(ctx, cost.Seconds())
}

// metricsRpcDemo Metrics（指标）- 调用分析场景
// 基于该指标规范上报，可在 APM 服务使用「调用分析」功能，省去自行配置仪表盘、告警等工作。
// 本样例更多演示如何定义、上报调用分析指标，实际使用时，可在客户端调用前、服务端处理请求前后进行埋点，以得到真实的调用数据。
func metricsRpcDemo(ctx context.Context, role string) {
	begin := time.Now()
	doSomething(100)
	cost := time.Since(begin)

	attrs := []attribute.KeyValue{
		attribute.String("rpc_system", "custom"),                        // RPC 系统，支持自定义。
		attribute.String("scope_name", fmt.Sprintf("%s_metrics", role)), // 指标分组，server_metrics/client_metrics。
		attribute.String("instance", "127.0.0.1"),                       // 实例，部署 IP 地址。
		attribute.String("namespace", "Development"),                    // 环境类型，支持自定义，e.g. Production/Development/..。
		attribute.String("env_name", "dev"),                             // 环境名称，支持自定义。
		attribute.String("caller_server", "helloworld"),                 // 主调服务。
		attribute.String("caller_service", "helloworld.timer"),          // 主调 Service，如果不区分服务/Service，可与 caller_server 保持一致。
		attribute.String("caller_method", "loopQueryHelloWorld"),        // 主调接口。
		attribute.String("callee_server", "helloworld"),                 // 被调服务。
		attribute.String("callee_service", "helloworld.http"),           // 被调 Service，如果不区分服务/Service，可与 callee_server 保持一致。
		attribute.String("callee_method", "/helloworld"),                // 被调接口。
		attribute.String("code", "200"),                                 // 返回码，支持自定义。
		attribute.String("code_type", "success"),                        // 返回码类型，可选：success / timeout / exception。
	}

	if role == "client" {
		rpcClientHandledTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
		rpcClientHandledSeconds.Record(ctx, cost.Seconds(), metric.WithAttributes(attrs...))
	} else {
		rpcServerHandledTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
		rpcServerHandledSeconds.Record(ctx, cost.Seconds(), metric.WithAttributes(attrs...))
	}
}

// tracesCustomSpanDemo Traces（调用链）- 增加自定义 Span
// Refer: https://opentelemetry.io/docs/languages/go/instrumentation/#creating-spans
func tracesCustomSpanDemo(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "CustomSpanDemo/doSomething")
	defer span.End()

	// 增加 Span 自定义属性
	// Refer: https://opentelemetry.io/docs/languages/go/instrumentation/#span-attributes
	span.SetAttributes(
		attribute.Int("helloworld.kind", 1),
		attribute.String("helloworld.step", "tracesCustomSpanDemo"),
	)

	doSomething(50)
}

// setCustomSpanAttributes Traces（调用链）- 在当前 Span 上设置自定义属性
func tracesSetCustomSpanAttributes(ctx context.Context) {
	currentSpan := trace.SpanFromContext(ctx)
	currentSpan.SetAttributes(attribute.String("ApiName", "ApiRequest"), attribute.Int("actId", 12345))
}

// tracesSpanEventDemo Traces（调用链）- Span 事件
// Refer: https://opentelemetry.io/docs/languages/go/instrumentation/#events
func tracesSpanEventDemo(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "SpanEventDemo/doSomething")
	defer span.End()

	opt := trace.WithAttributes(
		attribute.Key("helloworld.kind").Int(2),
		attribute.Key("helloworld.step").String("tracesSpanEventDemo"),
	)

	span.AddEvent("Before doSomething", opt)
	doSomething(50)
	span.AddEvent("After doSomething", opt)
}

// tracesRandomErrorDemo Traces（调用链）- 异常事件、状态
// Refer: https://opentelemetry.io/docs/languages/go/instrumentation/#record-errors
func tracesRandomErrorDemo(ctx context.Context, span trace.Span) error {
	if err := randErr(0.1); err != nil {
		logger.ErrorContext(ctx, fmt.Sprintf("[tracesRandomErrorDemo] got error -> %v", err))
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}
