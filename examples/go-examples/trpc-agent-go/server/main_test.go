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
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	valid := config{
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

func TestEnvBool(t *testing.T) {
	t.Setenv("BOOL_VALUE", "YES")
	if !envBool("BOOL_VALUE", false) {
		t.Fatal("YES should be parsed as true")
	}
	t.Setenv("BOOL_VALUE", "0")
	if envBool("BOOL_VALUE", true) {
		t.Fatal("0 should be parsed as false")
	}
	t.Setenv("BOOL_VALUE", "invalid")
	if !envBool("BOOL_VALUE", true) {
		t.Fatal("invalid values should use the fallback")
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

func TestShutdownProviders(t *testing.T) {
	traceCtx, cancelTrace := context.WithCancel(context.Background())
	defer cancelTrace()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelShutdown()
	started := time.Now()
	err := shutdownProviders(
		shutdownCtx,
		cancelTrace,
		func() error {
			<-traceCtx.Done()
			return traceCtx.Err()
		},
		func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
		func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("shutdown did not honor its deadline: %s", elapsed)
	}
}

func TestShutdownProvidersSuccess(t *testing.T) {
	_, cancelTrace := context.WithCancel(context.Background())
	err := shutdownProviders(
		context.Background(),
		cancelTrace,
		func() error { return nil },
		func(context.Context) error { return nil },
		func(context.Context) error { return nil },
	)
	if err != nil {
		t.Fatalf("successful shutdown returned an error: %v", err)
	}
}
