// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package define 定义了应用程序的常量和类型
package define

// ExporterType 定义了 exporter 的类型
type ExporterType string

const (
	// ExporterHttp HTTP 协议的 exporter 类型
	ExporterHttp ExporterType = "http"
	// ExporterGRPC gRPC 协议的 exporter 类型
	ExporterGRPC ExporterType = "grpc"
)
