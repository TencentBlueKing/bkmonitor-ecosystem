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
        "time"
)

// ==================== 日志级别定义 ====================

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

// ==================== 辅助函数 ====================

// rng 全局随机数生成器，用于随机选择日志级别
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// getCurrentNanoTimestamp 返回当前 UTC 时间的纳秒级 Unix 时间戳字符串
func getCurrentNanoTimestamp() string {
        return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getRandomLevel 随机返回一个日志级别
func getRandomLevel() LogLevel {
        return logLevels[rng.Intn(len(logLevels))]
}

// ==================== OTLP LogRecord 数据结构定义 ====================

// Payload 定义完整的 OTLP LogRecord 请求结构，对应 OTLP 协议中的 ResourceLogs 集合
type Payload struct {
        ResourceLogs []ResourceLog `json:"resourceLogs"`
}

// ResourceLog 定义资源与日志的关联结构
type ResourceLog struct {
        Resource  Resource   `json:"resource"`
        ScopeLogs []ScopeLog `json:"scopeLogs"`
}

// Resource 定义资源属性，如服务名称、部署环境等
type Resource struct {
        Attributes []Attribute `json:"attributes"`
}

// ScopeLog 定义 instrumentation scope 级别的日志记录集合
type ScopeLog struct {
        Scope       Scope       `json:"scope"`
        LogRecords  []LogRecord `json:"logRecords"`
}

// Scope 定义 instrumentation scope 信息
type Scope struct {
        Name string `json:"name"`
}

// LogRecord 定义单条日志记录，包含时间戳、严重程度、日志体及属性
type LogRecord struct {
        TimeUnixNano         string      `json:"timeUnixNano"`
        ObservedTimeUnixNano string      `json:"observedTimeUnixNano"`
        SeverityNumber       int         `json:"severityNumber"`
        SeverityText         string      `json:"severityText"`
        Body                 Body        `json:"body"`
        Attributes           []Attribute `json:"attributes"`
}

// Attribute 定义键值对属性
type Attribute struct {
        Key   string      `json:"key"`
        Value ValueObject `json:"value"`
}

// ValueObject 定义属性的值对象（当前仅支持字符串类型）
type ValueObject struct {
        StringValue string `json:"stringValue,omitempty"`
}

// Body 定义日志体内容
type Body struct {
        StringValue string `json:"stringValue"`
}

// ==================== OTLP 日志上报逻辑 ====================

// buildPayload 构造 OTLP LogRecord 请求体
func buildPayload() Payload {
        currentNano := getCurrentNanoTimestamp()
        level := getRandomLevel()

        return Payload{
                ResourceLogs: []ResourceLog{
                        {
                                Resource: Resource{
                                        Attributes: []Attribute{
                                                {Key: "service.name",
                                                 Value: ValueObject{StringValue: "custom-log-demo"}},
                                                {Key: "deployment.environment.name",
                                                 Value: ValueObject{StringValue: "local"}},
                                        },
                                },
                                ScopeLogs: []ScopeLog{
                                        {
                                                Scope: Scope{Name: "custom-log-demo"},
                                                // instrumentation scope 名称，标识日志产生组件
                                                LogRecords: []LogRecord{
                                                        {
                                                                TimeUnixNano:         currentNano,
                                                                ObservedTimeUnixNano: currentNano,
                                                                SeverityNumber:       level.SeverityNumber,
                                                                SeverityText:         level.SeverityText,
                                                                Body:                 Body{StringValue: level.Message},
                                                                Attributes: []Attribute{
                                                                        {Key: "demo.source",
                                                                         Value: ValueObject{StringValue: "golang"}},
                                                                },
                                                        },
                                                },
                                        },
                                },
                        },
                },
        }
}

// doPost 发送 HTTP POST 请求
func doPost(payload Payload) {
        logRecord := payload.ResourceLogs[0].ScopeLogs[0].LogRecords[0]
        log.Printf("Sending log level: %s (%d)", logRecord.SeverityText, logRecord.SeverityNumber)

        // ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
        token := os.Getenv("TOKEN")
        if token == "" {
                token = "fixme"
        }

        // ❗❗【非常重要】上报地址，国内站点默认是「 {{access_config.otlp.http_endpoint}}/v1/logs 」，
        // 其他环境、跨云场景请根据页面接入指引填写
        apiURL := os.Getenv("API_URL")
        if apiURL == "" {
                apiURL = "xxxx"
        }

        // 序列化为 JSON
        jsonData, err := json.Marshal(payload)
        if err != nil {
                log.Printf("Failed to marshal payload: %v", err)
                return
        }

        req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
        if err != nil {
                log.Printf("Failed to create request: %v", err)
                return
        }

        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("x-bk-token", token) // ❗❗【非常重要】注入认证 TOKEN

        client := &http.Client{Timeout: 10 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
                log.Printf("Failed to post request: %v", err)
                return
        }
        defer resp.Body.Close()

        // 读取响应体（简单处理）
        buf := new(bytes.Buffer)
        _, _ = buf.ReadFrom(resp.Body)
        log.Printf("response.status_code=%d, body=%s", resp.StatusCode, buf.String())
}

// ==================== 主函数 ====================
func main() {
        log.Println("Starting log reporter (press Ctrl+C to stop)...")

        for {
                payload := buildPayload()
                doPost(payload)
                time.Sleep(100 * time.Millisecond) // 每 0.1 秒上报一条随机级别的日志
        }
}
