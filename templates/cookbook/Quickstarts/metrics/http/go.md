# go-指标（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="{{docs.metrics.What}}" target="_blank">什么是指标</a>

* <a href="{{docs.metrics.Types}}" target="_blank">指标类型</a>

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

* Docker 或者其他平替的容器工具。

### 1.3 初始化 demo

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/metrics/http/go
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="{{docs.metrics.http.readme.metrics_http_readme}}" target="_blank">自定义指标 HTTP 上报</a> 创建一个上报协议为 `JSON` 的自定义指标，关注创建后提供的两个配置项：

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
| `API_URL`  | String  | ❗❗【非常重要】 数据上报接口地址（`Access URL`），国内站点请填写「 {{access_config.custom.http}} 」，其他环境、跨云场景请根据页面接入指引填写。 |
| `INTERVAL` | Integer | 数据上报间隔，默认值为 60 秒。    ​​                                                             |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/metrics/http/go" target="_blank">bkmonitor-ecosystem/examples/metrics/http/go</a> 中找到。

通过 docker build 构建名为 metrics-http-go 的镜像，并使用 docker run 运行容器，同时通过环境变量 TOKEN、DATA_ID、API_URL 传递配置参数，实现周期上报指标：

```bash
docker build -t metrics-http-go .

docker run -e TOKEN="{{access_config.token}}" \
 -e DATA_ID=000000 \
 -e API_URL="{{access_config.custom.http}}" \
 -e INTERVAL=60 metrics-http-go
```

运行输出：

```bash
🚀 开始自定义指标上报，间隔: 60秒
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

type Payload struct {
        DataID      int        `json:"data_id"`         // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
        AccessToken string     `json:"access_token"`    // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
        Data        []DataItem `json:"data"`
}

type DataItem struct {
        Metrics   Metrics   `json:"metrics"`
        Target    string    `json:"target"`
        Dimension Dimension `json:"dimension"`
}

type Metrics struct {
        CpuLoad     float64 `json:"cpu_load"`
        MemoryUsage float64 `json:"memory_usage"`
}

type Dimension struct {
        Module string `json:"module"`
        Region string `json:"region"`
}

func init() {
        rand.Seed(time.Now().UnixNano())
}

func collectMetrics() (float64, float64, error) {
        cpuLoad := rand.Float64() * 100

        memoryUsage := rand.Float64() * 100

        return cpuLoad, memoryUsage, nil
}

func sendReport(apiURL, token string, dataID int, cpuLoad, memUsage float64) int {
        payload := Payload{
                DataID:      dataID,
                AccessToken: token,
                Data: []DataItem{
                        {
                                Metrics: Metrics{
                                        CpuLoad:     cpuLoad,
                                        MemoryUsage: memUsage,
                                },
                                Target: "127.0.0.1",
                                Dimension: Dimension{
                                        Module: "server",
                                        Region: "guangdong",
                                },
                        },
                },
        }

        jsonData, err := json.Marshal(payload)
        if err != nil {
                fmt.Printf("⚠️ JSON序列化异常: %v\n", err)
                return 500
        }

        req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
        if err != nil {
                fmt.Printf("⚠️ 创建请求异常: %v\n", err)
                return 500
        }
        req.Header.Set("Content-Type", "application/json")

        client := &http.Client{Timeout: 10 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
                fmt.Printf("⚠️ 请求异常: %v\n", err)
                return 500
        }
        defer resp.Body.Close()

        return resp.StatusCode
}

func main() {
        // ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 {{access_config.custom.http}} 」，其他环境、跨云场景请根据页面接入指引填写
        apiURL := getEnv("API_URL", "")
        token := getEnv("TOKEN", "")             // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
        dataIDStr := getEnv("DATA_ID", "")       // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
        intervalStr := getEnv("INTERVAL", "60")  // 上报间隔，默认为60秒

        dataID, err := strconv.Atoi(dataIDStr)
        if err != nil {
                fmt.Printf("❌ DATA_ID 必须为整数: %v\n", err)
                return
        }

        interval, err := strconv.Atoi(intervalStr)
        if err != nil {
                fmt.Printf("❌ INTERVAL 必须为整数: %v\n", err)
                return
        }

        if apiURL == "" || token == "" || dataIDStr == "" {
                fmt.Println("❌ 缺少必要环境变量: API_URL, TOKEN 或 DATA_ID")
                return
        }

        fmt.Printf("🚀 开始自定义指标上报，间隔: %d秒\n", interval)

        for {
                cpuLoad, memUsage, err := collectMetrics()
                if err != nil {
                        fmt.Printf("⚠ 生成自定义指标失败: %v\n", err)
                        time.Sleep(time.Duration(interval) * time.Second)
                        continue
                }

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

func getEnv(key, defaultValue string) string {
        value := os.Getenv(key)
        if value == "" {
                return defaultValue
        }
        return value
}
```

## 3. 了解更多

* 进行 <a href="{{docs.metrics.learn.Index_search}}" target="_blank">指标检索</a>。

* 了解 <a href="{{docs.metrics.learn.Use_indicators}}" target="_blank">怎么使用监控指标</a>。

* 了解如何 <a href="{{docs.metrics.learn.configure_dashboard}}" target="_blank">配置仪表盘</a>。

* 了解如何使用 <a href="{{docs.metrics.learn.alarms}}" target="_blank">监控告警</a>。
