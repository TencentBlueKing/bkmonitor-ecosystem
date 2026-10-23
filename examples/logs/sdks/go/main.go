// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
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
	"net/url"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func run(ctx context.Context) error {
	// ❗❗【非常重要】填写页面提供的完整 OTLP/HTTP 日志地址，路径以 /v1/logs 结尾。
	endpoint := strings.TrimSpace(os.Getenv("API_URL"))
	// ❗❗【非常重要】填写当前日志数据源的 Token，Exporter 将其放在 x-bk-token 请求头中。
	token := strings.TrimSpace(os.Getenv("TOKEN"))
	if endpoint == "" || token == "" {
		return fmt.Errorf("请设置 API_URL 和 TOKEN")
	}
	u, err := url.Parse(endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" ||
		u.User != nil || u.RawQuery != "" || u.Fragment != "" || !strings.HasSuffix(u.Path, "/v1/logs") {
		return fmt.Errorf("API_URL 必须是以 /v1/logs 结尾的完整 HTTP(S) 地址，且不能包含查询参数或片段")
	}

	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(endpoint),
		otlploghttp.WithHeaders(map[string]string{"x-bk-token": token}),
	)
	if err != nil {
		return fmt.Errorf("创建 OTLP 日志 Exporter: %w", err)
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource.NewWithAttributes("",
			attribute.String("service.name", "logs-sdk-go"),
			attribute.String("deployment.environment.name", "local"),
		)),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
	logger := otelslog.NewLogger("logs-sdk-go", otelslog.WithLoggerProvider(provider))
	logger.InfoContext(ctx, "Go SDK 日志上报示例", "demo.source", "go-sdk")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := provider.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("发送日志: %w", err)
	}
	return nil
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("日志已发送，请到对应索引集检索")
}
