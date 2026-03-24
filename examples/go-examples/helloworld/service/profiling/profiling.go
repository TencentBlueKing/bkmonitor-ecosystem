// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package profiling 提供了基于 Pyroscope 的性能分析服务实现
package profiling

import (
	"context"
	"time"

	"github.com/grafana/pyroscope-go"

	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/config"
)

// Service 定义了 Profiling 性能分析服务的核心结构
type Service struct {
	config   *Config
	profiler *pyroscope.Profiler
}

// Type 返回服务类型标识
func (s *Service) Type() string {
	return "profiling"
}

// Init 初始化 Profiling 服务
func (s *Service) Init(conf *config.Config) error {
	s.config = &Config{
		Token:        conf.Token,
		Enabled:      conf.EnableProfiling,
		ServiceName:  conf.ServiceName,
		Addr:         conf.ProfilingEndpoint,
		EnableTraces: conf.EnableTraces,
	}
	return nil
}

// Start 启动 Profiling 服务
func (s *Service) Start(ctx context.Context) error {
	if !s.config.Enabled {
		return nil
	}

	var err error
	s.profiler, err = pyroscope.Start(
		pyroscope.Config{
			//❗❗【非常重要】请传入应用 Token
			AuthToken: s.config.Token,
			//❗❗【非常重要】应用服务唯一标识
			ApplicationName: s.config.ServiceName,
			//❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
			ServerAddress: s.config.Addr,
			// 上报周期，默认 15 s
			UploadRate: 15 * time.Second,
			Logger:     pyroscope.StandardLogger,
			ProfileTypes: []pyroscope.ProfileType{
				// these profile types are enabled by default:
				pyroscope.ProfileCPU,
				pyroscope.ProfileAllocObjects,
				pyroscope.ProfileAllocSpace,
				pyroscope.ProfileInuseObjects,
				pyroscope.ProfileInuseSpace,

				// these profile types are optional:
				pyroscope.ProfileGoroutines,
				pyroscope.ProfileMutexCount,
				pyroscope.ProfileMutexDuration,
				pyroscope.ProfileBlockCount,
				pyroscope.ProfileBlockDuration,
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// Stop 停止 Profiling 服务
func (s *Service) Stop() error {
	if s.profiler == nil {
		return nil
	}

	if err := s.profiler.Stop(); err != nil {
		return err
	}

	return nil
}
