//! 调用链埋点。
//!
//! 每个函数只负责创建当前业务步骤所需的 Span、属性、事件或链接；实际导出由
//! 初始化阶段注册的 OpenTelemetry tracing layer 和批处理导出器负责。

use opentelemetry::{trace::TraceContextExt, KeyValue};
use tracing_opentelemetry::OpenTelemetrySpanExt;

/// 创建描述内部操作的子 Span，并写入业务属性。
pub fn traces_custom_span_demo() {
    let span = tracing::info_span!("CustomSpanDemo/doSomething");
    span.set_attribute("custom_key", "custom_value");
    let _entered = span.enter();
    tracing::info!("custom span work completed");
}

/// 为当前请求 Span 设置业务属性。
pub fn traces_set_custom_span_attributes() {
    let span = tracing::Span::current();
    span.set_attribute("api_name", "ApiRequest");
    span.set_attribute("act_id", 12345_i64);
    tracing::info!("set custom span attributes");
}

/// 为处理过程记录开始和结束事件。
pub fn traces_span_event_demo() {
    let span = tracing::info_span!("SpanEventDemo/doSomething");
    let _entered = span.enter();
    span.add_event("Before doSomething", vec![]);
    span.add_event("After doSomething", vec![]);
}

/// 使用 Span Link 关联异步操作与当前请求。
pub fn traces_span_links_demo() {
    let async_span = tracing::info_span!("SpanLinkDemo/asyncCaller");
    let async_context = async_span.context();
    tracing::Span::current().add_link_with_attributes(
        async_context.span().span_context().clone(),
        vec![KeyValue::new("link_type", "async")],
    );
    tracing::info!("SpanLinkDemo async caller");
}
