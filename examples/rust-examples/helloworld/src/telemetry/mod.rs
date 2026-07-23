//! Trace、Metric、Log 的 OpenTelemetry 初始化。

pub mod logs;
pub mod metrics;
pub mod setup;
pub mod traces;

pub use setup::{setup, TelemetryGuard};
