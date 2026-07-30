// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//! helloworld 示例应用的主入口。

use std::error::Error;

use actix_web::{App, HttpServer};
use helloworld::http::{client::loop_query_hello_world, server::configure_routes};
use helloworld::{config::AppConfig, telemetry};
use opentelemetry_instrumentation_actix_web::{RequestMetrics, RequestTracing};

#[cfg(unix)]
async fn shutdown_signal() -> std::io::Result<()> {
    use tokio::signal::unix::{signal, SignalKind};

    let mut sigterm = signal(SignalKind::terminate())?;
    tokio::select! {
        result = tokio::signal::ctrl_c() => result,
        _ = sigterm.recv() => Ok(()),
    }
}

#[cfg(not(unix))]
async fn shutdown_signal() -> std::io::Result<()> {
    tokio::signal::ctrl_c().await
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let config = AppConfig::from_env();
    let telemetry = telemetry::setup(&config)?;

    let server = HttpServer::new(|| {
        App::new()
            .wrap(RequestTracing::new())
            .wrap(RequestMetrics::default())
            .configure(configure_routes)
    })
    .bind((&config.server_address[..], config.server_port))?
    .disable_signals()
    .run();
    let server_handle = server.handle();
    let client_task = tokio::spawn(loop_query_hello_world(format!(
        "http://{}:{}/helloworld",
        config.server_address, config.server_port
    )));

    println!(
        "HelloWorld server listening on http://{}:{}",
        config.server_address, config.server_port
    );
    let server_result = tokio::select! {
        result = server => result,
        signal_result = shutdown_signal() => {
            server_handle.stop(true).await;
            signal_result
        },
    };

    client_task.abort();
    let _ = client_task.await;
    telemetry.shutdown();
    server_result?;
    Ok(())
}
