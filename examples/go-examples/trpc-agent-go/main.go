// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2026 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 在同一进程中启动 A2A Server 和周期调用它的 loopQuery Client，
// 两者分别运行在 goroutine 中，并把 Agent 执行产生的 Traces 和 Metrics 上报到蓝鲸 APM。
//
// 该示例演示的是 tRPC-Agent 框架自身的可观测：Agent / LLM 调用 / Tool 执行的调用链和
// GenAI 指标由框架内置埋点自动产生，只要在进程启动时装好 OTel Provider 即可自动上报，
// 业务代码无需手动埋点。
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	"trpc.group/trpc-go/trpc-agent-go/agent/llmagent"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/model/openai"
	a2aserver "trpc.group/trpc-go/trpc-agent-go/server/a2a"
	ametric "trpc.group/trpc-go/trpc-agent-go/telemetry/metric"
	atrace "trpc.group/trpc-go/trpc-agent-go/telemetry/trace"
	"trpc.group/trpc-go/trpc-agent-go/tool"
	"trpc.group/trpc-go/trpc-agent-go/tool/function"
)

const (
	appName   = "trpc-agent-go-apm-demo"
	agentName = "calculator-agent"
)

type config struct {
	host         string
	token        string
	otlpEndpoint string
	serviceName  string
	modelAPIKey  string
	modelBaseURL string
	modelName    string
	prompt       string
	interval     time.Duration
	debugOutput  bool
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
		if err := shutdownTelemetry(); err != nil {
			log.Printf("shutdown telemetry: %v", err)
		}
	}()

	agent := newAgent(cfg)
	target, err := localAgentURL(cfg.host)
	if err != nil {
		return err
	}

	// ❗❗【关键点】WithAgent 时 A2A Server 会在内部为该 Agent 构建 Runner 并在当前进程执行，
	// 因此 Agent / LLM / Tool 的内置 Span 与 GenAI 指标都在当前进程产生并自动上报。
	srv, err := a2aserver.New(
		a2aserver.WithHost(target),
		a2aserver.WithAgent(agent, true),
	)
	if err != nil {
		return fmt.Errorf("create a2a server: %w", err)
	}

	log.Printf("A2A server listening on http://%s (local agent URL: %s)", cfg.host, target)

	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	group, groupCtx := errgroup.WithContext(signalCtx)
	group.Go(func() error {
		if err := srv.Start(cfg.host); err != nil {
			return fmt.Errorf("a2a server stopped: %w", err)
		}
		if groupCtx.Err() == nil {
			return fmt.Errorf("a2a server stopped unexpectedly")
		}
		return nil
	})
	group.Go(func() error {
		if err := loopQuery(groupCtx, cfg, target); err != nil {
			return fmt.Errorf("A2A client loopQuery stopped: %w", err)
		}
		return nil
	})
	group.Go(func() error {
		<-groupCtx.Done()
		log.Println("shutting down A2A server...")
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Stop(stopCtx); err != nil {
			return fmt.Errorf("stop a2a server: %w", err)
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		return err
	}
	return nil
}

// setupTelemetry installs the global Trace and Metric providers before the
// Agent is created, so tRPC Agent Go's built-in telemetry is exported.
func setupTelemetry(ctx context.Context, cfg config) (func() error, error) {
	header := "x-bk-token=" + url.QueryEscape(cfg.token)
	if err := os.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", header); err != nil {
		return nil, fmt.Errorf("set metrics headers: %w", err)
	}

	meterProvider, err := ametric.NewMeterProvider(
		ctx,
		// ❗❗【非常重要】请使用 APM 接入指引提供的 HTTP OTLP 地址，格式为 host:port。
		ametric.WithEndpoint(cfg.otlpEndpoint),
		ametric.WithProtocol("http"),
		ametric.WithServiceName(cfg.serviceName),
	)
	if err != nil {
		return nil, fmt.Errorf("create meter provider: %w", err)
	}
	if err := ametric.InitMeterProvider(meterProvider); err != nil {
		shutdownErr := meterProvider.Shutdown(ctx)
		return nil, errors.Join(fmt.Errorf("init meter provider: %w", err), shutdownErr)
	}

	shutdownTrace, err := atrace.Start(
		ctx,
		// ❗❗【非常重要】请使用 APM 接入指引提供的 HTTP OTLP 地址，格式为 host:port。
		atrace.WithEndpoint(cfg.otlpEndpoint),
		atrace.WithProtocol("http"),
		// ❗❗【非常重要】请将 APM 应用 Token 作为 x-bk-token Header 传入。
		atrace.WithHeaders(map[string]string{"x-bk-token": cfg.token}),
		atrace.WithServiceName(cfg.serviceName),
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

func loadConfig() config {
	return config{
		host:         envOrDefault("HOST", "127.0.0.1:8080"),
		token:        strings.TrimSpace(os.Getenv("TOKEN")),
		otlpEndpoint: strings.TrimSpace(os.Getenv("OTLP_ENDPOINT")),
		serviceName:  envOrDefault("SERVICE_NAME", appName),
		modelAPIKey:  strings.TrimSpace(os.Getenv("MODEL_API_KEY")),
		modelBaseURL: strings.TrimSpace(os.Getenv("MODEL_BASE_URL")),
		modelName:    envOrDefault("MODEL_NAME", "gpt-4o-mini"),
		prompt:       envOrDefault("PROMPT", defaultPrompt),
		interval:     envDurationSeconds("INTERVAL_SECONDS", 30*time.Second),
		debugOutput:  strings.EqualFold(envOrDefault("DEBUG_OUTPUT", "false"), "true"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
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
	if _, err := localAgentURL(cfg.host); err != nil {
		return err
	}
	return nil
}
