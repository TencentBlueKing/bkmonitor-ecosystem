use std::error::Error;

use helloworld::http::{
    client::query_hello_world_loop,
    server::{app, AppState},
};
use helloworld::{
    config::{load_dev_env, AppConfig},
    telemetry,
};
use tokio::net::TcpListener;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    if let Some(path) = load_dev_env()? {
        eprintln!("loaded local configuration from {}", path.display());
    }
    // 在服务接受请求前初始化 OTel，确保首个请求也能产生完整的 Trace、Metric 和 Log。
    let config = AppConfig::from_env();
    let telemetry = telemetry::setup(&config)?;

    let listener = TcpListener::bind((&config.server_address[..], config.server_port)).await?;
    let client_task = tokio::spawn(query_hello_world_loop(format!(
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

    // 先停止产生业务数据的后台 Client，再刷新三个 Provider 的缓冲队列。
    client_task.abort();
    telemetry.shutdown();
    Ok(())
}
