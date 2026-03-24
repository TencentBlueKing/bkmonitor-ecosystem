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
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// ===== 环境变量配置 =====
var (
	// ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
	apiURL = getEnv("API_URL", "")
	// ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
	dataID = getEnvInt("DATA_ID", 0)
	// ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
	token = getEnv("TOKEN", "")
	// 目标设备IP
	targetIP = getEnv("TARGET_IP", "127.0.0.1")
	// 上报间隔（秒）
	interval = getEnvInt("INTERVAL", 60)
	client   = &http.Client{Timeout: 5 * time.Second}
)

// Event 表示单个事件的数据结构
type Event struct {
	// 事件标识名
	EventName string `json:"event_name"`
	// 事件详细内容
	Event struct {
		Content string `json:"content"`
	} `json:"event"`
	// 事件上报目标
	Target string `json:"target"`
	// 事件维度，具体字段及内容可自定义填写
	Dimension map[string]string `json:"dimension"`
	// 事件发生时间戳
	Timestamp int64 `json:"timestamp"`
}

// Payload 发送到API的完整请求体结构
type Payload struct {
	// ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
	DataID int `json:"data_id"`
	// ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
	AccessToken string  `json:"access_token"`
	Data        []Event `json:"data"`
}

// ===== 日志功能 =====
func logf(format string, args ...interface{}) {
	fmt.Printf("\033[1m%s\033[0m | %s\n",
		time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
}

// sendEvents 发送事件并返回统一的 (status, message) 结构
func sendEvents() (string, string) {
	event := Event{
		EventName: "cpu_alert",
		Event: struct {
			Content string `json:"content"`
		}{Content: fmt.Sprintf("CPU告警: %d%%", rand.IntN(20)+80)},
		Target:    targetIP,
		Dimension: map[string]string{"module": "db", "location": "guangdong"},
		Timestamp: time.Now().UnixMilli(),
	}

	eventData := []Event{event}
	eventJSON, _ := json.MarshalIndent(eventData, "", "  ")
	logf("生成事件数据:\n%s", eventJSON)
	// ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
	// ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
	payload := Payload{DataID: dataID, AccessToken: token, Data: eventData}
	jsonData, _ := json.Marshal(payload)

	// ❗️❗️【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "error", fmt.Sprintf("上报失败: %v", err)
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == 200 {
		return "success", "上报成功"
	}
	return "error", fmt.Sprintf("HTTP %d", resp.StatusCode)
}

func main() {
	logf("事件上报服务启动 | 目标: %s | 间隔: %d秒", targetIP, interval)

	for {
		status, message := sendEvents()
		color := map[bool]string{true: "\033[32m", false: "\033[31m"}[status == "success"]
		logf("上报结果: %s%s %s\033[0m", color, status, message)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
