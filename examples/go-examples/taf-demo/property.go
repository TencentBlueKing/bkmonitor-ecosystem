// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

// Package main 提供了 TAF 框架示例应用的属性管理实现
package main

import "github.com/TarsCloud/TarsGo/tars"

// Properties 管理服务的性能指标属性
type Properties struct {
	add         *tars.PropertyReport
	addRequests *tars.PropertyReport
	sub         *tars.PropertyReport
	subRequests *tars.PropertyReport
}

func newAddP() *tars.PropertyReport {
	sum := tars.NewSum()
	max := tars.NewMax()
	min := tars.NewMin()
	avg := tars.NewAvg()
	disr := tars.NewDistr([]int{0, 50, 100, 200})
	p := tars.CreatePropertyReport("Add", sum, max, min, avg, disr)
	return p
}

func newAddRequestsP() *tars.PropertyReport {
	count := tars.NewCount()
	p := tars.CreatePropertyReport("AddRequests", count)
	return p
}

func newSubP() *tars.PropertyReport {
	sum := tars.NewSum()
	max := tars.NewMax()
	min := tars.NewMin()
	avg := tars.NewAvg()
	disr := tars.NewDistr([]int{-100, -50, 0, 50, 100})
	p := tars.CreatePropertyReport("Sub", sum, max, min, avg, disr)
	return p
}

func newSubRequestsP() *tars.PropertyReport {
	count := tars.NewCount()
	p := tars.CreatePropertyReport("SubRequests", count)
	return p
}

// NewProperties 创建并返回一个新的 Properties 实例
func NewProperties() *Properties {
	return &Properties{
		add:         newAddP(),
		addRequests: newAddRequestsP(),
		sub:         newSubP(),
		subRequests: newSubRequestsP(),
	}
}
