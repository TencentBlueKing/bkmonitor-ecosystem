// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

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
