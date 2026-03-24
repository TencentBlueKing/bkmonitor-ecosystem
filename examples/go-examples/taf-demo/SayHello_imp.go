// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 提供了 TAF 框架示例应用的 SayHello 服务实现
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/ZhuoZhuoCrayon/wasm-demo/src/TestApp/HelloGo/TestApp"
)

// RandomError 表示随机生成的错误
var RandomError = errors.New("random error")

func randSleep() {
	r := 100 + rand.Intn(200)
	time.Sleep(time.Duration(r) * time.Millisecond)
}

func randError(errRate float64) (int32, error) {
	if rand.Float64() < errRate {
		// 选择 -1 ～ -13 的错误码
		return int32(-1 * (rand.Intn(12) + 1)), RandomError
	}
	return 0, nil
}

// SayHelloImp 实现了 SayHello 服务的 servant 接口
type SayHelloImp struct {
	properties *Properties
}

// NewSayHelloImp 创建并返回一个新的 SayHelloImp 实例
func NewSayHelloImp(properties *Properties) *SayHelloImp {
	return &SayHelloImp{properties: properties}
}

// Init 初始化 servant
func (imp *SayHelloImp) Init() error {
	return nil
}

// Destroy 销毁 servant
func (imp *SayHelloImp) Destroy() {
}

// Add 执行两个整数的加法运算，将结果存储在 c 中，并返回错误码和可能的错误
func (imp *SayHelloImp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	rctx, ok := current.GetRequestContext(ctx)
	if ok {
		log.Printf("[Add] rctx -> %v", rctx)
	}
	imp.properties.addRequests.Report(1)
	randSleep()
	*c = a + b
	imp.properties.add.Report(int(*c))
	log.Printf("a(%v) + b(%v) = c(%v)", a, b, *c)
	return randError(0.01)
}

// Sub 执行两个整数的减法运算，将结果存储在 c 中，并返回错误码和可能的错误
func (imp *SayHelloImp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	imp.properties.subRequests.Report(1)
	randSleep()
	*c = a - b
	imp.properties.sub.Report(int(*c))
	log.Printf("a(%v) - b(%v) = c(%v)", a, b, *c)
	return randError(0.01)
}

func loopQueryAdd() {
	comm := tars.NewCommunicator()
	obj := fmt.Sprintf("TestApp.HelloGo.SayHelloObj@tcp -h 127.0.0.1 -p 13000 -t 60000")
	app := new(TestApp.SayHello)
	comm.StringToProxy(obj, app)

	tick := time.Tick(time.Second * 3)
	for range tick {
		var out int32
		ret, err := app.AddWithContext(context.Background(), rand.Int31n(1000), rand.Int31n(1000), &out)
		if err != nil {
			log.Printf("queryAdd got error -> %v", err)
		} else {
			log.Printf("queryAdd got rsp -> %v", ret)
		}
	}
}

func loopQuerySub() {
	comm := tars.NewCommunicator()
	obj := fmt.Sprintf("TestApp.HelloGo.SayHelloObj@tcp -h 127.0.0.1 -p 13000 -t 60000")
	app := new(TestApp.SayHello)
	comm.StringToProxy(obj, app)

	tick := time.Tick(time.Second * 3)
	for range tick {
		var out int32
		ret, err := app.SubWithContext(context.Background(), rand.Int31n(1000), rand.Int31n(1000), &out)
		if err != nil {
			log.Printf("querySub got error -> %v", err)
		} else {
			log.Printf("querySub got rsp -> %v", ret)
		}
	}
}
