// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//! HTTP 服务端与内置 HTTP 客户端。

use axum::http::StatusCode;
use opentelemetry::trace::Status;
use tracing_opentelemetry::OpenTelemetrySpanExt;

pub mod client;
pub mod server;

fn set_http_request_attributes(span: &tracing::Span) {
    span.set_attribute("http.request.method", "GET");
    span.set_attribute("http.route", "/helloworld");
}

fn set_http_response_attributes(span: &tracing::Span, status: StatusCode) {
    span.set_attribute("http.response.status_code", i64::from(status.as_u16()));
    if status.is_server_error() {
        span.set_status(Status::error(status.to_string()));
    }
}