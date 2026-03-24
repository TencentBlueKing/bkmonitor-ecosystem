// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 提供了 Jaeger 客户端示例应用的主入口
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"jaeger-client-demo/config"
	"jaeger-client-demo/service"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

// 初始化 Jaeger 追踪器
func initJaeger(serviceName string, conf *config.Config) (opentracing.Tracer, func(), error) {
	collectorEndpoint := ""
	if strings.HasPrefix(conf.BKEndpoint, "http://") || strings.HasPrefix(conf.BKEndpoint, "https://") {
		collectorEndpoint = fmt.Sprintf("%v/jaeger/v1/traces", conf.BKEndpoint)
	} else {
		collectorEndpoint = fmt.Sprintf("http://%v/jaeger/v1/traces", conf.BKEndpoint)
	}

	// 1. 创建 Jaeger 配置
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName, // 服务名称
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst, // 采样类型：全部采样
			Param: 1,                       // 采样率：1=100%
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
			HTTPHeaders: map[string]string{
				"x-bk-token": conf.Token,
			},

			CollectorEndpoint:   collectorEndpoint,
			BufferFlushInterval: 500 * time.Millisecond,
		},
	}

	// 2. 创建追踪器
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jaeger.StdLogger), // 使用标准日志
		// jaegercfg.Tag("bk.data.token", conf.Token),

	)
	if err != nil {
		return nil, nil, fmt.Errorf("无法创建 Jaeger 追踪器: %v", err)
	}

	// 3. 设置为全局追踪器
	opentracing.SetGlobalTracer(tracer)

	// 返回追踪器和关闭函数
	return tracer, func() {
		if err := closer.Close(); err != nil {
			log.Printf("关闭追踪器失败: %v", err)
		}
	}, nil
}

func main() {
	log.Printf("[main] 🚀")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := config.New()
	_, closer, err := initJaeger(conf.ServiceName, conf)
	if err != nil {
		log.Printf("初始化失败: %v", err)
	}
	defer closer() // 程序退出时关闭追踪器

	log.Println("Jaeger 追踪器已初始化")

	serviceList := []service.Service{
		&service.HelloWorldService{},
		&service.QuerierService{},
	}

	for _, s := range serviceList {
		if err := s.Init(conf, ctx); err != nil {
			log.Printf("[%v] failed to init: %v", s.Type(), err)
			return
		}
	}

	defer func() {
		for _, s := range serviceList {
			if err := s.Stop(); err != nil {
				log.Printf("[%v] failed to stop: %v", s.Type(), err)
				// 忽略失败，应停尽停
				// TODO 如果失败，最后可以 panic 掉
				continue
			}
			log.Printf("[%v] service stopped", s.Type())
		}
		log.Printf("[main] 👋")
	}()

	errors := make(chan error)
	for _, s := range serviceList {
		log.Printf("[%v] service starting", s.Type())
		go func() {
			if err := s.Start(); err != nil {
				log.Printf("[%v] failed to start: %v", s.Type(), err)
				errors <- err
			}
		}()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-sigCh:
		return
	case <-errors:
		log.Printf("服务启动报错，自动退出")
		return
	}
}
