// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package config 提供应用程序配置管理功能，支持从环境变量读取配置参数
package config

import (
	"os"
	"strconv"

	"bk-apm/bkmonitor-ecosystem/examples/go-examples/helloworld/define"
)

// Config 定义应用程序的配置信息，包括 OTLP、性能分析和 HTTP 服务等相关配置
type Config struct {
	Debug bool
	// otel
	Token            string
	ServiceName      string
	OtlpEndpoint     string
	OtlpExporterType define.ExporterType
	EnableLogs       bool
	EnableTraces     bool
	EnableMetrics    bool
	// profiling
	EnableProfiling   bool
	ProfilingEndpoint string
	// http
	ServerPort    int
	ServerAddress string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// New 创建并返回一个新的 Config 实例
func New() *Config {
	debug := getEnvAsBool("DEBUG", false)
	config := &Config{
		Debug:             debug,
		Token:             getEnv("TOKEN", "todo"),
		ServiceName:       getEnv("SERVICE_NAME", "helloworld"),
		OtlpEndpoint:      getEnv("OTLP_ENDPOINT", "localhost:4317"),
		OtlpExporterType:  define.ExporterType(getEnv("OTLP_EXPORTER_TYPE", string(define.ExporterGRPC))),
		EnableLogs:        getEnvAsBool("ENABLE_LOGS", debug),
		EnableTraces:      getEnvAsBool("ENABLE_TRACES", debug),
		EnableMetrics:     getEnvAsBool("ENABLE_METRICS", debug),
		EnableProfiling:   getEnvAsBool("ENABLE_PROFILING", debug),
		ProfilingEndpoint: getEnv("PROFILING_ENDPOINT", "http://localhost:4040"),
		ServerPort:        8080,
		ServerAddress:     getEnv("SERVER_ADDRESS", "127.0.0.1"),
	}
	return config
}
