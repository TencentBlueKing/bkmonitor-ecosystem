# 命令行-事件（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Quickstarts/events/http/README.md" target="_blank">自定义事件 HTTP 上报</a>

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

```shell
＃ 可输入命令查看 git 是否安装
git --version
```

### 1.3 初始化 demo

```shell
git clone https://github.com/TencentBlueKing/bkmonitor-ecosystem
cd bkmonitor-ecosystem/examples/events/curl
```

## 2. 快速接入

### 2.1 创建应用

参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/main/docs/cookbook/Quickstarts/events/http/README.md" target="_blank">自定义事件 HTTP 上报</a> 创建自定义事件后需关注提供的两个配置项：

* `TOKEN`：自定义事件数据源 Token，上报数据时使用。

* `数据 ID`: 数据 ID（Data ID），自定义事件数据源唯一标识，上报数据时使用。

同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image-1.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数     | 类型                | 描述                         |
| ------------ | ------------------- | ---------------------------- |
|`TOKEN`       |String      |❗❗【非常重要】自定义指标数据源 `Token`。  |
|`DATA_ID`       |Integer     |❗❗【非常重要】数据 ID（`Data ID`），自定义指标数据源唯一标识。|
|`API_URL`       |String         |❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写。|
|`INTERVAL` |Integer     |上报间隔（单位为秒），默认 60 秒上报一次。​ |

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/tree/main/examples/events/curl/report_event.sh" target="_blank">bkmonitor-ecosystem/examples/events/curl</a> 中找到。

该样例通过命令行实现周期上报事件：

```shell
# ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
export TOKEN="fixme:替换为申请到的 Token"
# ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
export DATA_ID="fixme:替换为申请到的 DataID"
# ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写。
export API_URL="fixme"
./report_event.sh -t $TOKEN -d $DATA_ID -a $API_URL -i $INTERVAL
```

运行输出：

```shell
启动监控上报服务
├─ 代理地址: fixme
├─ 数据源ID: 0000000
├─ 目标主机: 127.0.0.1
└─ 上报间隔: 60秒
{"code":"200","result":"true","message":""}[2025-08-07 19:59:03] 上报完成 CPU=93%
{"code":"200","result":"true","message":""}[2025-08-07 20:00:03] 上报完成 CPU=83%
...
```

### 2.4 样例代码

```shell
#!/bin/bash

# ===== 默认配置 =====
# ❗❗【非常重要】数据上报接口地址（`Access URL`），国内站点请填写「 http://127.0.0.1:10205/v2/push/ 」，其他环境、跨云场景请根据页面接入指引填写。
API_URL="fixme"
# ❗❗【非常重要】标识上报的数据类型，配置为应用数据 `ID`。
DATA_ID=0000000
# ❗❗【非常重要】认证令牌，用于接口鉴定，配置为应用 `TOKEN`。
TOKEN="xxxxxx"
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
      "location": "guangdong"
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
```

## 3. 了解更多

* <a href="#" target="_blank">事件数据接入</a>。

* <a href="#" target="_blank">主机事件</a>。

* <a href="#" target="_blank">容器事件</a>。