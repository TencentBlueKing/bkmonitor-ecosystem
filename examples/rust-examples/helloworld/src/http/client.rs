// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//! 定时发起请求、为示例持续产生观测数据的内置客户端。

use reqwest_middleware::{ClientBuilder, ClientWithMiddleware};
use reqwest_tracing::TracingMiddleware;
use tokio::time::sleep;

async fn query_hello_world(client: &ClientWithMiddleware, url: &str) {
    tracing::info!("[query_hello_world] send request");
    match client.get(url).send().await {
        Ok(response) => {
            tracing::info!(
                status = %response.status(),
                "[query_hello_world] received response"
            )
        }
        Err(error) => {
            tracing::error!(%error, "[query_hello_world] got error");
        }
    }
}

/// loop_query_hello_world 定期循环调用 HelloWorld 服务。
pub async fn loop_query_hello_world(url: String) {
    let client = ClientBuilder::new(reqwest::Client::new())
        .with(TracingMiddleware::default())
        .build();
    loop {
        sleep(std::time::Duration::from_secs(3)).await;
        query_hello_world(&client, &url).await;
    }
}
