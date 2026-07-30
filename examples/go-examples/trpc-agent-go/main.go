//
// Tencent is pleased to support the open source community by making trpc-agent-go available.
//
// Copyright (C) 2025 Tencent.  All rights reserved.
//
// trpc-agent-go is licensed under the Apache License Version 2.0.
//

// Package main 提供一个可自行产生观测流量的 tRPC-Agent 最小接入示例。
package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"trpc.group/trpc-go/trpc-agent-go/agent/llmagent"
	"trpc.group/trpc-go/trpc-agent-go/log"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/model/openai"
	"trpc.group/trpc-go/trpc-agent-go/runner"
	trpcagentrunner "trpc.group/trpc-go/trpc-agent-go/runner/trpcagent"
	servertrpcagent "trpc.group/trpc-go/trpc-agent-go/server/trpcagent"
	ametric "trpc.group/trpc-go/trpc-agent-go/telemetry/metric"
	atrace "trpc.group/trpc-go/trpc-agent-go/telemetry/trace"
	"trpc.group/trpc-go/trpc-agent-go/tool"
	"trpc.group/trpc-go/trpc-agent-go/tool/function"
	"trpc.group/trpc-go/trpc-go"
	thttp "trpc.group/trpc-go/trpc-go/http"
)

const (
	serviceName    = "trpc.test.trpcagent.api"
	appName        = "calculator"
	basePath       = "/trpc-agent/v1/apps"
	target         = "http://127.0.0.1:8080"
	modelName      = "deepseek-v4-flash"
	queryInterval  = 3 * time.Second
	requestTimeout = 2 * time.Minute
)

func main() {
	ctx := context.Background()
	shutdownTelemetry, err := setupTelemetry(ctx)
	if err != nil {
		log.Fatalf("failed to set up telemetry: %v", err)
	}
	defer func() {
		if err := shutdownTelemetry(); err != nil {
			log.Errorf("failed to shut down telemetry: %v", err)
		}
	}()

	agent := newAgent()
	agentRunner := runner.NewRunner(appName, agent)
	defer agentRunner.Close()
	trpcAgentServer, err := servertrpcagent.New(
		servertrpcagent.WithAppName(appName),
		servertrpcagent.WithAgent(agent),
		servertrpcagent.WithRunner(agentRunner),
		servertrpcagent.WithBasePath(basePath),
		servertrpcagent.WithTimeout(requestTimeout),
	)
	if err != nil {
		log.Fatalf("failed to create tRPC-Agent API server: %v", err)
	}
	server := trpc.NewServer()
	// 将 tRPC-Agent 提供的标准 HTTP Handler 挂载到 trpc-go 服务。
	thttp.RegisterNoProtocolServiceMux(server.Service(serviceName), trpcAgentServer.Handler())
	log.Infof("tRPC-Agent API: serving app %q on service %q%s/%s", appName, serviceName, basePath, appName)
	go loopQuery()
	if err := server.Serve(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}

func setupTelemetry(ctx context.Context) (func() error, error) {
	endpoint := strings.TrimSpace(os.Getenv("OTLP_ENDPOINT"))
	if endpoint == "" {
		return nil, errors.New("OTLP_ENDPOINT is required")
	}
	token := strings.TrimSpace(os.Getenv("TOKEN"))
	if token == "" {
		return nil, errors.New("TOKEN is required")
	}
	otelServiceName := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if otelServiceName == "" {
		otelServiceName = serviceName
	}

	// ❗❗【非常重要】请将 APM 应用 Token 作为 x-bk-token Header 传入。
	metricHeaders := "x-bk-token=" + url.QueryEscape(token)
	if err := os.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", metricHeaders); err != nil {
		return nil, fmt.Errorf("set metrics headers: %w", err)
	}
	meterProvider, err := ametric.NewMeterProvider(
		ctx,
		// ❗❗【非常重要】请填写 APM 接入指引提供的 HTTP OTLP 地址，格式为 host:port。
		ametric.WithEndpoint(endpoint),
		ametric.WithProtocol("http"),
		ametric.WithServiceName(otelServiceName),
	)
	if err != nil {
		return nil, fmt.Errorf("create meter provider: %w", err)
	}
	if err := ametric.InitMeterProvider(meterProvider); err != nil {
		shutdownErr := meterProvider.Shutdown(ctx)
		return nil, errors.Join(fmt.Errorf("initialize meter provider: %w", err), shutdownErr)
	}

	shutdownTrace, err := atrace.Start(
		ctx,
		// ❗❗【非常重要】请填写 APM 接入指引提供的 HTTP OTLP 地址，格式为 host:port。
		atrace.WithEndpoint(endpoint),
		atrace.WithProtocol("http"),
		// ❗❗【非常重要】请将 APM 应用 Token 作为 x-bk-token Header 传入。
		atrace.WithHeaders(map[string]string{"x-bk-token": token}),
		atrace.WithServiceName(otelServiceName),
	)
	if err != nil {
		shutdownErr := meterProvider.Shutdown(ctx)
		return nil, errors.Join(fmt.Errorf("start trace provider: %w", err), shutdownErr)
	}

	return func() error {
		return errors.Join(
			shutdownTrace(),
			meterProvider.Shutdown(context.Background()),
		)
	}, nil
}

func loopQuery() {
	// runner 在整个循环中复用，避免每次查询都重新创建 HTTP 客户端。
	r, err := trpcagentrunner.New(
		appName,
		trpcagentrunner.WithTarget(target),
		trpcagentrunner.WithBasePath(basePath),
	)
	if err != nil {
		log.Errorf("failed to create tRPC-Agent runner: %v", err)
		return
	}
	defer r.Close()

	ticker := time.NewTicker(queryInterval)
	defer ticker.Stop()

	for range ticker.C {
		// 异步执行，避免单次模型请求阻塞后续 tick。
		go query(r)
	}
}

func query(r runner.Runner) {
	events, err := r.Run(
		context.Background(),
		"demo-user",
		"demo-session",
		model.NewUserMessage("Use the multiply tool to compute 12 * 7."),
	)
	if err != nil {
		log.Errorf("loop query failed: %v", err)
		return
	}
	// 必须消费完整个事件通道，确保远端 Run 正常结束并释放相关资源。
	for evt := range events {
		if evt != nil && evt.Response != nil && evt.Error != nil {
			log.Errorf("loop query event failed: %s", evt.Error.Message)
			return
		}
	}
	log.Infof("loop query succeeded")
}

func newAgent() *llmagent.LLMAgent {
	// 保留一个简单工具，用于展示完整的 LLM -> Tool 调用链路。
	multiplyTool := function.NewFunctionTool(
		multiply,
		function.WithName("multiply"),
		function.WithDescription("Multiplies two numbers."),
	)
	return llmagent.New(
		"calculator-agent",
		llmagent.WithModel(openai.New(modelName)),
		llmagent.WithInstruction("Use the multiply tool for multiplication."),
		llmagent.WithTools([]tool.Tool{multiplyTool}),
	)
}

func multiply(_ context.Context, args multiplyArgs) (multiplyResult, error) {
	return multiplyResult{Result: args.A * args.B}, nil
}

type multiplyArgs struct {
	A float64 `json:"a" description:"First number."`
	B float64 `json:"b" description:"Second number."`
}

type multiplyResult struct {
	Result float64 `json:"result"`
}
