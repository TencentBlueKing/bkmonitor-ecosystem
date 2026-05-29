# bkmonitor-ecosystem cookbook

## 📣 简介

`cookbook` 汇总 APM（应用性能监控）之外的接入和上报知识，聚焦自定义事件、自定义指标和相关术语。

## 📦 开箱即用

> 我们以语言为维度，汇总 cookbook 场景中的文档和样例。阅读时可按协议和上报方式继续下钻。

| 目录 | 文档 | 涉及场景 | 源码 |
| --- | --- | --- | --- |
| 场景总览 | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/events/http/README.md" target="_blank">自定义事件上报</a> | `Events（事件）` | |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/http/README.md" target="_blank">自定义指标 HTTP 上报</a> | `Metrics（指标）` | |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/sdks/README.md" target="_blank">自定义指标 Prometheus SDK 上报</a> | `Metrics（指标）` | |
| 术语介绍 | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Term/metrics/what.md" target="_blank">什么是指标</a> | `Metrics（指标）` | |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Term/metrics/type.md" target="_blank">指标类型</a> | `Metrics（指标）` | |
| Shell | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/events/http/curl.md" target="_blank">命令行-事件（HTTP）上报</a> | `Events（事件）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/events/curl" target="_blank">events-curl</a> |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/events/http/bkmonitorbeat.md" target="_blank">命令行-事件（bkmonitorbeat）上报</a> | `Events（事件）` | |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/http/curl.md" target="_blank">命令行-指标（HTTP）上报</a> | `Metrics（指标）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/metrics/http/curl" target="_blank">metrics-http-curl</a> |
| GO | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/events/http/go.md" target="_blank">Go-事件（HTTP）上报</a> | `Events（事件）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/events/go" target="_blank">events-go</a> |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/http/go.md" target="_blank">Go-指标（HTTP）上报</a> | `Metrics（指标）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/metrics/http/go" target="_blank">metrics-http-go</a> |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/sdks/go.md" target="_blank">Go-指标（Prometheus SDK）上报</a> | `Metrics（指标）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/metrics/sdks/go" target="_blank">metrics-sdk-go</a> |
| Python | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/events/http/python.md" target="_blank">Python-事件（HTTP）上报</a> | `Events（事件）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/events/python" target="_blank">events-python</a> |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/http/python.md" target="_blank">Python-指标（HTTP）上报</a> | `Metrics（指标）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/metrics/http/python" target="_blank">metrics-http-python</a> |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/sdks/python.md" target="_blank">Python-指标（Prometheus SDK）上报</a> | `Metrics（指标）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/metrics/sdks/python" target="_blank">metrics-sdk-python</a> |
| Java | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/events/http/java.md" target="_blank">Java-事件（HTTP）上报</a> | `Events（事件）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/events/java" target="_blank">events-java</a> |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/http/java.md" target="_blank">Java-指标（HTTP）上报</a> | `Metrics（指标）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/metrics/http/java" target="_blank">metrics-http-java</a> |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/sdks/java.md" target="_blank">Java-指标（Prometheus SDK）上报</a> | `Metrics（指标）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/metrics/sdks/java" target="_blank">metrics-sdk-java</a> |
| C++ | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/events/http/cpp.md" target="_blank">C++-事件（HTTP）上报</a> | `Events（事件）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/events/cpp" target="_blank">events-cpp</a> |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/http/cpp.md" target="_blank">C++-指标（HTTP）上报</a> | `Metrics（指标）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/metrics/http/cpp" target="_blank">metrics-http-cpp</a> |
| | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/blob/master/docs/open/cookbook/Quickstarts/metrics/sdks/cpp.md" target="_blank">C++-指标（Prometheus SDK）上报</a> | `Metrics（指标）` | <a href="{{ECOSYSTEM_REPOSITORY_URL}}/tree/master/examples/metrics/sdks/cpp" target="_blank">metrics-sdk-cpp</a> |
