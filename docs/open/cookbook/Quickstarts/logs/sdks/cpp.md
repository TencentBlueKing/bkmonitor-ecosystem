# C++-日志（SDK）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/sdks/README.md" target="_blank">自定义日志 OpenTelemetry SDK 上报</a>

* <a href="https://opentelemetry.io/docs/concepts/signals/logs/" target="_blank">Logs（OTel 日志）</a>：OpenTelemetry 中用于描述离散日志事件的信号类型。

* <a href="https://opentelemetry.io/docs/specs/otel/logs/data-model/" target="_blank">Logs Data Model（OTel 日志数据模型）</a>：定义 `resourceLogs`、`scopeLogs`、`logRecords`、`body`、`attributes`、`severityNumber` 等字段含义。

* <a href="https://opentelemetry.io/docs/specs/otlp/#otlphttp" target="_blank">OTLP/HTTP（OpenTelemetry HTTP 上报协议）</a>：定义通过 HTTP 上报 OTel 数据的协议方式，C++ SDK 通过 `/v1/logs` 上报日志。

### 1.2 接入样例获取

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/logs/sdks/cpp
```

- *[1] Git 用于获取样例，Docker（或兼容容器工具）用于构建和运行容器样例，仅供快速验证。*
- *[2] 数据上报不依赖 Git 或 Docker，业务可参考后文代码或上报协议，自行组织上报方式。*

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/cookbook/Quickstarts/logs/sdks/README.md" target="_blank">自定义日志 OpenTelemetry SDK 上报</a> 创建一个自定义日志数据源，关注创建后提供的两个配置项：

* `TOKEN`：日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。

* `API_URL`：国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，其他环境、跨云场景请根据页面接入指引填写。

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 配置 | 必填 | 说明 |
| --- | --- | --- |
| `API_URL` | 是 | ❗❗【非常重要】日志上报接口地址（`Access URL`），请根据页面接入指引填写；如果页面提供的是 OTLP HTTP Endpoint 根地址，请在末尾追加 `/v1/logs`。 |
| `TOKEN` | 是 | ❗❗【非常重要】日志数据源 Token，上报时必须通过 `x-bk-token` Header 传递。 |

#### 2.2.1 关键配置

使用 OpenTelemetry C++ SDK（v1.18.0）进行日志上报时，需正确配置上报地址、鉴权 Header 和日志导出链路：

```cpp
#include <cstdlib>
#include <memory>
#include <string>

#include "opentelemetry/logs/provider.h"
#include "opentelemetry/sdk/logs/logger_provider.h"
#include "opentelemetry/sdk/logs/batch_log_record_processor.h"
#include "opentelemetry/sdk/resource/resource.h"
#include "opentelemetry/exporters/otlp/otlp_http_log_record_exporter.h"
#include "opentelemetry/exporters/otlp/otlp_http_log_record_exporter_options.h"

namespace logs_api      = opentelemetry::logs;
namespace logs_sdk      = opentelemetry::sdk::logs;
namespace otlp_exporter = opentelemetry::exporter::otlp;
namespace nostd         = opentelemetry::nostd;

std::string GetEnv(const char* key, const std::string& fallback) {
    const char* val = std::getenv(key);
    return val ? std::string(val) : fallback;
}

// ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
std::string token = GetEnv("TOKEN", "fixme");
// ❗❗【非常重要】上报地址，国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，
// 其他环境、跨云场景请根据页面接入指引填写。
std::string api_url = GetEnv("API_URL", "http://127.0.0.1:4318/v1/logs");

otlp_exporter::OtlpHttpLogRecordExporterOptions opts;
opts.url          = api_url;
opts.content_type = otlp_exporter::HttpRequestContentType::kBinary;
// ❗❗【非常重要】必须通过 x-bk-token Header 传递 Token
opts.http_headers.insert({"x-bk-token", token});

auto exporter = std::unique_ptr<logs_sdk::LogRecordExporter>(
    new otlp_exporter::OtlpHttpLogRecordExporter(opts));
auto processor = std::unique_ptr<logs_sdk::LogRecordProcessor>(
    new logs_sdk::BatchLogRecordProcessor(std::move(exporter)));

// 声明资源属性，用于在平台侧区分日志来源
auto resource = opentelemetry::sdk::resource::Resource::Create({
    {"service.name", "custom-log-sdk-demo"},
    {"deployment.environment.name", "local"},
});

auto provider = nostd::shared_ptr<logs_sdk::LoggerProvider>(
    new logs_sdk::LoggerProvider(std::move(processor), resource));
logs_api::Provider::SetLoggerProvider(std::move(provider));
```

* <a href="https://github.com/open-telemetry/opentelemetry-cpp/tree/v1.18.0/exporters/otlp" target="_blank">OpenTelemetry C++ OTLP Exporter</a>

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/master/examples/logs/sdks/cpp" target="_blank">bkmonitor-ecosystem/examples/logs/sdks/cpp</a> 中找到。

通过 docker build 构建名为 logs-sdk-cpp 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、API_URL 传递配置参数，实现周期上报日志：

```bash
docker build -t logs-sdk-cpp .

docker run -e TOKEN="fixme" \
 -e API_URL="http://127.0.0.1:4318/v1/logs" \
 logs-sdk-cpp
```

> 首次构建会从 GitHub 拉取 `opentelemetry-cpp` v1.18.0 源码并本地编译，耗时较长，请耐心等待。

> 运行容器前，请确认容器网络可访问 `API_URL` 对应的上报地址，且 `TOKEN` 与页面创建数据源时生成的值一致。

运行输出：

```bash
2026-11-23 10:41:52,588 - INFO - Starting log reporter (press Ctrl+C/SIGTERM to stop)...
2026-11-23 10:41:52,614 - INFO - log from cpp sdk
2026-11-23 10:41:53,614 - WARN - log from cpp sdk
2026-11-23 10:41:54,614 - ERROR - log from cpp sdk
2026-11-23 10:41:55,614 - DEBUG - log from cpp sdk
...
```

> 日志通过 OTLP/HTTP 上报到平台，同时标准输出会同步打印启动信息、每条日志及退出提示；上报的日志可在日志检索页面查看。

### 2.4 样例代码

该样例使用 OpenTelemetry C++ SDK（`opentelemetry-cpp` v1.18.0），通过 OTLP/HTTP（Protobuf）协议上报日志。代码通过 `OtlpHttpLogRecordExporter` 创建日志导出器，并在 `x-bk-token` Header 中注入 Token；通过 `Resource` 声明服务名与环境，使用 `BatchLogRecordProcessor` 批量处理日志；使用 `LoggerProvider` 创建 Logger 后，手动构造 `LogRecord` 并设置时间、级别与内容，最后调用 `EmitLogRecord` 完成上报。退出时调用 `Shutdown` 刷新并释放资源。

```cpp
#include <atomic>
#include <iostream>
#include <string>
#include <csignal>
#include <cstdlib>
#include <memory>
#include <chrono>
#include <thread>
#include <ctime>

#include "opentelemetry/logs/provider.h"
#include "opentelemetry/logs/severity.h"
#include "opentelemetry/sdk/logs/logger_provider.h"
#include "opentelemetry/sdk/logs/batch_log_record_processor.h"
#include "opentelemetry/sdk/resource/resource.h"
#include "opentelemetry/exporters/otlp/otlp_http_log_record_exporter.h"
#include "opentelemetry/exporters/otlp/otlp_http_log_record_exporter_options.h"

namespace logs_api      = opentelemetry::logs;
namespace logs_sdk      = opentelemetry::sdk::logs;
namespace otlp_exporter = opentelemetry::exporter::otlp;
namespace nostd         = opentelemetry::nostd;

std::atomic<bool> running{true};

void SignalHandler(int) { running = false; }

struct LogLevel { logs_api::Severity severity; const char* text; };
const LogLevel kLevels[] = {
    {logs_api::Severity::kInfo, "INFO"},  {logs_api::Severity::kWarn, "WARN"},
    {logs_api::Severity::kError, "ERROR"}, {logs_api::Severity::kDebug, "DEBUG"},
};

std::string CurrentTimeStr() {
    auto now = std::chrono::system_clock::now();
    auto tt = std::chrono::system_clock::to_time_t(now);
    auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(now.time_since_epoch()) % 1000;
    char buf[64];
    std::strftime(buf, sizeof(buf), "%Y-%m-%d %H:%M:%S", std::localtime(&tt));
    return std::string(buf) + "," + std::to_string(ms.count());
}

std::string GetEnv(const char* key, const std::string& fallback) {
    const char* val = std::getenv(key);
    return val ? std::string(val) : fallback;
}

int main() {
    std::signal(SIGTERM, SignalHandler);
    std::signal(SIGINT,  SignalHandler);

    // ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
    std::string token   = GetEnv("TOKEN", "fixme");
    // ❗❗【非常重要】上报地址，国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，
    // 其他环境、跨云场景请根据页面接入指引填写。
    std::string api_url = GetEnv("API_URL", "http://127.0.0.1:4318/v1/logs");

    // 配置 OTLP HTTP 日志导出器
    otlp_exporter::OtlpHttpLogRecordExporterOptions opts;
    opts.url            = api_url;
    opts.content_type   = otlp_exporter::HttpRequestContentType::kBinary;
    opts.http_headers.insert({"x-bk-token", token});

    auto exporter = std::unique_ptr<logs_sdk::LogRecordExporter>(
        new otlp_exporter::OtlpHttpLogRecordExporter(opts));

    // 创建批量处理器
    auto processor = std::unique_ptr<logs_sdk::LogRecordProcessor>(
        new logs_sdk::BatchLogRecordProcessor(std::move(exporter)));

    // 声明资源属性，用于在平台侧区分日志来源
    auto resource = opentelemetry::sdk::resource::Resource::Create({
        {"service.name", "custom-log-sdk-demo"},
        {"deployment.environment.name", "local"},
    });

    // 创建 LoggerProvider
    auto provider = nostd::shared_ptr<logs_sdk::LoggerProvider>(
        new logs_sdk::LoggerProvider(std::move(processor), resource));

    nostd::shared_ptr<logs_sdk::LoggerProvider> g_provider = provider;

    // 设置为全局 LoggerProvider（必须使用 std::move 转为右值）
    logs_api::Provider::SetLoggerProvider(std::move(provider));

    // logger_name, library_name, library_version
    auto logger = g_provider->GetLogger("otel", "otel", "1.0.0");

    std::cout << CurrentTimeStr() << " - INFO - Starting log reporter (press Ctrl+C/SIGTERM to stop)..." << std::endl;

    int count = 0;
    while (running) {
        auto record = logger->CreateLogRecord();
        const auto& level = kLevels[count % 4];

        // 显式设置事件发生时间，避免上报的时间戳为 0
        auto now = std::chrono::system_clock::now();
        record->SetTimestamp(now);
        record->SetSeverity(level.severity);
        record->SetBody("log from cpp sdk");
        logger->EmitLogRecord(std::move(record));

        std::cout << CurrentTimeStr() << " - " << level.text << " - log from cpp sdk" << std::endl;

        count++;
        std::this_thread::sleep_for(std::chrono::seconds(1));
    }

    std::cout << CurrentTimeStr() << " - INFO - shutting down..." << std::endl;
    g_provider->Shutdown();
    return 0;
}
```

* <a href="https://github.com/open-telemetry/opentelemetry-cpp/tree/v1.18.0" target="_blank">OpenTelemetry C++ SDK</a>

* <a href="https://opentelemetry.io/docs/specs/otel/logs/api/" target="_blank">OpenTelemetry Logs API 规范</a>

## 3. 了解更多

进一步了解以下内容：

* 进行 <a href="#" target="_blank">日志检索</a>。

* 了解 <a href="#" target="_blank">容器日志自定义上报使用文档</a>。

* 了解 <a href="#" target="_blank">容器日志采集器安装</a>。