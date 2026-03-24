# Python-指标（HTTP）上报

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
cd bkmonitor-ecosystem/examples/metrics/http/python
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Quickstarts/metrics/http/README.md" target="_blank">自定义指标 HTTP 上报</a> 创建一个上报协议为 `JSON` 的自定义指标，关注创建后提供的两个配置项：

* `TOKEN`：自定义指标数据源 Token，上报数据时使用。

* `数据 ID`: 数据 ID（Data ID），自定义指标数据源唯一标识，上报数据时使用。

同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数         | 类型      | 描述                                                                                                 |
|------------|---------|----------------------------------------------------------------------------------------------------|
| `TOKEN`    | String  | ❗❗【非常重要】 自定义指标数据源 `Token`。                                                                               |
| `DATA_ID`  | Integer | ❗❗【非常重要】 数据 ID（`Data ID`），自定义指标数据源唯一标识。                                                                         |
| `API_URL`  | String  | ❗❗【非常重要】 数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写。 |
| `INTERVAL` | Integer | 数据上报间隔，默认值为 60 秒。    ​​                                                             |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/metrics/http/python" target="_blank">bkmonitor-ecosystem/examples/metrics/http/python</a> 中找到。

通过 docker build 构建名为 metrics-http-python 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、DATA_ID、API_URL 传递配置参数，实现周期上报指标：

```bash
docker build -t metrics-http-python .

docker run -e TOKEN="xxx" \
 -e DATA_ID=00000 \
 -e API_URL="http://127.0.0.1:10205/v2/push/" \
 -e INTERVAL=60 metrics-http-python
```

运行输出：

```bash
[2025-07-04 09:16:21] ✅ 上报成功 | CPU: 4.91% 内存: 30.13%
[2025-07-04 09:17:22] ✅ 上报成功 | CPU: 17.17% 内存: 30.51%
[2025-07-04 09:18:24] ✅ 上报成功 | CPU: 17.12% 内存: 30.15%
[2025-07-04 09:19:25] ✅ 上报成功 | CPU: 2.41% 内存: 30.12%
...
```

### 2.4 样例代码

该样例通过模拟周期上报 CPU 及内存使用率（数值随机生成），演示如何进行自定义指标上报：

```python
# -*- coding: utf-8 -*-
import json
import argparse
import time
import requests
import random
import os


def collect_metrics():
    """采集模拟周期上报 CPU 及内存使用率（数值随机生成）"""
    cpu_load = random.uniform(0, 100)
    mem_usage =  random.uniform(0, 100)
    return cpu_load, mem_usage


def send_report(api_url, token, data_id, metrics):
    """发送数据到HTTP接口"""
    payload = {
        "data_id": int(data_id),  # ❗❗【非常重要】必须是整数类型，否则上报会失败
        "access_token": token,
        "data": [
            {
                "metrics": {"cpu_load": metrics[0], "memory_usage": metrics[1]},
                "target": "127.0.0.1",
                "dimension": {"module": "server", "region": "guangdong"},
            }
        ],
    }
    try:
        response = requests.post(api_url, json=payload, timeout=10, headers={"Content-Type": "application/json"})
        return response.status_code
    except requests.exceptions.RequestException as e:
        print(f"⚠️ 请求异常: {e}")
        return 500


def main():
    # ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写
    api_url = os.getenv("API_URL", "")
    token = os.getenv("TOKEN", "")  # ❗❗【非常重要】access_token:认证令牌，用于接口鉴定，配置为应用 TOKEN
    data_id = os.getenv("DATA_ID", "")  # ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
    interval = int(os.getenv("INTERVAL", "60"))  # 默认间隔60秒
    while True:
        metrics = collect_metrics()
        status = send_report(api_url, token, data_id, metrics)

        timestamp = time.strftime("%Y-%m-%d %H:%M:%S")
        if status == 200:
            print(f"[{timestamp}] ✅ 上报成功 | CPU: {metrics[0]:.2f}% 内存: {metrics[1]:.2f}%")
        else:
            print(f"[{timestamp}] ❌ 上报失败 | 状态码: {status}")

        time.sleep(interval)


if __name__ == "__main__":
    main()
```

## 3. 了解更多

* 进行 <a href="#" target="_blank">指标检索</a>。

* 了解 <a href="#" target="_blank">怎么使用监控指标</a>。

* 了解如何 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/dashboard.md" target="_blank">配置仪表盘</a>。

* 了解如何使用 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/alarm-configurations/rules.md" target="_blank">监控告警</a>。