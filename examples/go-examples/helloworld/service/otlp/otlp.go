// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package otlp 提供了 OpenTelemetry Protocol (OTLP) 的集成服务实现
package otlp

import (
	"context"
	"fmt"
	"log"
	"sync"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/config"
	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/define"
)

type closer interface {
	Shutdown(context.Context) error
}

// newHttpTracerExporter Initialize a new HTTP tracer exporter.
func newHttpTracerExporter(ctx context.Context, endpoint string, headers map[string]string) (*otlptrace.Exporter, error) {
	return otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithHeaders(headers),
	)
}

// newGRPCTracerExporter Initialize a new gRPC tracer exporter.
func newGRPCTracerExporter(ctx context.Context, conn *grpc.ClientConn, headers map[string]string) (*otlptrace.Exporter, error) {
	return otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithGRPCConn(conn),
		otlptracegrpc.WithHeaders(headers),
	)
}

// newHttpMeterExporter Initialize a new HTTP meter exporter.
func newHttpMeterExporter(ctx context.Context, endpoint string, headers map[string]string) (*otlpmetrichttp.Exporter, error) {
	return otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithHeaders(headers),
	)
}

// newGRPCMeterExporter Initialize a new gRPC meter exporter.
func newGRPCMeterExporter(ctx context.Context, conn *grpc.ClientConn, headers map[string]string) (*otlpmetricgrpc.Exporter, error) {
	return otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithGRPCConn(conn),
		otlpmetricgrpc.WithHeaders(headers),
	)
}

// newHttpLoggerExporter Initialize a new HTTP log exporter.
func newHttpLoggerExporter(ctx context.Context, endpoint string, headers map[string]string) (*otlploghttp.Exporter, error) {
	return otlploghttp.New(
		ctx,
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithHeaders(headers),
	)
}

// newGRPCLoggerExporter Initialize a new gRPC log exporter.
func newGRPCLoggerExporter(ctx context.Context, conn *grpc.ClientConn, headers map[string]string) (*otlploggrpc.Exporter, error) {
	return otlploggrpc.New(
		ctx,
		otlploggrpc.WithGRPCConn(conn),
		otlploggrpc.WithHeaders(headers),
	)
}

// newTracerProvider Initializes a new trace provider
func newTracerProvider(res *resource.Resource, exporter *otlptrace.Exporter) *sdktrace.TracerProvider {
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
}

// newMeterProvider Initializes a new meter Provider
func newMeterProvider(res *resource.Resource, exporter sdkmetric.Exporter) *sdkmetric.MeterProvider {
	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(res),
	)
}

// newLoggerProvider Initializes a new log Provider
func newLoggerProvider(res *resource.Resource, exporter sdklog.Exporter) *sdklog.LoggerProvider {
	return sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
}

// newPropagator Initializes a new propagator
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// Service 定义了 OTLP 服务的核心结构
type Service struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	config         *Config
	gRPCConn       *grpc.ClientConn
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
}

// Type 返回服务类型标识
func (s *Service) Type() string {
	return "otlp"
}

// Init 初始化 OTLP 服务
func (s *Service) Init(conf *config.Config) error {
	s.config = &Config{
		Token:           conf.Token,
		ServiceName:     conf.ServiceName,
		Endpoint:        conf.OtlpEndpoint,
		ExporterType:    conf.OtlpExporterType,
		EnableTraces:    conf.EnableTraces,
		EnableMetrics:   conf.EnableMetrics,
		EnableLogs:      conf.EnableLogs,
		EnableProfiling: conf.EnableProfiling,
	}

	var err error
	if s.config.ExporterType == define.ExporterGRPC {
		// ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
		// 格式为 ip:port 或 domain:port，不要带 schema
		s.gRPCConn, err = grpc.NewClient(s.config.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("[%v] failed to create gRPC connection to collector: %v", s.Type(), err)
			return err
		}
	}

	return nil
}

// Start 启动 OTLP 服务并初始化各项功能
func (s *Service) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	res, err := s.newResource()
	if err != nil {
		return err
	}

	if err := s.setUpTraces(s.ctx, res); err != nil {
		return err
	}

	if err := s.setUpMetrics(s.ctx, res); err != nil {
		return err
	}

	if err := s.setUpLogs(s.ctx, res); err != nil {
		return err
	}

	return nil
}

// Stop 停止 OTLP 服务并清理资源
func (s *Service) Stop() error {
	defer s.cancel()

	shutdownFunc := func(provider closer) {
		defer s.wg.Done()
		if err := provider.Shutdown(s.ctx); err != nil {
			log.Printf("[%v] ignored error during provider shutdown: %v", s.Type(), err)
		}
	}

	if s.tracerProvider != nil {
		go shutdownFunc(s.tracerProvider)
	}
	if s.meterProvider != nil {
		go shutdownFunc(s.meterProvider)
	}
	if s.loggerProvider != nil {
		go shutdownFunc(s.loggerProvider)
	}

	s.wg.Wait()

	return nil
}

// setUpTraces
func (s *Service) setUpTraces(ctx context.Context, res *resource.Resource) error {
	if !s.config.EnableTraces {
		return nil
	}

	tracerExporter, err := s.newTracerExporter(ctx)
	if err != nil {
		return err
	}
	s.tracerProvider = newTracerProvider(res, tracerExporter)
	s.wg.Add(1)

	if s.config.EnableProfiling {
		// 关键代码，注入 otelpyroscope TracerProvider
		otel.SetTracerProvider(otelpyroscope.NewTracerProvider(s.tracerProvider))
	} else {
		otel.SetTracerProvider(s.tracerProvider)
	}

	otel.SetTextMapPropagator(newPropagator())
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Printf("[otel] error: %v", err)
	}))

	return nil
}

// setUpMetrics
func (s *Service) setUpMetrics(ctx context.Context, res *resource.Resource) error {
	if !s.config.EnableMetrics {
		return nil
	}

	meterExporter, err := s.newMeterExporter(ctx)
	if err != nil {
		return err
	}
	s.wg.Add(1)
	s.meterProvider = newMeterProvider(res, meterExporter)
	otel.SetMeterProvider(s.meterProvider)

	return nil
}

// setUpLogs
func (s *Service) setUpLogs(ctx context.Context, res *resource.Resource) error {
	if !s.config.EnableLogs {
		return nil
	}

	logExporter, err := s.newLoggerExporter(ctx)
	if err != nil {
		return err
	}
	s.wg.Add(1)
	s.loggerProvider = newLoggerProvider(res, logExporter)
	global.SetLoggerProvider(s.loggerProvider)

	return nil
}

func (s *Service) newResource() (*resource.Resource, error) {
	extraRes, err := resource.New(
		s.ctx,
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			// ❗❗【非常重要】应用服务唯一标识
			semconv.ServiceNameKey.String(s.config.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// resource.Default() 提供了部分 SDK 默认属性
	res, err := resource.Merge(resource.Default(), extraRes)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// newTracerExporter Initialize a new tracer exporter based on ExporterType
func (s *Service) newTracerExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	switch s.config.ExporterType {
	case define.ExporterHttp:
		return newHttpTracerExporter(
			ctx,
			// ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
			// 格式为 ip:port 或 domain:port，不要带 schema
			s.config.Endpoint,
			// ❗❗【非常重要】请传入应用 Token
			map[string]string{"x-bk-token": s.config.Token},
		)
	case define.ExporterGRPC:
		return newGRPCTracerExporter(ctx, s.gRPCConn, map[string]string{"x-bk-token": s.config.Token})
	}
	return nil, fmt.Errorf("[%v] invalid exporter type", s.Type())
}

// newMeterExporter Initialize a new meter exporter based on ExporterType
func (s *Service) newMeterExporter(ctx context.Context) (sdkmetric.Exporter, error) {
	switch s.config.ExporterType {
	case define.ExporterHttp:
		return newHttpMeterExporter(
			ctx,
			// ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
			// 格式为 ip:port 或 domain:port，不要带 schema
			s.config.Endpoint,
			// ❗❗【非常重要】请传入应用 Token
			map[string]string{"x-bk-token": s.config.Token},
		)
	case define.ExporterGRPC:
		return newGRPCMeterExporter(ctx, s.gRPCConn, map[string]string{"x-bk-token": s.config.Token})
	}
	return nil, fmt.Errorf("[%v] invalid exporter type", s.Type())
}

// newLoggerExporter Initialize a new log exporter based on ExporterType
func (s *Service) newLoggerExporter(ctx context.Context) (sdklog.Exporter, error) {
	switch s.config.ExporterType {
	case define.ExporterHttp:
		return newHttpLoggerExporter(
			ctx,
			// ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
			// 格式为 ip:port 或 domain:port，不要带 schema
			s.config.Endpoint,
			// ❗❗【非常重要】请传入应用 Token
			map[string]string{"x-bk-token": s.config.Token},
		)
	case define.ExporterGRPC:
		return newGRPCLoggerExporter(ctx, s.gRPCConn, map[string]string{"x-bk-token": s.config.Token})
	}
	return nil, fmt.Errorf("[%v] invalid exporter type", s.Type())
}
