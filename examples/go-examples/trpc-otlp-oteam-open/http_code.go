// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 提供了 TRPC 与 OpenTelemetry 集成示例应用的 HTTP 状态码测试功能
package main

import (
	"fmt"
	"net/http"
	"time"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
	thttp "trpc.group/trpc-go/trpc-go/http"
	"trpc.group/trpc-go/trpc-go/log"

	"bk-apm/bkmonitor-ecosystem/examples/go-examples/trpc-otlp-oteam-open/greeter"
)

func periodicHTTPGet() {
	for {
		cli := thttp.NewClientProxy("trpc.example.greeter.http", client.WithTimeout(time.Second*2))
		cli.Get(trpc.BackgroundContext(), "/404", nil)
		cli.Get(trpc.BackgroundContext(), "/200", nil)
		cli.Get(trpc.BackgroundContext(), "/500", nil)
		cli.Get(trpc.BackgroundContext(), "/timeout", nil)
		// 关键代码，通过 client.WithCalleeMethod("xxxx") 将 path 映射为固定名称以降低主调场景「被调接口」基数。
		cli.Get(trpc.BackgroundContext(), "/trpc_info_test/guc_info/6666", nil, client.WithCalleeMethod("GetGucInfo"))
		cli.Get(
			trpc.BackgroundContext(),
			"/trpc_info_test/guc_info/6666/_update",
			nil,
			client.WithCalleeMethod("UpdateGucInfo"),
		)
		header := &thttp.ClientReqHeader{Header: http.Header{"Tencent-Leakscan": []string{"true"}}}
		cli.Get(trpc.BackgroundContext(), "/700", nil, client.WithReqHead(header))
		time.Sleep(3 * time.Second)
	}
}

func send(status int) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		proxy := greeter.NewGreeterClientProxy(client.WithTimeout(time.Second))
		ctx := r.Context()
		rsp, err := proxy.SayHello(
			ctx, &greeter.HelloRequest{
				Msg: fmt.Sprint("hello ", status),
			},
		)
		if err != nil {
			log.ErrorContextf(ctx, "err: %v", err)
			return err
		}
		w.WriteHeader(status)
		w.Write([]byte(rsp.Msg))
		return nil
	}
}

func timeout(w http.ResponseWriter, r *http.Request) error {
	time.Sleep(3 * time.Second)
	return nil
}
