// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package http 提供了 HTTP 服务的 HelloWorld 接口实现
package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var countries = []string{
	"United States", "Canada", "United Kingdom", "Germany", "France", "Japan", "Australia", "China", "India", "Brazil",
}

var (
	errMySQLConnectTimeout = errors.New("mysql connect timeout")
	errUserNotFound        = errors.New("user not found")
	errNetworkUnreachable  = errors.New("network unreachable")
	errFileNotFound        = errors.New("file not found")

	customErrors = []error{
		errMySQLConnectTimeout, errUserNotFound, errNetworkUnreachable, errFileNotFound,
	}
)

func queryHelloWorld(ctx context.Context, url string) {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	log.Printf("[queryHelloWorld] send request")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[queryHelloWorld] got error -> %v", err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("[queryHelloWorld] got error -> %v", err)
		return
	}
	log.Printf("[queryHelloWorld] received: %s", body)
}

// LoopQueryHelloWorld 定期循环调用 HelloWorld 服务
func LoopQueryHelloWorld(ctx context.Context, url string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			queryHelloWorld(ctx, url)
		case <-ctx.Done():
			return
		}
	}
}

// HelloWorld 处理 HTTP 请求并返回问候语
func HelloWorld(w http.ResponseWriter, req *http.Request) {
	logsDemo(req)

	country := choiceCountry()
	log.Printf("get country -> %s", country)

	if err := randomErrorDemo(); err != nil {
		log.Printf("[randomErrorDemo] got error -> %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	greeting := generateGreeting(country)
	w.Write([]byte(greeting))
}

func choiceCountry() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(countries))
	return countries[randomIndex]
}

func choiceErr() error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(customErrors))
	return customErrors[randomIndex]
}

func generateGreeting(country string) string {
	return fmt.Sprintf("Hello World, %s!", country)
}

func randErr(errRate float64) error {
	if rand.Float64() < errRate {
		return choiceErr()
	}
	return nil
}

// logsDemo 打印请求日志
func logsDemo(req *http.Request) {
	log.Printf("received request: %s %s", req.Method, req.URL)
}

func randomErrorDemo() error {
	return randErr(0.1)
}
