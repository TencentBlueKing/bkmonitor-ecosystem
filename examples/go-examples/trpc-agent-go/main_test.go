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
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestConfigValidate verifies required runtime configuration before startup.
func TestConfigValidate(t *testing.T) {
	valid := config{
		host:         "127.0.0.1:8080",
		token:        "token",
		otlpEndpoint: "collector.example.com:4318",
		modelAPIKey:  "model-key",
	}
	if err := valid.validate(); err != nil {
		t.Fatalf("valid config returned an error: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*config)
	}{
		{name: "missing token", mutate: func(cfg *config) { cfg.token = "" }},
		{name: "missing endpoint", mutate: func(cfg *config) { cfg.otlpEndpoint = "" }},
		{name: "missing model key", mutate: func(cfg *config) { cfg.modelAPIKey = "" }},
		{name: "endpoint contains scheme", mutate: func(cfg *config) {
			cfg.otlpEndpoint = "http://collector.example.com:4318"
		}},
		{name: "invalid host", mutate: func(cfg *config) { cfg.host = "127.0.0.1" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := valid
			test.mutate(&cfg)
			if err := cfg.validate(); err == nil {
				t.Fatal("invalid config did not return an error")
			}
		})
	}
}

func TestLocalAgentURL(t *testing.T) {
	tests := []struct {
		listenAddress string
		want          string
	}{
		{listenAddress: "0.0.0.0:8080", want: "http://127.0.0.1:8080"},
		{listenAddress: ":8080", want: "http://127.0.0.1:8080"},
		{listenAddress: "[::]:8080", want: "http://127.0.0.1:8080"},
		{listenAddress: "localhost:8080", want: "http://localhost:8080"},
	}
	for _, test := range tests {
		t.Run(test.listenAddress, func(t *testing.T) {
			got, err := localAgentURL(test.listenAddress)
			if err != nil {
				t.Fatalf("localAgentURL returned an error: %v", err)
			}
			if got != test.want {
				t.Fatalf("localAgentURL(%q) = %q, want %q", test.listenAddress, got, test.want)
			}
		})
	}
}

func TestEnvDurationSeconds(t *testing.T) {
	t.Setenv("INTERVAL_SECONDS", "5")
	if got := envDurationSeconds("INTERVAL_SECONDS", 30*time.Second); got != 5*time.Second {
		t.Fatalf("envDurationSeconds = %s, want 5s", got)
	}
	t.Setenv("INTERVAL_SECONDS", "invalid")
	if got := envDurationSeconds("INTERVAL_SECONDS", 30*time.Second); got != 30*time.Second {
		t.Fatalf("invalid envDurationSeconds = %s, want fallback 30s", got)
	}
}

func TestCalculator(t *testing.T) {
	cases := []struct {
		op   string
		a, b float64
		want float64
	}{
		{"add", 1, 2, 3},
		{"subtract", 5, 3, 2},
		{"multiply", 12, 7, 84},
		{"divide", 8, 2, 4},
		{"power", 2, 10, 1024},
	}
	for _, c := range cases {
		got, err := calculator(context.Background(), calculatorArgs{Operation: c.op, A: c.a, B: c.b})
		if err != nil {
			t.Fatalf("%s returned an error: %v", c.op, err)
		}
		if got.Result != c.want {
			t.Fatalf("%s(%g,%g) = %g, want %g", c.op, c.a, c.b, got.Result, c.want)
		}
	}
	if _, err := calculator(context.Background(), calculatorArgs{Operation: "divide", A: 1, B: 0}); err == nil {
		t.Fatal("divide by zero should return an error")
	}
	if _, err := calculator(context.Background(), calculatorArgs{Operation: "modulo", A: 1, B: 2}); err == nil {
		t.Fatal("unsupported operation should return an error")
	}
}

func TestWaitForLocalRunnerHonorsContext(t *testing.T) {
	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	var startedOnce sync.Once
	var canceledOnce sync.Once
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		startedOnce.Do(func() { close(requestStarted) })
		<-request.Context().Done()
		canceledOnce.Do(func() { close(requestCanceled) })
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	startedAt := time.Now()
	_, err := waitForLocalRunner(ctx, server.URL)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waitForLocalRunner error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(startedAt); elapsed > time.Second {
		t.Fatalf("waitForLocalRunner did not honor its context: %s", elapsed)
	}

	select {
	case <-requestStarted:
	default:
		t.Fatal("agent card request was not started")
	}
	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("agent card request context was not canceled")
	}
}
