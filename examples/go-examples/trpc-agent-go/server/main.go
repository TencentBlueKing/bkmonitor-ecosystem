// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2026 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 启动一个 A2A Server，将 trpc-agent-go 的 Agent 暴露为可远程调用的服务，
// 并把 Agent 执行产生的 Traces、Metrics 和 Logs 上报到蓝鲸 APM。
//
// 该示例演示的是 tRPC-Agent 框架自身的可观测：Agent / LLM 调用 / Tool 执行的调用链和
// GenAI 指标由框架内置埋点自动产生，只要在 server 端装好 OTel Provider 即可自动上报，
// 业务代码无需手动埋点。因为 Agent 在 server 进程内执行，这些信号也都在 server 端产生。
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	otellog "go.opentelemetry.io/otel/log"
	"trpc.group/trpc-go/trpc-agent-go/agent/llmagent"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/model/openai"
	a2aserver "trpc.group/trpc-go/trpc-agent-go/server/a2a"
	"trpc.group/trpc-go/trpc-agent-go/tool"
	"trpc.group/trpc-go/trpc-agent-go/tool/function"
)

const (
	appName   = "trpc-agent-go-apm-demo"
	agentName = "calculator-agent"
)

// applicationLogger 通过 OTel Bridge 将结构化日志上报到 APM。
var applicationLogger *slog.Logger

type config struct {
	host             string
	token            string
	otlpEndpoint     string
	serviceName      string
	serviceNamespace string
	serviceVersion   string
	modelAPIKey      string
	modelBaseURL     string
	modelName        string
}

func main() {
	if err := run(context.Background(), loadConfig()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cfg config) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	shutdownTelemetry, err := setupTelemetry(ctx, cfg)
	if err != nil {
		return fmt.Errorf("setup telemetry: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			log.Printf("shutdown telemetry: %v", err)
		}
	}()

	agent := newAgent(cfg)

	// ❗❗【关键点】WithAgent 时 A2A Server 会在内部为该 Agent 构建 Runner 并在 server 进程执行，
	// 因此 Agent / LLM / Tool 的内置 Span 与 GenAI 指标都在 server 端产生并自动上报。
	srv, err := a2aserver.New(
		a2aserver.WithHost(cfg.host),
		a2aserver.WithAgent(agent, true),
	)
	if err != nil {
		return fmt.Errorf("create a2a server: %w", err)
	}

	applicationLogger.LogAttrs(
		ctx,
		slog.LevelInfo,
		"A2A server starting",
		slog.String("event.name", "a2a.server.starting"),
		slog.String("agent.name", agentName),
		slog.String("host", cfg.host),
	)
	log.Printf("A2A server listening on http://%s (agent card: http://%s/.well-known/agent.json)", cfg.host, cfg.host)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start(cfg.host)
	}()

	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("a2a server stopped: %w", err)
		}
		return nil
	case <-signalCtx.Done():
		log.Println("shutting down A2A server...")
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Stop(stopCtx); err != nil {
			return fmt.Errorf("stop a2a server: %w", err)
		}
		return nil
	}
}

func newAgent(cfg config) *llmagent.LLMAgent {
	modelOptions := []openai.Option{openai.WithAPIKey(cfg.modelAPIKey)}
	if cfg.modelBaseURL != "" {
		modelOptions = append(modelOptions, openai.WithBaseURL(cfg.modelBaseURL))
	}
	modelInstance := openai.New(cfg.modelName, modelOptions...)

	calculatorTool := function.NewFunctionTool(
		calculator,
		function.WithName("calculator"),
		function.WithDescription("Performs add, subtract, multiply, divide, and power operations."),
	)
	generationConfig := model.GenerationConfig{
		Stream: true,
	}
	return llmagent.New(
		agentName,
		llmagent.WithModel(modelInstance),
		llmagent.WithDescription("A minimal calculator assistant used to verify BlueKing APM telemetry."),
		llmagent.WithInstruction("You are a concise assistant. Use the calculator tool when arithmetic is needed. "+
			"Never reveal credentials or secrets."),
		llmagent.WithGenerationConfig(generationConfig),
		llmagent.WithTools([]tool.Tool{calculatorTool}),
	)
}

func calculator(_ context.Context, args calculatorArgs) (calculatorResult, error) {
	var result float64
	switch args.Operation {
	case "add", "+":
		result = args.A + args.B
	case "subtract", "-":
		result = args.A - args.B
	case "multiply", "*":
		result = args.A * args.B
	case "divide", "/":
		if args.B == 0 {
			return calculatorResult{}, fmt.Errorf("division by zero")
		}
		result = args.A / args.B
	case "power", "^":
		result = math.Pow(args.A, args.B)
	default:
		return calculatorResult{}, fmt.Errorf("unsupported operation %q", args.Operation)
	}
	return calculatorResult{Result: result}, nil
}

type calculatorArgs struct {
	Operation string  `json:"operation" description:"Operation to apply: add, subtract, multiply, divide, or power."`
	A         float64 `json:"a" description:"First operand."`
	B         float64 `json:"b" description:"Second operand."`
}

type calculatorResult struct {
	Result float64 `json:"result"`
}

func newApplicationLogger(loggerProvider otellog.LoggerProvider) *slog.Logger {
	return otelslog.NewLogger(appName, otelslog.WithLoggerProvider(loggerProvider))
}

func loadConfig() config {
	return config{
		host:             envOrDefault("HOST", "127.0.0.1:8080"),
		token:            strings.TrimSpace(os.Getenv("TOKEN")),
		otlpEndpoint:     strings.TrimSpace(os.Getenv("OTLP_ENDPOINT")),
		serviceName:      envOrDefault("SERVICE_NAME", appName),
		serviceNamespace: envOrDefault("SERVICE_NAMESPACE", "trpc-agent"),
		serviceVersion:   envOrDefault("SERVICE_VERSION", "1.0.0"),
		modelAPIKey:      strings.TrimSpace(os.Getenv("MODEL_API_KEY")),
		modelBaseURL:     strings.TrimSpace(os.Getenv("MODEL_BASE_URL")),
		modelName:        envOrDefault("MODEL_NAME", "gpt-4o-mini"),
	}
}

func (cfg config) validate() error {
	missing := make([]string, 0, 3)
	for key, value := range map[string]string{
		"TOKEN":         cfg.token,
		"OTLP_ENDPOINT": cfg.otlpEndpoint,
		"MODEL_API_KEY": cfg.modelAPIKey,
	} {
		if value == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	if strings.Contains(cfg.otlpEndpoint, "://") {
		return fmt.Errorf("OTLP_ENDPOINT must be host:port without a URL scheme for the Go demo")
	}
	return nil
}
