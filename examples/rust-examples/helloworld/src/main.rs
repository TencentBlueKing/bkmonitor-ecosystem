// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//! helloworld 示例应用的主入口。

use std::error::Error;

use helloworld::http::{
    client::loop_query_hello_world,
    server::{app, AppState},
};
use helloworld::{config::AppConfig, telemetry};
use tokio::net::TcpListener;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let config = AppConfig::from_env();
    let telemetry = telemetry::setup(&config)?;

    let listener = TcpListener::bind((&config.server_address[..], config.server_port)).await?;
    let client_task = tokio::spawn(loop_query_hello_world(format!(
        "http://{}:{}/helloworld",
        config.server_address, config.server_port
    )));

    println!(
        "HelloWorld server listening on http://{}:{}",
        config.server_address, config.server_port
    );
    tokio::select! {
        result = axum::serve(listener, app(AppState { error_rate: 0.1 })) => result?,
        signal = tokio::signal::ctrl_c() => signal?,
    }

    client_task.abort();
    telemetry.shutdown();
    Ok(())
}
