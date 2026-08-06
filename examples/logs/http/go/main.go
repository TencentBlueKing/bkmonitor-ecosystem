// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 提供了上报 OTLP 日志数据到蓝鲸监控平台的 HTTP 示例应用
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ==================== 环境变量辅助函数 ====================

// getEnv 从环境变量读取值，若为空则返回默认值
func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// getEnvInt 从环境变量读取整数值，若无效则返回默认值
func getEnvInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// ==================== 默认配置 ====================
var (
	// ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
	token = getEnv("TOKEN", "fixme")
	// ❗❗【非常重要】上报地址，国内站点默认是「 {{access_config.otlp.http_endpoint}}/v1/logs 」，
	// 其他环境、跨云场景请根据页面接入指引填写
	apiURL = getEnv("API_URL", "{{access_config.otlp.http_endpoint}}/v1/logs")
	// 上报间隔（秒）
	interval = getEnvInt("INTERVAL", 1)
	// HTTP 请求超时时间（秒）
	timeoutSec = getEnvInt("TIMEOUT_SEC", 10)
)

// ==================== 数据类型定义 ====================

// LogLevel 定义日志级别结构，包含 OTLP 协议中的严重程度编号和文本描述
type LogLevel struct {
	SeverityNumber int    `json:"severityNumber"`
	SeverityText   string `json:"severityText"`
	Message        string `json:"message"`
}

// logLevels 预定义的日志级别列表，覆盖 DEBUG/INFO/WARN/ERROR 四种级别
var logLevels = []LogLevel{
	{SeverityNumber: 5, SeverityText: "DEBUG", Message: "debug log from go http"},
	{SeverityNumber: 9, SeverityText: "INFO", Message: "info log from go http"},
	{SeverityNumber: 13, SeverityText: "WARN", Message: "warn log from go http"},
	{SeverityNumber: 17, SeverityText: "ERROR", Message: "error log from go http"},
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// ==================== 辅助函数 ====================

// getCurrentNanoTimestamp 返回当前 UTC 时间的纳秒级 Unix 时间戳字符串
func getCurrentNanoTimestamp() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getRandomLevel 随机返回一个日志级别
func getRandomLevel() LogLevel {
	return logLevels[rng.Intn(len(logLevels))]
}

// ==================== OTLP 日志上报逻辑 ====================

// buildPayload 构建 OTLP 格式的日志上报负载，包含随机选择的日志级别
func buildPayload() map[string]interface{} {
	currentNano := getCurrentNanoTimestamp()
	level := getRandomLevel()
	return map[string]interface{}{
		"resourceLogs": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"attributes": []interface{}{
						map[string]interface{}{
							"key": "service.name",
							"value": map[string]interface{}{
								"stringValue": "custom-log-demo",
							},
						},
						map[string]interface{}{
							"key": "deployment.environment.name",
							"value": map[string]interface{}{
								"stringValue": "local",
							},
						},
					},
				},
				"scopeLogs": []interface{}{
					map[string]interface{}{
						"scope": map[string]interface{}{
							"name": "java-http-demo",
						},
						"logRecords": []interface{}{
							map[string]interface{}{
								"timeUnixNano":         currentNano,
								"observedTimeUnixNano": currentNano,
								"severityNumber":       level.SeverityNumber,
								"severityText":         level.SeverityText,
								"body": map[string]interface{}{
									"stringValue": level.Message,
								},
								"attributes": []interface{}{
									map[string]interface{}{
										"key": "demo.source",
										"value": map[string]interface{}{
											"stringValue": "golang",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// doPost 发送 HTTP POST 请求上报日志
func doPost(client *http.Client, payload map[string]interface{}) {
	// 提取日志记录信息
	resourceLogs := payload["resourceLogs"].([]interface{})
	firstResourceLog := resourceLogs[0].(map[string]interface{})
	scopeLogs := firstResourceLog["scopeLogs"].([]interface{})
	firstScopeLog := scopeLogs[0].(map[string]interface{})
	logRecords := firstScopeLog["logRecords"].([]interface{})
	logRecord := logRecords[0].(map[string]interface{})

	log.Printf("Sending log level: %s (%v)",
		logRecord["severityText"], logRecord["severityNumber"])

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return
	}
	req, err := http.NewRequest("POST", apiURL,
		bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-bk-token", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to post request: %v", err)
		return
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	log.Printf("response.status_code=%d, body=%s",
		resp.StatusCode, buf.String())
}

// ==================== 主函数 ====================
func main() {
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	log.Println("Starting log reporter (press Ctrl+C to stop)...")

	// 捕获 SIGINT（Ctrl+C）和 SIGTERM 实现优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{}, 1)
	go func() {
		<-sigCh
		log.Println("Received signal, exiting...")
		close(done)
	}()

	for {
		select {
		case <-done:
			return
		default:
			payload := buildPayload()
			doPost(client, payload)
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
}
