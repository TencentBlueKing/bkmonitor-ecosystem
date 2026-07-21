// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2026 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 通过 A2A 协议远程调用 server 暴露的 Agent。
//
// client 只是一个调用方：它拿到远端 Agent 的 Agent Card，再用 Runner 发起请求。
// Agent 的真正执行（调用 LLM、调用 Tool）发生在 server 进程，因此 Agent 的调用链与
// GenAI 指标由 server 端上报；client 端只负责发起调用与消费返回事件。
//
// 为了让 APM 持续有可观测数据，client 默认按固定间隔循环发起请求（数据自生成）。
// 通过 INTERVAL_SECONDS 控制间隔；设为 0 则只跑一次就退出（便于本地调试）。
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"trpc.group/trpc-go/trpc-agent-go/agent/a2aagent"
	"trpc.group/trpc-go/trpc-agent-go/event"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/runner"
)

func main() {
	// 监听退出信号，收到后取消 context，循环优雅退出。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	target := envOrDefault("TARGET", "http://127.0.0.1:8080")
	prompt := envOrDefault("PROMPT", "Use the calculator to compute 12 * 7.")
	interval := envDurationSeconds("INTERVAL_SECONDS", 30*time.Second)

	// 连接远端 Agent（读取 Agent Card），得到一个可被 Runner 驱动的 Agent。
	remoteAgent, err := a2aagent.New(a2aagent.WithAgentCardURL(target))
	if err != nil {
		return fmt.Errorf("create a2a agent: %w", err)
	}
	agentRunner := runner.NewRunner("trpc-agent-go-apm-demo-client", remoteAgent)

	// 启动先立即调用一次，让 APM 尽快有数据。
	if err := invoke(ctx, agentRunner, prompt); err != nil {
		return err
	}

	// interval <= 0：只跑一次就退出（本地调试）。
	if interval <= 0 {
		return nil
	}

	// 定期循环发起请求，持续自生成可观测数据。
	log.Printf("client running in loop, interval=%s (send SIGINT/SIGTERM to stop)", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Print("client stopped")
			return nil
		case <-ticker.C:
			if err := invoke(ctx, agentRunner, prompt); err != nil {
				// 单次失败不退出，记录后继续下一轮，保证长期运行。
				log.Printf("invoke error: %v", err)
			}
		}
	}
}

// invoke 发起一次远程 Agent 调用并打印返回事件。
func invoke(ctx context.Context, agentRunner runner.Runner, prompt string) error {
	events, err := agentRunner.Run(
		ctx,
		"demo-user",
		"demo-session",
		model.NewUserMessage(prompt),
	)
	if err != nil {
		return fmt.Errorf("remote run: %w", err)
	}

	fmt.Printf("User: %s\n", prompt)
	return printEvents(events)
}

func printEvents(events <-chan *event.Event) error {
	var final strings.Builder
	streamed := false
	for evt := range events {
		if evt == nil || evt.Response == nil {
			continue
		}
		if evt.Error != nil {
			return fmt.Errorf("runner event error: %s", evt.Error.Message)
		}
		printToolActivity(evt)
		for _, choice := range evt.Choices {
			if choice.Delta.Content != "" {
				final.WriteString(choice.Delta.Content)
				streamed = true
				continue
			}
			if !streamed && choice.Message.Role == model.RoleAssistant && choice.Message.Content != "" {
				final.WriteString(choice.Message.Content)
			}
		}
		if evt.IsRunnerCompletion() {
			break
		}
	}
	if final.Len() > 0 {
		fmt.Printf("Assistant: %s\n", final.String())
	}
	return nil
}

func printToolActivity(evt *event.Event) {
	for _, choice := range evt.Choices {
		for _, toolCall := range choice.Message.ToolCalls {
			fmt.Printf("[tool call] %s(%s)\n", toolCall.Function.Name, string(toolCall.Function.Arguments))
		}
		if choice.Message.Role == model.RoleTool && choice.Message.Content != "" {
			fmt.Printf("[tool result] %s\n", choice.Message.Content)
		}
	}
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

// envDurationSeconds 读取秒数环境变量；非法或缺省时返回 fallback，"0" 表示只跑一次。
func envDurationSeconds(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("invalid %s=%q, use default %s", key, raw, fallback)
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
