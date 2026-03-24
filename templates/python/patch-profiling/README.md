# Profiling-Python（Datadog SDK）接入

本指南将帮助您使用 Datadog SDK 接入蓝鲸应用性能监控，以入门项目-Profiling 为例，介绍性能分析数据接入及 SDK 使用场景。

## 1. 前置准备

### 1.1 术语介绍

{{TERM_INTRO}}

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：
* Git
* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/python-examples/patch-profiling
docker build -t patch-profiling-python:latest .
```

## 2. 快速体验

### 2.1 运行样例

{{PROFILING_RUN_PARAMETERS}}

复制以下命令参数在你的终端运行：

```shell
docker run -e TOKEN="{{access_config.token}}" \
-e SERVICE_NAME="{{service_name}}" \
-e PROFILING_ENDPOINT="{{access_config.profiling.endpoint}}" \
-e ENABLE_PROFILING="{{access_config.profiling.enabled}}" \
-e ENABLE_MEMORY_PROFILING="{{access_config.profiling.enable_memory_profiling}}" patch-profiling-python:latest
```
* 样例已设置定时请求以产生监控数据，如需本地访问调试，可增加运行参数 `-p {本地端口}:8080`。

### 2.2 查看数据

等待片刻，便可在「服务详情-Profiling」看到应用数据。

![](image/img.png)

## 3. 快速接入

### 3.1 Datadog SDK

示例项目提供集成 Datadog Python SDK 并通过自定义导出器，将性能数据发送到 bk-collector 的方式。
这是通过 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/python-examples/patch-profiling/src/patch.py" target="_blank">patch.py</a> 模块中的 patch_ddtrace_to_pyroscope 补丁实现的。其他项目如有同样需求，可以简单复制该模块来使用补丁。
使用方法可以参考 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/python-examples/patch-profiling/src/main.py" target="_blank">main.py</a>:

```python
from ddtrace.profiling.profiler import Profiler

from patch import patch_ddtrace_to_pyroscope


patch_ddtrace_to_pyroscope(
    # ❗❗【非常重要】应用服务唯一标识
    service_name=config.service_name,
    # ❗❗【非常重要】请传入应用 Token
    token=config.token,
    # ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
    endpoint=config.profiling_endpoint,
    enable_memory_profiling=config.enable_memory_profiling,
)
prof = Profiler()
prof.start()
```

参考官方文档以获得更多信息：<a href="https://docs.datadoghq.com/profiler/enabling/python/" target="_blank">Enabling the Python Profiler</a>

## 4. 了解更多

{{LEARN_MORE}}
