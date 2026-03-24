// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package define 定义了应用程序的常量和类型
package define

import "fmt"

// RouteInfo 路由信息
type RouteInfo struct {
	Source     string
	HttpMethod string
	Path       string
}

// Key 路由唯一标识
func (r RouteInfo) Key() string {
	return fmt.Sprintf("%s %s", r.HttpMethod, r.Path)
}

// ID 路由唯一标识
func (r RouteInfo) ID() string {
	return fmt.Sprintf("%s/%s/%s", r.Source, r.HttpMethod, r.Path)
}
