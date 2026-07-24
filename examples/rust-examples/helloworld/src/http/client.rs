// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//! 定时发起请求、为示例持续产生观测数据的内置客户端。

use opentelemetry::global;
use opentelemetry_http::HeaderInjector;
use tokio::time::sleep;
use tracing::Instrument;
use tracing_opentelemetry::OpenTelemetrySpanExt;

async fn query_hello_world(client: &reqwest::Client, url: &str) {
    let span = tracing::info_span!("Caller/queryHelloWorld", otel.kind = "client");

    let mut request = match client.get(url).build() {
        Ok(request) => request,
        Err(error) => {
            tracing::error!(%error, "[query_hello_world] got error");
            return;
        }
    };
    global::get_text_map_propagator(|propagator| {
        propagator.inject_context(&span.context(), &mut HeaderInjector(request.headers_mut()));
    });

    async {
        tracing::info!("[query_hello_world] send request");
        match client.execute(request).await {
            Ok(response) => {
                tracing::info!(
                    status = %response.status(),
                    "[query_hello_world] received response"
                )
            }
            Err(error) => tracing::error!(%error, "[query_hello_world] got error"),
        }
    }
    .instrument(span)
    .await;
}

/// loop_query_hello_world 定期循环调用 HelloWorld 服务。
pub async fn loop_query_hello_world(url: String) {
    let client = reqwest::Client::new();
    loop {
        sleep(std::time::Duration::from_secs(3)).await;
        query_hello_world(&client, &url).await;
    }
}
