//! 从环境变量读取运行配置。

/// 示例程序与 exporter 的运行配置。
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
    pub fn from_env() -> Self {
        Self {
            service_name: std::env::var("SERVICE_NAME").unwrap_or_else(|_| "helloworld".to_owned()),
            // ❗❗【非常重要】请根据 APM 应用接入指引设置 TOKEN，不能提交真实 Token。
            token: std::env::var("TOKEN").unwrap_or_else(|_| "todo".to_owned()),
            // ❗❗【非常重要】OTLP HTTP 上报地址必须不带协议前缀。
            // 例如：collector.example.com:4318；程序会统一补充 http://。
            otlp_endpoint: normalize_otlp_endpoint(
                &std::env::var("OTLP_ENDPOINT").unwrap_or_else(|_| "127.0.0.1:4318".to_owned()),
            ),
            // Rust 示例固定使用推荐的 HTTP/protobuf。
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
            // HTTP 服务固定监听 8080，方便直接访问示例接口。
            server_port: 8080,
        }
    }
}

/// 将无协议 OTLP HTTP 地址转换为 exporter 可使用的完整 URL。
///
/// 环境变量必须不带协议前缀，避免同一配置出现不同传输方案。
fn normalize_otlp_endpoint(endpoint: &str) -> String {
    assert!(
        !endpoint.starts_with("http://") && !endpoint.starts_with("https://"),
        "OTLP_ENDPOINT must not include a protocol; provide host[:port] only"
    );
    format!("http://{endpoint}")
}

#[cfg(test)]
mod tests {
    use super::normalize_otlp_endpoint;

    #[test]
    #[should_panic(expected = "OTLP_ENDPOINT must not include a protocol")]
    fn rejects_endpoint_with_protocol() {
        normalize_otlp_endpoint("http://collector.example.com:4318");
    }

    #[test]
    fn adds_http_protocol_to_endpoint_without_protocol() {
        assert_eq!(
            normalize_otlp_endpoint("collector.example.com:4318"),
            "http://collector.example.com:4318"
        );
    }
}

/// 从开发环境文件加载变量，且不覆盖已有进程环境变量。
///
/// `ENV_FILE` 可指定显式环境文件；未指定时，crate 根目录缺少 `dev.env` 也属于合法情况。
pub fn load_dev_env() -> Result<Option<std::path::PathBuf>, Box<dyn std::error::Error>> {
    let explicit_path = std::env::var_os("ENV_FILE");
    let path = explicit_path
        .as_ref()
        .map(std::path::PathBuf::from)
        .unwrap_or_else(|| std::path::PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("dev.env"));

    if !path.is_file() && explicit_path.is_none() {
        return Ok(None);
    }

    dotenvy::from_path(&path)?;
    Ok(Some(path))
}
