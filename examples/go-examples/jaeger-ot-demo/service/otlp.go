// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package service 提供了 OpenTelemetry 服务的实现
package service

import (
	"context"
	"fmt"
	"log"

	"jaeger-ot-demo/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// newTracerProvider Initializes a new trace provider
func newTracerProvider(res *resource.Resource, exporter *otlptrace.Exporter) *sdktrace.TracerProvider {
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
}

// newPropagator Initializes a new propagator
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// OTelService 定义了 OpenTelemetry 服务的核心结构
type OTelService struct {
	ctx            context.Context
	cancel         context.CancelFunc
	ServiceName    string
	Endpoint       string
	ExporterType   config.ExporterType
	Token          string
	EnableTraces   bool
	gRPCConn       *grpc.ClientConn
	tracerProvider *sdktrace.TracerProvider
}

// Type 返回服务类型标识
func (ots *OTelService) Type() string {
	return "OTelService"
}

// Init 初始化 OTelService 服务
func (ots *OTelService) Init(conf *config.Config, ctx context.Context) error {
	var err error
	ots.ExporterType = conf.OtlpExporterType
	ots.Endpoint = conf.BKEndpoint
	ots.ServiceName = conf.ServiceName
	ots.Token = conf.Token
	ots.EnableTraces = conf.EnableTraces
	ots.ctx = ctx
	if ots.ExporterType == config.ExporterGRPC {
		// ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
		// 格式为 ip:port 或 domain:port，不要带 schema
		ots.gRPCConn, err = grpc.NewClient(ots.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("[%v] failed to create gRPC connection to collector: %v", ots.Type(), err)
			return err
		}
	}

	return nil
}

// Start 启动 OTelService 服务
func (ots *OTelService) Start() error {
	ots.ctx, ots.cancel = context.WithCancel(ots.ctx)

	res, err := ots.newResource()
	if err != nil {
		return err
	}

	if err := ots.setUpTraces(ots.ctx, res); err != nil {
		return err
	}

	return nil
}

// Stop 停止 OTelService 服务并清理资源
func (ots *OTelService) Stop() error {
	defer ots.cancel()

	shutdownFunc := func(provider closer) {
		if err := provider.Shutdown(ots.ctx); err != nil {
			log.Printf("[%v] ignored error during provider shutdown: %v", ots.Type(), err)
		}
	}

	if ots.tracerProvider != nil {
		go shutdownFunc(ots.tracerProvider)
	}

	return nil
}

// setUpTraces
func (ots *OTelService) setUpTraces(ctx context.Context, res *resource.Resource) error {
	if !ots.EnableTraces {
		return nil
	}

	tracerExporter, err := ots.newTracerExporter(ctx)
	if err != nil {
		return err
	}
	ots.tracerProvider = newTracerProvider(res, tracerExporter)
	otel.SetTracerProvider(ots.tracerProvider)
	otel.SetTextMapPropagator(newPropagator())
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Printf("[otel] error: %v", err)
	}))

	return nil
}

func (ots *OTelService) newResource() (*resource.Resource, error) {
	extraRes, err := resource.New(
		ots.ctx,
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			// ❗❗【非常重要】应用服务唯一标识
			semconv.ServiceNameKey.String(ots.ServiceName),
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
func (ots *OTelService) newTracerExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	switch ots.ExporterType {
	case config.ExporterHttp:
		return newHttpTracerExporter(
			ctx,
			// ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
			// 格式为 ip:port 或 domain:port，不要带 schema
			ots.Endpoint,
			// ❗❗【非常重要】请传入应用 Token
			map[string]string{"x-bk-token": ots.Token},
		)
	case config.ExporterGRPC:
		return newGRPCTracerExporter(ctx, ots.gRPCConn, map[string]string{"x-bk-token": ots.Token})
	}
	return nil, fmt.Errorf("[%v] invalid exporter type", ots.Type())
}
