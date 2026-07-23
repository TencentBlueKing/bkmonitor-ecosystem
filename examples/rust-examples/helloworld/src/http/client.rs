//! 定时发起请求、为示例持续产生观测数据的内置客户端。

use opentelemetry::global;
use opentelemetry_http::HeaderInjector;
use tokio::time::sleep;
use tracing::Instrument;
use tracing_opentelemetry::OpenTelemetrySpanExt;

pub async fn query_hello_world_loop(url: String) {
    let client = reqwest::Client::new();
    loop {
        // Client 按固定 3 s 周期产生请求，为 demo 持续生成观测数据。
        sleep(std::time::Duration::from_secs(3)).await;
        let span = tracing::info_span!("Caller/queryHelloWorld", otel.kind = "client");

        let mut request = match client.get(&url).build() {
            Ok(request) => request,
            Err(error) => {
                tracing::error!(%error, "创建内置客户端请求失败");
                continue;
            }
        };
        // 将当前 Client Span 的上下文写入 traceparent 请求头，供 Server 提取。
        global::get_text_map_propagator(|propagator| {
            propagator.inject_context(&span.context(), &mut HeaderInjector(request.headers_mut()));
        });

        async {
            match client.execute(request).await {
                Ok(response) => {
                    tracing::info!(status = %response.status(), "内置客户端收到响应")
                }
                Err(error) => tracing::error!(%error, "内置客户端请求失败"),
            }
        }
        .instrument(span)
        .await;
    }
}
