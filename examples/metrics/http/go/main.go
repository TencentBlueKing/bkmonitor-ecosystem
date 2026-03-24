// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 提供了自定义指标上报的示例实现
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
	DataID      int        `json:"data_id"`      // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
	AccessToken string     `json:"access_token"` // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
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
	token := getEnv("TOKEN", "")            // ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 TOKEN
	dataIDStr := getEnv("DATA_ID", "")      // ❗❗【非常重要】 data_id，标识上报的数据类型，配置为应用数据 ID
	intervalStr := getEnv("INTERVAL", "60") // 上报间隔，默认为60秒

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

// getEnv 获取环境变量值，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
