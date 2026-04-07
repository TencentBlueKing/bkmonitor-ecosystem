# go-指标（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Term/metrics/what.md" target="_blank">什么是指标</a>

* <a href="{{COOKBOOK_METRICS_TYPES}}" target="_blank">指标类型</a>

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/metrics/http/go
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/http/README.md" target="_blank">自定义指标 HTTP 上报</a> 创建一个上报协议为 `JSON` 的自定义指标，关注创建后提供的两个配置项：

* `TOKEN`：自定义指标数据源 Token，上报数据时使用。

* `数据 ID`: 数据 ID（Data ID），自定义指标数据源唯一标识，上报数据时使用。

同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数         | 类型      | 描述                                                                                                 |
|------------|---------|----------------------------------------------------------------------------------------------------|
| `TOKEN`    | String  | ❗❗【非常重要】 自定义指标数据源 `Token`。                                                                              |
| `DATA_ID`  | Integer | ❗❗【非常重要】 数据 ID（`Data ID`），自定义指标数据源唯一标识。                                                         |
| `API_URL`  | String  | ❗❗【非常重要】 数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写。 |
| `INTERVAL` | Integer | 数据上报间隔，默认值为 60 秒。    ​​                                                             |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/metrics/http/go" target="_blank">bkmonitor-ecosystem/examples/metrics/http/go</a> 中找到。

通过 docker build 构建名为 metrics-http-go 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、DATA_ID、API_URL 传递配置参数，实现周期上报指标：

```bash
docker build -t metrics-http-go .

docker run -e TOKEN="xxx" \
 -e DATA_ID=000000 \
 -e API_URL="http://127.0.0.1:10205/v2/push/" \
 -e INTERVAL=60 metrics-http-go
```

运行输出：

```bash
🚀 启动指标上报服务
API地址: http://127.0.0.1:10205/v2/push/
数据ID: 000000
上报间隔: 60秒
=================================
[2025-11-04 12:51:29] ✅ 上报成功 | CPU: 1.57% 内存: 18.61%
[2025-11-04 12:52:30] ✅ 上报成功 | CPU: 0.94% 内存: 18.63%
[2025-11-04 12:53:31] ✅ 上报成功 | CPU: 1.10% 内存: 18.61%
...
```

### 2.4 样例代码

该样例通过模拟周期上报 CPU 及内存使用率（数值随机生成），演示如何进行自定义指标上报：

```go
package main

import (
        "bytes"
        "encoding/json"
        "fmt"
        "math/rand"
        "net/http"
        "os"
        "strconv"
        "time"
)

// Payload 定义了自定义指标上报的数据结构
type Payload struct {
        // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
        DataID      int        `json:"data_id"`
        // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
        AccessToken string     `json:"access_token"`
        Data        []DataItem `json:"data"`
}

// DataItem 定义了单条指标数据项的结构
type DataItem struct {
        Metrics   Metrics   `json:"metrics"`
        Target    string    `json:"target"`
        Dimension Dimension `json:"dimension"`
}

// Metrics 定义了监控指标的数据结构
type Metrics struct {
        CpuLoad     float64 `json:"cpu_load"`
        MemoryUsage float64 `json:"memory_usage"`
}

// Dimension 定义了指标数据的维度信息结构
type Dimension struct {
        Module string `json:"module"`
        Region string `json:"region"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func sendReport(apiURL, token string, dataID int, cpuLoad, memUsage float64) int {
        payload := Payload{
                DataID: dataID, AccessToken: token,
                Data: []DataItem{
                        {
                                Metrics:   Metrics{CpuLoad: cpuLoad, MemoryUsage: memUsage},
                                Target:    "127.0.0.1",
                                Dimension: Dimension{Module: "server", Region: "guangdong"},
                        },
                },
        }
        jsonData, _ := json.Marshal(payload)
        req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")
        resp, err := httpClient.Do(req)
        if err != nil {
                fmt.Printf("⚠️ 请求异常: %v\n", err)
                return 500
        }
        defer resp.Body.Close()
        return resp.StatusCode
}

func main() {
        // ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写
        apiURL := os.Getenv("API_URL")
        // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
        token := os.Getenv("TOKEN")
        // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
        dataIDStr := os.Getenv("DATA_ID")
        dataID, _ := strconv.Atoi(dataIDStr)
        // 上报间隔，默认为60秒
        interval := 60
        if v := os.Getenv("INTERVAL"); v != "" {
                interval, _ = strconv.Atoi(v)
        }

        fmt.Println("🚀 启动指标上报服务")
        fmt.Println("API地址:", apiURL)
        fmt.Println("数据ID:", dataID)
        fmt.Printf("上报间隔: %d秒\n", interval)
        fmt.Println("=================================")
        for {
                cpuLoad := rand.Float64() * 100
                memUsage := rand.Float64() * 100
                status := sendReport(apiURL, token, dataID, cpuLoad, memUsage)
                timestamp := time.Now().Format("2006-01-02 15:04:05")
                if status == 200 {
                        fmt.Printf("[%s] ✅ 上报成功 | CPU: %.2f%% 内存: %.2f%%\n",
                                timestamp, cpuLoad, memUsage)
                } else {
                        fmt.Printf("[%s] ❌ 上报失败 | 状态码: %d\n", timestamp, status)
                }
                time.Sleep(time.Duration(interval) * time.Second)
        }
}
```

## 3. 了解更多

* 进行 <a href="#" target="_blank">指标检索</a>。

* 了解 <a href="#" target="_blank">怎么使用监控指标</a>。

* 了解如何 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/data-visualization/dashboard.md" target="_blank">配置仪表盘</a>。

* 了解如何使用 <a href="https://bk.tencent.com/docs/markdown/ZH/Monitor/3.9/UserGuide/ProductFeatures/alarm-configurations/rules.md" target="_blank">监控告警</a>。