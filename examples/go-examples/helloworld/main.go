// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 提供了 helloworld 示例应用的主入口
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/config"
	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/service"
	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/service/http"
	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/service/otlp"
	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/service/profiling"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	serviceList := []service.Service{
		&otlp.Service{},
		&profiling.Service{},
		&http.Service{},
	}

	conf := config.New()
	for _, s := range serviceList {
		if err := s.Init(conf); err != nil {
			log.Fatalf("[%v] failed to init: %v", s.Type(), err)
		}
	}

	stop := func() {
		defer cancel()
		for _, s := range serviceList {
			if err := s.Stop(); err != nil {
				log.Printf("[%v] failed to stop: %v", s.Type(), err)
				// 忽略失败，应停尽停
				// TODO 如果失败，最后可以 panic 掉
				continue
			}
			log.Printf("[%v] service stopped", s.Type())
		}
	}

	for _, s := range serviceList {
		if err := s.Start(ctx); err != nil {
			log.Printf("[%v] failed to start: %v", s.Type(), err)
			stop()
			os.Exit(1)
		}
		log.Printf("[%v] service started", s.Type())
	}

	log.Printf("[main] 🚀")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	<-sigCh
	stop()
	log.Printf("[main] 👋")
}
