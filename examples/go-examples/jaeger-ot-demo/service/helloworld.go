// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package service 提供了 Jaeger 与 OpenTelemetry 集成示例应用的服务实现
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"jaeger-ot-demo/config"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// HelloWorldService 提供 Hello World 服务的实现
type HelloWorldService struct {
	ctx       context.Context
	cancel    context.CancelFunc
	address   string
	port      int
	countries []string
	server    *http.Server
}

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

// Init 初始化 HelloWorld 服务
func (hws *HelloWorldService) Init(conf *config.Config, ctx context.Context) error {
	hws.ctx, hws.cancel = context.WithCancel(ctx)
	hws.address = conf.ServerAddress
	hws.port = conf.ServerPort
	hws.countries = countries
	// 1. 创建路由器（避免使用全局 DefaultServeMux）
	router := http.NewServeMux()
	router.HandleFunc("/helloworld", hws.helloWorldHandler)

	// 2. 创建并存储 http.Server 实例
	hws.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", hws.address, hws.port),
		Handler: router, // 使用自定义路由器
	}

	// 3. 记录启动日志
	log.Printf("服务启动，监听地址：%s", hws.server.Addr)
	return nil
}

// Start 启动 HelloWorld 服务
func (hws *HelloWorldService) Start() error {
	if err := hws.server.ListenAndServe(); err != nil {
		log.Printf("HTTP 服务器错误: %v", err)
		return err
	}
	log.Printf("[%v] service started", hws.Type())

	return nil
}

// Stop 停止 HelloWorld 服务
func (hws *HelloWorldService) Stop() error {
	if hws.server != nil {
		if err := hws.server.Shutdown(hws.ctx); err != nil {
			return err
		}
	}

	if hws.cancel != nil {
		hws.cancel()
	}

	return nil
}

// Type 返回服务的类型名称
// 该方法用于标识当前服务实例的类型，便于日志记录和服务管理
func (hws *HelloWorldService) Type() string {
	return "HelloWorldService"
}

func (hws *HelloWorldService) helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 从 HTTP 请求中提取追踪上下文
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header),
	)

	// 2. 创建 Span（如果存在上游上下文则继承）
	var span opentracing.Span
	if err == nil {
		span = opentracing.StartSpan("helloWorldHandler", ext.RPCServerOption(wireContext))
	} else {
		span = opentracing.StartSpan("helloWorldHandler")
	}
	defer span.Finish()

	// 3. 将 Span 放入上下文
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	// Traces（调用链）- 自定义 Span
	hws.traces_custom_span_demo(ctx)
	// Traces（调用链）- Span 事件
	hws.traces_span_event_demo(ctx)
	// Traces（调用链）- 模拟错误
	hws.tracesRandomErrorDemo(ctx)

	// 6. 返回响应
	country := hws.choiceCountry()
	log.Println(fmt.Sprintf("get country -> %s", country))
	w.WriteHeader(http.StatusOK)
	greeting := hws.generateGreeting(country)
	w.Write([]byte(greeting))
}

func (hws *HelloWorldService) logsDemo(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "logsDemo")
	defer span.Finish()
	span.LogFields(
		otlog.Int("helloworld.kind", 1),
		otlog.String("helloworld.step", "Hello World Handler logsDemo."),
	)
	log.Println("Hello World Handler logsDemo.")
}

func (hws *HelloWorldService) traces_custom_span_demo(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "traces_custom_span_demo")
	defer span.Finish()
	span.SetTag("helloworld.kind", 2)
	span.SetTag("helloworld.step", "traces_custom_span_demo")
}

func (hws *HelloWorldService) traces_span_event_demo(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "traces_span_event_demo")
	defer span.Finish()
	span.LogKV("helloworld.kind", 3)
	span.LogKV("helloworld.step", "traces_span_event_demo")
}

func (hws *HelloWorldService) tracesRandomErrorDemo(ctx context.Context) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "tracesRandomErrorDemo")
	defer span.Finish()
	if err := randErr(1); err != nil {
		ext.Error.Set(span, true)
		span.LogFields(
			otlog.String("event", "error"),
			otlog.String("error.message", err.Error()),
			otlog.String("error.type", fmt.Sprintf("%T", err)),
		)
		log.Printf(fmt.Sprintf("[tracesRandomErrorDemo] got error -> %v", err))
		return err
	}
	return nil
}

func (hws *HelloWorldService) choiceCountry() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(hws.countries))
	return countries[randomIndex]
}

func (hws *HelloWorldService) generateGreeting(country string) string {
	return fmt.Sprintf("Hello World, %s!", country)
}

func randErr(errRate float64) error {
	if rand.Float64() < errRate {
		return choiceErr()
	}
	return nil
}

func choiceErr() error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(customErrors))
	return customErrors[randomIndex]
}
