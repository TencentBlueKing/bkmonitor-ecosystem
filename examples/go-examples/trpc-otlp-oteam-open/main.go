// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 是由 trpc-go-cmdline v2.4.0 生成的服务端示例代码
// 注意：本文件并非必须存在，而仅为示例，用户应按需进行修改使用，如不需要，可直接删去
package main

import (
	"context"
	"net/http"
	"time"

	trpc "trpc.group/trpc-go/trpc-go"
	thttp "trpc.group/trpc-go/trpc-go/http"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/server"
	_ "trpc.group/trpc-go/trpc-opentelemetry/oteltrpc"
	"trpc.group/trpc-go/trpc-opentelemetry/sdk/metric"

	pb "bk-apm/bkmonitor-ecosystem/examples/go-examples/trpc-otlp-oteam-open/greeter"
)

func main() {
	s := trpc.NewServer()
	defer metric.DeletePrometheusPush()

	// register tRPC Service
	pb.RegisterGreeterService(s, &greeterImpl{proxy: pb.NewGreeterClientProxy()})
	// register HTTP service
	registerHTTPService(s.Service("trpc.example.greeter.http"))

	go loopQuery()
	go periodicHTTPGet()
	go panicEventGoer()

	if err := s.Serve(); err != nil {
		log.Fatal(err)
	}
}

func registerHTTPService(s server.Service) {
	// 注册 `method regex` -> `pattern` 映射，将含有不可枚举值的高基数 `method` 转换为低基数的 `method pattern`。
	metric.RegisterMethodMapping("/404", "/404")
	metric.RegisterMethodMapping(".+/guc_info/\\d+", ".+/guc_info/:id")
	metric.RegisterMethodMapping(".+/guc_info/\\d+/_update", ".+/guc_info/:id/_update")
	// metric.RegisterMethodMapping("/500", "/500")
	// metric.RegisterMethodMapping("/200", "/200")
	// metric.RegisterMethodMapping("/700", "/700")
	// metric.RegisterMethodMapping("/timeout", "/timeout")

	thttp.HandleFunc("/404", send(http.StatusNotFound))
	thttp.HandleFunc("/500", send(http.StatusInternalServerError))
	thttp.HandleFunc("/200", send(http.StatusOK))
	thttp.HandleFunc("/700", send(http.StatusNotFound))
	thttp.HandleFunc("/trpc_info_test/guc_info/6666", send(http.StatusOK))
	thttp.HandleFunc("/trpc_info_test/guc_info/6666/_update", send(http.StatusOK))
	thttp.HandleFunc("/timeout", timeout)
	thttp.RegisterNoProtocolService(s)
}

// panicEvent 每小时 panic 一次，通过 trpc.Go 捕获 panic 事件，在错误日志中过滤出 panic，并上报 panic 事件。
func panicEventGoer() {
	for range time.Tick(time.Hour) {
		trpc.Go(
			context.Background(), time.Second, func(ctx context.Context) {
				panic("mock panic trpc.Go")
			},
		)
	}
}
