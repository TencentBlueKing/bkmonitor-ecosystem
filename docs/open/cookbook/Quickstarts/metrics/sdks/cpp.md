# C++-指标（Prometheus）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Term/metrics/what.md" target="_blank">什么是指标</a>

* <a href="{{COOKBOOK_METRICS_TYPES}}" target="_blank">指标类型</a>

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/metrics/sdks/cpp
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Quickstarts/metrics/sdks/README.md" target="_blank">自定义指标 Prometheus SDK 上报</a> 创建一个上报协议为 `Prometheus` 的自定义指标，关注创建后提供的配置项：

* `TOKEN`：数据源 Token，后续需要在上报数据时使用。

同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image-1.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数                      | 类型         | 描述                                       |
|-------------------------|------------|------------------------------------------|
| `TOKEN`                 | String     | ❗❗【非常重要】 自定义指标数据源 `Token`。               |
| `API_URL`               | String     | ❗❗【非常重要】 数据上报接口地址（`Access URL`），国内站点请填写「 |127.0.0.1:4318| 」，其他环境、跨云场景请根据页面接入指引填写。 |            |                                          |
| `INTERVAL`              | Integer  　 | 数据上报间隔，默认值为 60 秒。                        |
| `METRICS_PORT`          | Integer  　 | 指标暴露端口，默认 2323。                          |

#### 2.2.1 关键配置

蓝鲸监控支持原生 Prometheus 协议，如果业务已接入 Prometheus SDK，只需在 `push_to_gateway` 方法，修改上报地址为 `API_URL`，增加注入 `X-BK-TOKEN` 的 handler。

```cpp
private:
    void initializeGateway() {
        try {
            gateway = std::make_unique<prometheus::Gateway>(
                config.api_url,         // 完整URL
                [](CURL* /*curl*/) {},  // 空的CURL配置函数
                config.job,
                config.getGroupingKey()
            );

            // ❗️❗️【非常重要】 创建使用X-BK-TOKEN认证的HTTP客户端
            if (!config.token.empty()) {
                gateway->AddHttpHeader("X-BK-TOKEN: " + config.token);
            }

            gateway->RegisterCollectable(registry);

            std::cout << "Push模式启动: " << config.api_url << std::endl;

        } catch (const std::exception& e) {
            std::cerr << "Gateway初始化失败: " << e.what() << std::endl;
        }
    }
```


### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/metrics/sdks/cpp" target="_blank">bkmonitor-ecosystem/examples/metrics/sdks/cpp</a> 中找到。

PUSH 上报（metric 服务主动上报到端点）：

```bash
docker build -t metrics-sdk-cpp .

docker run \
-e JOB="default_monitor_job" \
-e INSTANCE="127.0.0.1" \
-e API_URL="127.0.0.1:4318" \
-e TOKEN="xxx" \
-e INTERVAL=60 metrics-sdk-cpp
```

### 2.4 使用示例

#### 2.4.1 Counter

用于记录累计值（如 API 调用总量、错误次数），只能递增。 可用于统计接口请求量、错误率（结合 rate / increase 等函数计算）。

例如，可以通过以下方式上报请求总数：

```cpp

class MetricsManager {
private:
    Config config;
    std::shared_ptr<prometheus::Registry> registry;   // Prometheus指标注册表
    std::unique_ptr<PushClient> push_client;          // Pushgateway客户端
    std::unique_ptr<prometheus::Exposer> exposer;

    std::random_device rd;
    std::mt19937 gen;
    std::uniform_real_distribution<> real_dis;
    std::uniform_int_distribution<> int_dis;

    prometheus::Family<prometheus::Counter>& api_counter_family;

public:
    MetricsManager(const Config& cfg)
        : config(cfg)
        , gen(rd())
        , real_dis(5.0, 95.0)
        , int_dis(0, 9)
        , registry(std::make_shared<prometheus::Registry>())
        , api_counter_family(prometheus::BuildCounter()
            .Name("api_called_total")  // 指标名称
            .Help("API调用次数")       // 指标说明
            .Register(*registry))
private:
    void generateMetrics() {
        // ==================== Counter指标 ====================
        // Counter类型 - API调用统计
        // Refer:https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/counter.h
        std::vector<std::string> apis{"/login", "/query", "/order"};
        std::string api = apis[gen() % apis.size()];
        std::string status = (int_dis(gen) == 0) ? "500" : "200";

        auto& counter = api_counter_family.Add({{"api", api}, {"status", status}});
        counter.Increment();
        std::cout << "API: " << api << " 状态: " << status << std::endl;
    }

```

<a href="https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/counter.h" target="_blank">Prometheus C++ SDK - Counter</a>。

#### 2.4.2 Gauge

用于记录瞬时值（可任意增减），如实时资源状态、队列长度、活跃连接数等。

```cpp

class MetricsManager {
private:
    Config config;
    std::shared_ptr<prometheus::Registry> registry;   // Prometheus指标注册表
    std::unique_ptr<PushClient> push_client;          // Pushgateway客户端
    std::unique_ptr<prometheus::Exposer> exposer;

    std::random_device rd;
    std::mt19937 gen;
    std::uniform_real_distribution<> real_dis;
    std::uniform_int_distribution<> int_dis;

    prometheus::Family<prometheus::Gauge>& cpu_gauge_family;

public:
    MetricsManager(const Config& cfg)
        : config(cfg)
        , gen(rd())
        , real_dis(5.0, 95.0)
        , int_dis(0, 9)
        , registry(std::make_shared<prometheus::Registry>())
        , cpu_gauge_family(prometheus::BuildGauge()
            .Name("cpu_usage_percent")
            .Help("CPU使用率")
            .Register(*registry))

private:
    void generateMetrics() {
        // ==================== Gauge指标 ====================
        // Gauge类型 - CPU使用率
        // Refer:https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/gauge.h
        std::string host = "host" + std::to_string((gen() % 3) + 1);
        double usage = std::round(real_dis(gen) * 10) / 10.0;

        auto& gauge = cpu_gauge_family.Add({{"host", host}});
        gauge.Set(usage);
        std::cout << "CPU: " << host << " = " << usage << "%" << std::endl;
    }

```

<a href="https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/gauge.h" target="_blank">Prometheus C++ SDK - Gauge</a>。

#### 2.4.3 Histogram

用于记录数值分布情况（如任务耗时、响应大小），通过预定的桶（bucket）统计观测值落入各区间的频率，并自动生成 _sum （总和）、_count （总数）等衍生指标。适用于分析耗时分布、计算分位数（P90/P95）等场景。

```cpp

class MetricsManager {
private:
    Config config;
    std::shared_ptr<prometheus::Registry> registry;   // Prometheus指标注册表
    std::unique_ptr<PushClient> push_client;          // Pushgateway客户端
    std::unique_ptr<prometheus::Exposer> exposer;

    std::random_device rd;
    std::mt19937 gen;
    std::uniform_real_distribution<> real_dis;
    std::uniform_int_distribution<> int_dis;

    prometheus::Family<prometheus::Histogram>& task_histogram_family;
    const std::vector<double> histogram_buckets = {0.1, 0.5, 1, 2, 5};

public:
    MetricsManager(const Config& cfg)
        : config(cfg)
        , gen(rd())
        , real_dis(5.0, 95.0)
        , int_dis(0, 9)
        , registry(std::make_shared<prometheus::Registry>())
        , task_histogram_family(prometheus::BuildHistogram()
            .Name("task_duration_seconds")
            .Help("任务耗时")
            .Register(*registry))

private:
    void generateMetrics() {
        // ==================== Histogram指标 ====================
        // Histogram类型 - 请求耗时分布
        // Refer:https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/histogram.h
        std::vector<std::string> tasks{"import", "export", "process"};
        std::string task = tasks[gen() % tasks.size()];
        double duration = static_cast<double>(gen() % 600) / 100.0;

        auto& histogram = task_histogram_family.Add({{"task", task}}, histogram_buckets);
        histogram.Observe(duration);
        std::cout << "任务: " << task << " = " << duration << "s" << std::endl;
    }

```

<a href="https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/histogram.h" target="_blank">Prometheus C++ SDK - Histogram</a>。

#### 2.4.4 Summary

用于在客户端直接计算分位数（如 P95/P99 请求耗时），适用于需高精度分位数且无需跨实例聚合场景。

```cpp

class MetricsManager {
private:
    Config config;
    std::shared_ptr<prometheus::Registry> registry;   // Prometheus指标注册表
    std::unique_ptr<PushClient> push_client;          // Pushgateway客户端
    std::unique_ptr<prometheus::Exposer> exposer;

    std::random_device rd;
    std::mt19937 gen;
    std::uniform_real_distribution<> real_dis;
    std::uniform_int_distribution<> int_dis;

    prometheus::Family<prometheus::Summary>& process_summary_family;
    const prometheus::Summary::Quantiles summary_quantiles = {
        {0.5, 0.05}, {0.9, 0.01}, {0.99, 0.001}
    };

public:
    MetricsManager(const Config& cfg)
        : config(cfg)
        , gen(rd())
        , real_dis(5.0, 95.0)
        , int_dis(0, 9)
        , registry(std::make_shared<prometheus::Registry>())
        , process_summary_family(prometheus::BuildSummary()
            .Name("task_processing_seconds")
            .Help("处理时间")
            .Register(*registry))

private:
    void generateMetrics() {
        // ==================== Summary指标 ====================
        // Summary类型 - 处理时间摘要
        // Refer:https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/summary.h
        std::vector<std::string> stages{"val", "exec", "clean"};
        std::string stage = stages[gen() % stages.size()];
        double process_time = static_cast<double>(gen() % 300) / 100.0;

        auto& summary = process_summary_family.Add({{"stage", stage}}, summary_quantiles);
        summary.Observe(process_time);
        std::cout << "处理: " << stage << " = " << process_time << "s" << std::endl;
    }

```

<a href="https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/summary.h" target="_blank">Prometheus C++ SDK - Summary</a>。

### 2.5 样例代码

该样例使用 Prometheus_client 库实现四种指标类型（`Counter`、`Gauge`、`Histogram`、`Summary`）上报：

```cpp

#include <prometheus/counter.h>
#include <prometheus/gauge.h>
#include <prometheus/histogram.h>
#include <prometheus/summary.h>
#include <prometheus/exposer.h>
#include <prometheus/registry.h>
#include <prometheus/gateway.h>
#include <prometheus/text_serializer.h>

#include <chrono>
#include <iostream>
#include <memory>
#include <random>
#include <string>
#include <thread>
#include <cstdlib>
#include <sstream>
#include <map>
#include <vector>
#include <functional>

// ==================== 配置信息 ====================
class Config {
 public:
  std::string token, api_url, job, instance;
  int interval, metrics_port;

  Config() {
    // ❗️❗️【非常重要】请填写为申请到的自定义指标认证令牌（`Token`）。
    token = getEnv("TOKEN", "fixme");
    // ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写。
    api_url = getEnv("API_URL", "fixme");
    job = getEnv("JOB", "cpp_monitor");    // 任务名称
    instance = getEnv("INSTANCE", "127.0.0.1");  // 实例名称
    interval = getEnvInt("INTERVAL", 60);    // 上报间隔，默认60秒
    metrics_port = getEnvInt("METRICS_PORT", 2323);  // 默认2323端口暴露/metrics端点
  }

  std::map<std::string, std::string> getGroupingKey() const {
    return {{"job", job}, {"instance", instance}};
  }

 private:
  static std::string getEnv(const std::string& key, const std::string& defaultValue) {
    const char* value = std::getenv(key.c_str());
    return value ? value : defaultValue;
  }

  static int getEnvInt(const std::string& key, int defaultValue) {
    const char* value = std::getenv(key.c_str());
    return value ? std::stoi(value) : defaultValue;
  }
};

// ==================== 指标管理 ====================
class MetricsManager {
 private:
  Config config;
  std::shared_ptr<prometheus::Registry> registry;  // Prometheus指标注册表
  std::unique_ptr<prometheus::Gateway> gateway;    // Pushgateway客户端
  std::unique_ptr<prometheus::Exposer> exposer;

  std::random_device rd;
  std::mt19937 gen;
  std::uniform_real_distribution<> real_dis;
  std::uniform_int_distribution<> int_dis;

  prometheus::Family<prometheus::Counter>& api_counter_family;
  prometheus::Family<prometheus::Gauge>& cpu_gauge_family;
  prometheus::Family<prometheus::Histogram>& task_histogram_family;
  prometheus::Family<prometheus::Summary>& process_summary_family;

  const std::vector<double> histogram_buckets = {0.1, 0.5, 1, 2, 5};
  const prometheus::Summary::Quantiles summary_quantiles = {
    {0.5, 0.05}, {0.9, 0.01}, {0.99, 0.001}
  };

 public:
  explicit MetricsManager(const Config& cfg)
      : config(cfg),
        gen(rd()),
        real_dis(5.0, 95.0),
        int_dis(0, 9),
        registry(std::make_shared<prometheus::Registry>()),
        api_counter_family(prometheus::BuildCounter()
            .Name("api_called_total")    // 指标名称
            .Help("API调用次数")         // 指标说明
            .Register(*registry)),
        cpu_gauge_family(prometheus::BuildGauge()
            .Name("cpu_usage_percent")
            .Help("CPU使用率")
            .Register(*registry)),
        task_histogram_family(prometheus::BuildHistogram()
            .Name("task_duration_seconds")
            .Help("任务耗时")
            .Register(*registry)),
        process_summary_family(prometheus::BuildSummary()
            .Name("task_processing_seconds")
            .Help("处理时间")
            .Register(*registry)) {
    // 启动Pull模式
    std::string addr = "0.0.0.0:" + std::to_string(config.metrics_port);
    exposer = std::make_unique<prometheus::Exposer>(addr);
    exposer->RegisterCollectable(registry);
    std::cout << "Pull模式启动: http://" << addr << "/metrics" << std::endl;

    // 启动Push模式
    if (!config.api_url.empty()) {
      initializeGateway();
    }
  }

 private:
  void initializeGateway() {
    try {
      gateway = std::make_unique<prometheus::Gateway>(
          config.api_url,          // 完整URL
          [](CURL* /*curl*/) {},  // 空的CURL配置函数
          config.job,
          config.getGroupingKey());

      // ❗️❗️【非常重要】 创建使用X-BK-TOKEN认证的HTTP客户端
      if (!config.token.empty()) {
        gateway->AddHttpHeader("X-BK-TOKEN: " + config.token);
      }
      gateway->RegisterCollectable(registry);
      std::cout << "Push模式启动: " << config.api_url << std::endl;
    } catch (const std::exception& e) {
      std::cerr << "Gateway初始化失败: " << e.what() << std::endl;
    }
  }

 public:
  void run() {
    int cycle = 0;
    while (true) {
      cycle++;
      auto start = std::chrono::steady_clock::now();

      std::cout << "\n第 " << cycle << " 轮上报" << std::endl;

      // 生成指标
      generateMetrics();

      // 推送指标
      if (gateway) {
        pushWithGateway();
      }

      auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(
          std::chrono::steady_clock::now() - start).count();
      int sleep = std::max(config.interval - static_cast<int>(elapsed), 1);
      std::this_thread::sleep_for(std::chrono::seconds(sleep));
    }
  }

 private:
  void pushWithGateway() {
    try {
      int returnCode = gateway->Push();

      if (returnCode == 200) {
        std::cout << "推送成功" << std::endl;
      } else if (returnCode == 400) {
        std::cout << "推送失败: 请求错误" << std::endl;
      } else if (returnCode == 401) {
        std::cout << "推送失败: 认证失败" << std::endl;
      } else if (returnCode == 500) {
        std::cout << "推送失败: 服务器错误" << std::endl;
      } else {
        std::cout << "推送失败: 错误码 " << returnCode << std::endl;
      }
    } catch (const std::exception& e) {
      std::cerr << "推送异常: " << e.what() << std::endl;
    }
  }

  void generateMetrics() {
    // ==================== Counter指标 ====================
    // Counter类型 - API调用统计
    // Refer:https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/counter.h
    std::vector<std::string> apis{"/login", "/query", "/order"};
    std::string api = apis[gen() % apis.size()];
    std::string status = (int_dis(gen) == 0) ? "500" : "200";

    auto& counter = api_counter_family.Add({{"api", api}, {"status", status}});
    counter.Increment();
    std::cout << "API: " << api << " 状态: " << status << std::endl;
    // ==================== Gauge指标 ====================
    // Gauge类型 - CPU使用率
    // Refer:https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/gauge.h
    std::string host = "host" + std::to_string((gen() % 3) + 1);
    double usage = std::round(real_dis(gen) * 10) / 10.0;

    auto& gauge = cpu_gauge_family.Add({{"host", host}});
    gauge.Set(usage);
    std::cout << "CPU: " << host << " = " << usage << "%" << std::endl;
    // ==================== Histogram指标 ====================
    // Histogram类型 - 请求耗时分布
    // Refer:https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/histogram.h
    std::vector<std::string> tasks{"import", "export", "process"};
    std::string task = tasks[gen() % tasks.size()];
    double duration = static_cast<double>(gen() % 600) / 100.0;

    auto& histogram = task_histogram_family.Add({{"task", task}}, histogram_buckets);
    histogram.Observe(duration);
    std::cout << "任务: " << task << " = " << duration << "s" << std::endl;
    // ==================== Summary指标 ====================
    // Summary类型 - 处理时间摘要
    // Refer:https://github.com/jupp0r/prometheus-cpp/blob/master/core/include/prometheus/summary.
    std::vector<std::string> stages{"val", "exec", "clean"};
    std::string stage = stages[gen() % stages.size()];
    double process_time = static_cast<double>(gen() % 300) / 100.0;

    auto& summary = process_summary_family.Add({{"stage", stage}}, summary_quantiles);
    summary.Observe(process_time);
    std::cout << "处理: " << stage << " = " << process_time << "s" << std::endl;
  }
};

int main() {
  try {
    Config config;
    MetricsManager manager(config);
    manager.run();
  } catch (const std::exception& e) {
    std::cerr << "错误: " << e.what() << std::endl;
    return 1;
  }
  return 0;
}

```

### 2.5 PULL 模式

上文主要介绍将指标数据，**主动推送**到蓝鲸监控平台，也可以通过 HTTP 暴露指标，通过 ServiceMonitor（BCS）或采集插件的方式拉取。

样例代码同时兼容 PULL 和 PUSH，通过 `start_http_server` 在给定端口上的守护进程线程中启动 HTTP 服务器，暴露指标：

```cpp
class Config {
public:
    int interval, metrics_port;

    Config() {
        metrics_port = getEnvInt("METRICS_PORT", 2323);  //  默认2323端口暴露/metrics端点
    }
        // 启动Pull模式
        std::string addr = "0.0.0.0:" + std::to_string(config.metrics_port);
        exposer = std::make_unique<prometheus::Exposer>(addr);
        exposer->RegisterCollectable(registry);
        std::cout << "🌐 Pull模式启动: http://" << addr << "/metrics" << std::endl;
    ...
```

运行样例：

```bash
docker build -t metrics-sdk-cpp .
docker run -p 2323:2323 --name sdk-pull-cpp metrics-sdk-cpp
```

获取指标：

```bash
curl http://127.0.0.1:2323/metrics
```

得到类似输出说明启动成功：

```cpp
# HELP api_called_total API调用次数
# TYPE api_called_total counter
api_called_total{api="/login",status="200"} 1
# HELP cpu_usage_percent CPU使用率
# TYPE cpu_usage_percent gauge
cpu_usage_percent{host="host1"} 50.6
# HELP task_duration_seconds 任务耗时
# TYPE task_duration_seconds histogram
task_duration_seconds_count{task="process"} 1
task_duration_seconds_sum{task="process"} 0.8100000000000001
task_duration_seconds_bucket{task="process",le="0.1"} 0
task_duration_seconds_bucket{task="process",le="0.5"} 0
task_duration_seconds_bucket{task="process",le="1"} 1
task_duration_seconds_bucket{task="process",le="2"} 1
task_duration_seconds_bucket{task="process",le="5"} 1
task_duration_seconds_bucket{task="process",le="+Inf"} 1
# HELP task_processing_seconds 处理时间
# TYPE task_processing_seconds summary
task_processing_seconds_count{stage="clean"} 1
task_processing_seconds_sum{stage="clean"} 1.19
task_processing_seconds{stage="clean",quantile="0.5"} 1.19
task_processing_seconds{stage="clean",quantile="0.9"} 1.19
task_processing_seconds{stage="clean",quantile="0.99"} 1.19
```

## 3. 了解更多

* 进行 <a href="#" target="_blank">指标检索</a>。

* 了解 <a href="#" target="_blank">怎么使用监控指标</a>。

* 了解如何 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/dashboard.md" target="_blank">配置仪表盘</a>。

* 了解如何使用 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/alarm-configurations/rules.md" target="_blank">监控告警</a>。

* 了解 <a href="https://prometheus.github.io/client_python/" target="_blank"> Prometheus Python SDK</a>。

* 了解 <a href="https://prometheus.github.io/client_python/exporting/" target="_blank">Promethues Python SDK 指标导出方式 </a>。