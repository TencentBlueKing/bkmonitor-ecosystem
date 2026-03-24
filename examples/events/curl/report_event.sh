#!/bin/bash
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


# ===== 默认配置 =====
API_URL="fixme"      # ❗❗【非常重要】数据上报接口地址（`Access URL`），请根据页面接入指引填写。
DATA_ID=0000000                # ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
TOKEN="xxxxxx"               # ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
TARGET_IP="127.0.0.1"
INTERVAL=60            # 默认上报间隔

# 使用getopts解析参数
while getopts ":a:d:t:i:T:h" opt; do
  case $opt in
    a) API_URL="$OPTARG" ;;
    d) DATA_ID="$OPTARG" ;;
    t) TOKEN="$OPTARG" ;;
    i) INTERVAL="$OPTARG" ;;
    T) TARGET_IP="$OPTARG" ;;
    h) usage ;;
    \?) echo "无效选项: -$OPTARG" >&2; usage ;;
    :) echo "选项 -$OPTARG 需要参数值" >&2; usage ;;
  esac
done

# 验证必填参数
if [[ -z "$API_URL" || -z "$DATA_ID" || -z "$TOKEN" ]]; then
  echo "错误：缺少必要参数！" >&2
  usage
fi

# ===== 上报函数 =====
report_event() {
  local timestamp=$(date +%s%3N)
  local cpu_value=$(( RANDOM % 20 + 80 ))

  curl -s -X POST $API_URL \
    -H "Content-Type: application/json" \
    -d "$(cat <<EOF
{
  "data_id": $DATA_ID,
  "access_token": "$TOKEN",
  "data": [{
    "event_name": "cpu_alert",
    "event": { "content": "CPU告警: $cpu_value%" },
    "target": "$TARGET_IP",
    "dimension": {
      "module": "db",
      "location": "guangdong",
      "language": "shell"
    },
    "timestamp": $timestamp
  }]
}
EOF
  )"
  echo "[$(date +'%F %T')] 上报完成 CPU=${cpu_value}%"
}

# ===== 执行流程 =====
echo "启动监控上报服务"
echo "├─ 代理地址: $API_URL"
echo "├─ 数据源ID: $DATA_ID"
echo "├─ 目标主机: $TARGET_IP"
echo "└─ 上报间隔: ${INTERVAL}秒"

# 首次立即上报
report_event

# 定时循环上报
while true; do
  sleep $INTERVAL
  report_event
done
