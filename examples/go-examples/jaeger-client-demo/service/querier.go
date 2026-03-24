// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package service 提供了 Jaeger 客户端示例应用的查询器服务实现
package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"jaeger-client-demo/config"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// QuerierService 提供查询器服务的实现
type QuerierService struct {
	ctx     context.Context
	cancel  context.CancelFunc
	address string
	port    int
}

// Init 初始化查询器服务
func (qs *QuerierService) Init(conf *config.Config, ctx context.Context) error {
	qs.ctx, qs.cancel = context.WithCancel(ctx)
	qs.address = conf.ServerAddress
	qs.port = conf.ServerPort
	return nil
}

// Start 启动查询器服务
func (qs *QuerierService) Start() error {
	url := fmt.Sprintf("http://%s:%d/helloworld", qs.address, qs.port)
	log.Printf("url:", url)
	if err := qs.loopQueryHelloWorld(qs.ctx, url); err != nil {
		return err
	}
	return nil
}

// Stop 停止查询器服务
func (qs *QuerierService) Stop() error {
	return nil
}

// Type 返回服务类型标识
func (qs *QuerierService) Type() string {
	return "QuerierService"
}

func (qs *QuerierService) queryHelloWorld(ctx context.Context, url string) error {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("Caller/queryHelloWorld")
	defer span.Finish()
	ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)

	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		// 记录错误到日志和追踪
		log.Printf("[queryHelloWorld] 创建请求失败: %v", err)
		return err
	}

	err = tracer.Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
	if err != nil {
		log.Printf("trace 跟踪上下文注入失败: %v", err)
		return err
	}

	log.Printf(fmt.Sprintf("[queryHelloWorld] send request"))
	res, err := client.Do(req)
	if err != nil {
		log.Printf(fmt.Sprintf("[queryHelloWorld] got error -> %v", err))
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf(fmt.Sprintf("[queryHelloWorld] got error -> %v", err))
		return err
	}
	log.Printf(fmt.Sprintf("[queryHelloWorld] received: %s", body))
	return nil
}

func (qs *QuerierService) loopQueryHelloWorld(ctx context.Context, url string) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := qs.queryHelloWorld(ctx, url); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}
