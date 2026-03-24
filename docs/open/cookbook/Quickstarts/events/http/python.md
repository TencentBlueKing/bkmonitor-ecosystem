# Python-事件（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Quickstarts/events/http/README.md" target="_blank">自定义事件 HTTP 上报</a>

### 1.2 上报速率限制

默认的 API 接收频率，单个 dataid 限制 1000 次／ min，单次上报 Body 最大为 500 KB。

如超过频率限制，请联系`蓝鲸助手`调整。

### 1.3 初始化 demo

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/events/python
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Quickstarts/events/http/README.md" target="_blank">自定义事件 HTTP 上报</a> 创建自定义事件后需关注提供的两个配置项：

* `TOKEN`：自定义事件数据源 Token，上报数据时使用。

* `数据 ID`: 数据 ID（Data ID），自定义事件数据源唯一标识，上报数据时使用。

同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image-1.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数     | 类型                | 描述                         |
| ------------ | ------------------- | ---------------------------- |
|`TOKEN`       |String      |❗❗【非常重要】自定义事件数据源 `Token`。  |
|`DATA_ID`       |Integer     |❗❗【非常重要】数据 ID（`Data ID`），自定义事件数据源唯一标识。|
|`API_URL`       |String         |❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写。|
|`INTERVAL` |Integer     |上报间隔（单位为秒），默认 60 秒上报一次。​ |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/events/python" target="_blank">bkmonitor-ecosystem/examples/events/python</a> 中找到。

通过 docker build 构建名为 events-http-python 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、DATA_ID、API_URL 传递配置参数，实现周期上报事件：

```bash
docker build -t events-http-python .

docker run -e TOKEN="xxx" \
 -e DATA_ID=000000 \
 -e API_URL="http://127.0.0.1:10205/v2/push/" \
 -e INTERVAL=60 events-http-python
```

运行输出：

```shell
2026-03-23 07:57:59 | 事件上报服务启动 | 目标: 127.0.0.1 | 间隔: 60秒
2026-03-23 07:57:59 | 生成事件数据:
[
  {
    "event_name": "cpu_alert",
    "event": {
      "content": "CPU告警: 96%"
    },
    "target": "127.0.0.1",
    "dimension": {
      "module": "db",
      "location": "guangdong"
    },
    "timestamp": 1774252679629
  }
]
2026-03-23 07:57:59 | 上报结果: success 上报成功
```

### 2.4 样例代码

该样例通过模拟周期上报 CPU 及内存使用率（数值随机生成），演示如何进行自定义事件上报：

```python
import json
import os
import random
import time
from datetime import datetime

import requests

# ===== 配置层 =====
# ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
API_URL = os.getenv("API_URL", "fixme")
# ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
DATA_ID = int(os.getenv("DATA_ID", 0))
# ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
TOKEN = os.getenv("TOKEN", "fixme")
# 目标设备IP
TARGET_IP = os.getenv("TARGET_IP", "127.0.0.1")
# 上报间隔（秒）
INTERVAL = int(os.getenv("INTERVAL", 60))


# ===== 日志功能 =====
def log(msg):
    """输出带时间戳的日志"""
    print(f"\033[1m{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\033[0m | {msg}")


# ===== 核心功能 =====
def send_events():
    """构造事件数据并发送到上报接口，返回 (status, message)"""
    # 步骤1：构造事件数据
    event_data = [
        {
            "event_name": "cpu_alert",
            "event": {"content": f"CPU告警: {random.randint(80, 99)}%"},
            "target": TARGET_IP,
            "dimension": {"module": "db", "location": "guangdong"},
            "timestamp": int(time.time() * 1000),
        }
    ]
    log(f"生成事件数据:\n{json.dumps(event_data, indent=2, ensure_ascii=False)}")

    # 步骤2：构造请求负载
    # ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
    # ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
    payload = {"data_id": DATA_ID, "access_token": TOKEN, "data": event_data}

    # 步骤3：发送请求
    # ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
    resp = requests.post(API_URL, json=payload, timeout=5, headers={"Content-Type": "application/json"})

    if resp.status_code == 200:
        return "success", "上报成功"
    return "error", f"HTTP {resp.status_code}"

def main():
    log(f"事件上报服务启动 | 目标: {TARGET_IP} | 间隔: {INTERVAL}秒")
    # 持续上报，每次上报后等待 INTERVAL 秒
    while True:
        status, message = send_events()
        info_color = "\033[32m" if status == "success" else "\033[31m"
        log(f"上报结果: {info_color}{status} {message}\033[0m")
        time.sleep(INTERVAL)


# ===== 执行入口 =====
if __name__ == "__main__":
    main()
```

## 3. 了解更多

* <a href="#" target="_blank">事件数据接入</a>。

* <a href="#" target="_blank">主机事件</a>。

* <a href="#" target="_blank">容器事件</a>。