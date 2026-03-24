// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 提供了 TRPC 与 OpenTelemetry 集成示例应用的主入口
package main

import (
	"context"
	"math/rand"
	"time"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/metrics"

	pb "bk-apm/bkmonitor-ecosystem/examples/go-examples/trpc-otlp-oteam-open/greeter"
)

var countries = []string{
	"United States", "Canada", "United Kingdom", "Germany", "France", "Japan", "Australia", "China", "India", "Brazil",
}

type greeterImpl struct {
	pb.UnimplementedGreeter
	proxy pb.GreeterClientProxy
}

func (s *greeterImpl) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Infof("SayHello recv req: %s", req.Msg)

	metricsCounterDemo()
	metricsCustomMetricsDemo(req.Msg)

	hi, err := s.proxy.SayHi(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.HelloReply{Msg: "Hello " + hi.Msg}, nil
}

func (s *greeterImpl) SayHi(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Msg: "Hi" + req.Msg}, nil
}

func (s *greeterImpl) BusinessServer(ctx context.Context, req *pb.BusinessServerRequest) (*pb.BusinessServerReply, error) {
	return &pb.BusinessServerReply{Msg: "rsp success"}, nil
}

func (s *greeterImpl) BusinessError(ctx context.Context, req *pb.BusinessErrorRequest) (*pb.BusinessErrorReply, error) {
	return &pb.BusinessErrorReply{Msg: "rsp error"}, nil
}

func loopQuery() {
	proxy := pb.NewGreeterClientProxy(client.WithTimeout(time.Second * 5))
	tick := time.Tick(time.Second * 3)
	for range tick {
		go querySayHello(proxy)
	}
}

func querySayHello(proxy pb.GreeterClientProxy) {
	rsp, err := proxy.SayHello(trpc.BackgroundContext(), &pb.HelloRequest{Msg: choiceCountry()})
	if err != nil {
		log.Error("querySayHello got error -> %v", err)
	}
	log.Infof("querySayHello got rsp -> %v", rsp.Msg)
}

func choiceCountry() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(countries))
	return countries[randomIndex]
}

func metricsCustomMetricsDemo(country string) {
	err := metrics.ReportMultiDimensionMetricsX(
		"requests",
		[]*metrics.Dimension{{Name: "country", Value: country}},
		[]*metrics.Metrics{
			metrics.NewMetrics("metrics1", 1, metrics.PolicySUM),
			metrics.NewMetrics("metrics2", 1, metrics.PolicySET),
			metrics.NewMetrics("metrics3", 1, metrics.PolicyAVG),
			// metrics.NewMetrics("metrics4", 1, metrics.PolicyHistogram),
		},
	)
	if err != nil {
		log.Errorf("[metricsCustomMetricsDemo] failed to create custom metrics -> %v", err)
	}
}

func metricsCounterDemo() {
	metrics.IncrCounter("requests_total", 1)
}
