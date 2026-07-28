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
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	a2aclient "trpc.group/trpc-go/trpc-a2a-go/client"
	"trpc.group/trpc-go/trpc-agent-go/agent/a2aagent"
	"trpc.group/trpc-go/trpc-agent-go/event"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/runner"
)

const (
	defaultPrompt       = "Use the calculator to compute 12 * 7."
	clientRetryInterval = 200 * time.Millisecond
	clientStartTimeout  = 10 * time.Second
	clientQueryTimeout  = 2 * time.Minute
)

// localAgentURL converts the server listen address into a loopback URL.
// Wildcard listen addresses cannot be advertised to the in-process client.
func localAgentURL(listenAddress string) (string, error) {
	host, port, err := net.SplitHostPort(listenAddress)
	if err != nil {
		return "", fmt.Errorf("parse HOST %q: %w", listenAddress, err)
	}
	switch host {
	case "", "0.0.0.0", "::":
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port), nil
}

// loopQuery runs the client side of the C/S demo in the same process. It creates
// one A2A client proxy and periodically starts queryAgent in a separate goroutine.
func loopQuery(ctx context.Context, cfg config, target string) error {
	startCtx, cancelStart := context.WithTimeout(ctx, clientStartTimeout)
	defer cancelStart()

	agentRunner, err := waitForLocalRunner(startCtx, target)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	defer func() {
		if err := agentRunner.Close(); err != nil {
			log.Printf("close A2A client runner: %v", err)
		}
	}()

	var queries sync.WaitGroup
	queryRunning := make(chan struct{}, 1)
	startQuery := func() {
		select {
		case queryRunning <- struct{}{}:
		default:
			log.Print("A2A client skipped query because the previous query is still running")
			return
		}
		queries.Add(1)
		go func() {
			defer queries.Done()
			defer func() { <-queryRunning }()
			if err := queryAgent(ctx, agentRunner, cfg.prompt, cfg.debugOutput); err != nil {
				if ctx.Err() == nil {
					log.Printf("A2A client query failed: %v", err)
				}
			}
		}()
	}

	startQuery()
	if cfg.interval <= 0 {
		queries.Wait()
		log.Print("A2A client completed one query")
		return nil
	}

	log.Printf("A2A client loopQuery started, interval=%s", cfg.interval)
	ticker := time.NewTicker(cfg.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			queries.Wait()
			log.Print("A2A client loopQuery stopped")
			return nil
		case <-ticker.C:
			startQuery()
		}
	}
}

func waitForLocalRunner(ctx context.Context, target string) (runner.Runner, error) {
	ticker := time.NewTicker(clientRetryInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		remoteAgent, err := newRemoteAgent(ctx, target)
		if err == nil {
			return runner.NewRunner(appName+"-client", remoteAgent), nil
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for local A2A server: %w (last error: %v)", ctx.Err(), lastErr)
		case <-ticker.C:
		}
	}
}

func newRemoteAgent(ctx context.Context, target string) (*a2aagent.A2AAgent, error) {
	client, err := a2aclient.NewA2AClient(target)
	if err != nil {
		return nil, fmt.Errorf("create A2A client: %w", err)
	}

	card, err := client.GetAgentCard(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("get agent card: %w", err)
	}
	if card.URL == "" {
		card.URL = target
	}

	return a2aagent.New(a2aagent.WithAgentCard(card))
}

func queryAgent(
	ctx context.Context,
	agentRunner runner.Runner,
	prompt string,
	debugOutput bool,
) error {
	queryCtx, cancel := context.WithTimeout(ctx, clientQueryTimeout)
	defer cancel()

	events, err := agentRunner.Run(
		queryCtx,
		"demo-user",
		fmt.Sprintf("demo-session-%d", time.Now().UnixNano()),
		model.NewUserMessage(prompt),
	)
	if err != nil {
		return fmt.Errorf("remote run: %w", err)
	}
	if debugOutput {
		fmt.Printf("User: %s\n", prompt)
	}
	return consumeEvents(events, debugOutput)
}

func consumeEvents(events <-chan *event.Event, debugOutput bool) error {
	var final strings.Builder
	streamed := false
	for evt := range events {
		if evt == nil {
			continue
		}
		if evt.Error != nil {
			return fmt.Errorf("runner event error: %s", evt.Error.Message)
		}
		if evt.Response == nil {
			continue
		}
		if debugOutput {
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
		}
		if evt.IsRunnerCompletion() {
			break
		}
	}
	if debugOutput && final.Len() > 0 {
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

// envDurationSeconds reads a duration in seconds. Zero or negative values make
// loopQuery send one request without starting the periodic loop.
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
