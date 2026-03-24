// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package http 提供了 HTTP 服务的实现
package http

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/config"
	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/define"
)

type routerManager struct {
	httpRouter *mux.Router
	httpRoutes map[string]define.RouteInfo
}

var routerMgr = &routerManager{httpRouter: mux.NewRouter(), httpRoutes: map[string]define.RouteInfo{}}

func init() {
	registerHttpRoute("helloworld", http.MethodGet, "/helloworld", HelloWorld, routerMgr)
}

// Router 返回 HTTP 路由器
func Router() *mux.Router {
	return routerMgr.httpRouter
}

// registerHttpRoute 注册路由
func registerHttpRoute(source, httpMethod, relativePath string, handleFunc http.HandlerFunc, r *routerManager) {
	ri := define.RouteInfo{
		Source:     source,
		HttpMethod: httpMethod,
		Path:       relativePath,
	}
	if _, ok := r.httpRoutes[ri.Key()]; ok {
		panic(fmt.Errorf("[http] duplicated http route '%v'", ri))
	}

	r.httpRoutes[ri.Key()] = ri
	r.httpRouter.HandleFunc(relativePath, handleFunc).Methods(httpMethod)
}

func fmtHyperlink(url string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\\n", url, url)
}

// Service 定义了 HTTP 服务的核心结构
type Service struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	config *Config
	Server *http.Server
}

// Type 返回服务类型标识
func (s *Service) Type() string {
	return "http"
}

// Init 初始化 HTTP 服务
func (s *Service) Init(conf *config.Config) error {
	s.config = &Config{Endpoint: fmt.Sprintf("%s:%d", conf.ServerAddress, conf.ServerPort)}
	s.Server = &http.Server{
		// 增加 HTTP Server Instrument
		// TODO 抽象成中间件
		Handler:      otelhttp.NewHandler(Router(), "HTTP Server"),
		ReadTimeout:  time.Minute * 5,
		WriteTimeout: time.Minute * 5,
	}
	return nil
}

// Start 启动 HTTP 服务
func (s *Service) Start(ctx context.Context) error {
	errs := make(chan error, 4)
	s.ctx, s.cancel = context.WithCancel(ctx)

	// 启动 HTTP 服务
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		// ErrServerClosed 是 Shutdown 抛出的错误，属于预期内退出
		if err := s.startHttpServer(); err != nil && err != http.ErrServerClosed {
			errs <- err
		}
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		url := fmt.Sprintf("http://%s/helloworld", s.config.Endpoint)
		log.Printf("[%v] start LoopQueryHelloWorld to periodically request %s", s.Type(), fmtHyperlink(url))
		LoopQueryHelloWorld(s.ctx, url)
		log.Printf("[%v] LoopQueryHelloWorld stopped", s.Type())
	}()

	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		go func() {
			for err := range errs {
				log.Printf("[%v] server got err: %v", s.Type(), err)
			}
		}()
		return nil
	case err := <-errs:
		return err
	}
}

// Stop 停止 HTTP 服务
func (s *Service) Stop() error {
	if s.Server != nil {
		if err := s.Server.Shutdown(s.ctx); err != nil {
			return err
		}
		log.Printf("[%v] server stopped", s.Type())
	}

	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()

	return nil
}

func (s *Service) startHttpServer() error {
	endpoint := s.config.Endpoint
	log.Printf("[%v] start to listen http server at %v", s.Type(), endpoint)
	l, err := net.Listen("tcp", endpoint)
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}
