// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//! config 提供应用程序配置管理功能，支持从环境变量读取配置参数。

/// AppConfig 定义应用程序的配置信息，包括 OTLP 和 HTTP 服务等相关配置。
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct AppConfig {
    pub service_name: String,
    pub token: String,
    pub otlp_endpoint: String,
    pub otlp_exporter_type: String,
    pub enable_traces: bool,
    pub enable_metrics: bool,
    pub enable_logs: bool,
    pub server_address: String,
    pub server_port: u16,
}

impl AppConfig {
    /// from_env 创建并返回一个新的 AppConfig 实例。
    pub fn from_env() -> Self {
        Self {
            service_name: std::env::var("SERVICE_NAME").unwrap_or_else(|_| "helloworld".to_owned()),
            token: std::env::var("TOKEN").unwrap_or_else(|_| "todo".to_owned()),
            otlp_endpoint: std::env::var("OTLP_ENDPOINT")
                .unwrap_or_else(|_| "127.0.0.1:4318".to_owned()),
            otlp_exporter_type: std::env::var("OTLP_EXPORTER_TYPE")
                .unwrap_or_else(|_| "http".to_owned()),
            enable_traces: std::env::var("ENABLE_TRACES")
                .ok()
                .and_then(|value| value.parse().ok())
                .unwrap_or(false),
            enable_metrics: std::env::var("ENABLE_METRICS")
                .ok()
                .and_then(|value| value.parse().ok())
                .unwrap_or(false),
            enable_logs: std::env::var("ENABLE_LOGS")
                .ok()
                .and_then(|value| value.parse().ok())
                .unwrap_or(false),
            server_address: std::env::var("SERVER_ADDRESS")
                .unwrap_or_else(|_| "127.0.0.1".to_owned()),
            server_port: 8080,
        }
    }
}
