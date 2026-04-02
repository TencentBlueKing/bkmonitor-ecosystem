// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

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
        // ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 {{access_config.custom.http}} 」，其他环境、跨云场景请根据页面接入指引填写
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
