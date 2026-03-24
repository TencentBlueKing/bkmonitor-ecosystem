// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package service 定义了服务接口和基础实现
package service

import (
	"context"

	"jaeger-client-demo/config"
)

// Service 定义了查询器服务的核心结构
type Service interface {
	Init(conf *config.Config, ctx context.Context) error
	Start() error
	Stop() error
	Type() string
}
