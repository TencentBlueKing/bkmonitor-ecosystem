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

# 参数解析
while getopts "t:d:a:i:h" opt; do
    case $opt in
        t) TOKEN="$OPTARG" ;;
        d) DATA_ID="$OPTARG" ;;
        a) API_URL="$OPTARG" ;;
        i) INTERVAL="$OPTARG" ;;
        \?) echo "无效选项: -$OPTARG" >&2;;
        :) echo "选项 -$OPTARG 需要参数值" >&2;;
    esac
done

# 参数校验
if [ -z "$TOKEN" ] || [ -z "$DATA_ID" ] || [ -z "$API_URL" ]; then
    echo "错误：缺少必要参数！"
    exit 1
fi

generate_random_cpu() {
    echo "scale=2; $RANDOM / 327.67" | bc
}

generate_random_memory() {
    echo "scale=2; 70 + ($RANDOM % 2601) / 100" | bc
}

# 循环上报逻辑
while true; do
   cpu_load=$(generate_random_cpu)
   memory_usage=$(generate_random_memory)
    # 构建上报数据体（JSON结构）
    # ❗❗【非常重要】 data_id，标识上报的数据类型
    # ❗❗【非常重要】access_token:认证令牌，用于接口鉴定
    payload=$(cat <<EOF
{
    "data_id": $DATA_ID,
    "access_token": "$TOKEN",
    "data": [{
        "metrics": {
            "cpu_load": $cpu_load,
            "memory_usage": $memory_usage
        },
        "target": "127.0.0.1",
        "dimension": {
            "module": "db",
            "location": "guangdong",
            "language": "shell"
        }
    }]
}
EOF
)

    # 发送请求（curl 参数）
    response=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -d "$payload")

    # 结果处理
    if [ "$response" -eq 200 ]; then
        echo "$(date +"%F %T") 上报成功 | CPU负载: ${cpu_load}"  内存使用率:"${memory_usage}%"
    else
        echo "$(date +"%F %T") 上报失败 | 状态码: $response"
        exit 1
    fi

    # 间隔等待
    sleep $INTERVAL
done
