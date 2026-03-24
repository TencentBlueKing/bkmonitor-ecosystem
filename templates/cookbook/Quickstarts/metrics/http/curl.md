# 命令行-指标（HTTP）上报

## 1. 前置准备

### 1.1 术语介绍

* <a href="{{docs.metrics.What}}" target="_blank">什么是指标</a>

* <a href="{{docs.metrics.Types}}" target="_blank">指标类型</a>

### 1.2 开发环境要求

在开始之前，请确保您已经安装了以下软件：

* Git

```shell
＃ 可输入命令查看 git 是否安装
git --version
```

### 1.3 初始化 demo

```shell
git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/metrics/http/curl
```

## 2. 快速接入

### 2.1 前置准备

参考 <a href="{{docs.metrics.http.readme.metrics_http_readme}}" target="_blank">自定义指标 HTTP 上报</a> 创建一个上报协议为 `JSON` 的自定义指标，关注创建后提供的两个配置项：

* `TOKEN`：自定义指标数据源 Token，上报数据时使用。

* `数据 ID`: 数据 ID（Data ID），自定义指标数据源唯一标识，上报数据时使用。

同时，阅读上述文档「上报数据协议」章节。

![alt text](./images/image.png)

**有任何问题可企微联系`蓝鲸助手`协助处理**。

### 2.2 样例运行参数

运行参数说明：

| 参数名称     | 类型                | 描述                         |
| ------------ | ------------------- | ---------------------------- |
|`-t`       |String     |❗❗【非常重要】 自定义指标数据源 `Token`。   |
|`-d`       |Integer    |❗❗【非常重要】 数据 ID（`Data ID`），自定义指标数据源唯一标识。|
|`-a`       |String     |❗❗【非常重要】 数据上报接口地址（`Access URL`），国内站点请填写「 {{access_config.custom.http}} 」，其他环境、跨云场景请根据页面接入指引填写。 |
|`-i`       |Integer    |上报间隔时间（秒）。|

### 2.3 运行样例

示例代码也可以在样例仓库 <a href="{{ECOSYSTEM_CODE_ROOT_URL}}/examples/metrics/http/curl/report_metrics.sh" target="_blank">bkmonitor-ecosystem/examples/metrics/http/curl</a> 中找到。

该样例通过命令行实现周期上报指标：

```shell
export TOKEN="fixme:替换为申请到的 Token" # ❗❗【非常重要】 自定义指标数据源 Token。
export DATA_ID="fixme:替换为申请到的 DataID"  # ❗❗【非常重要】 数据 ID（`Data ID`），自定义指标数据源唯一标识。
#❗❗【非常重要】指标上报地址，国内站点请填写「 {{access_config.custom.http}} 」，其他环境、跨云场景请根据页面接入指引填写。
export ACCESS_URL="fixme"
./report_metrics.sh -t $TOKEN -d $DATA_ID -a $ACCESS_URL -i $INTERVAL
```

运行输出：

```shell
2025-06-06 15:36:26 上报成功 | CPU负载: 14.89 内存使用率:70.21%
2025-06-06 15:36:36 上报成功 | CPU负载: 14.48 内存使用率:70.57%
2025-06-06 15:36:46 上报成功 | CPU负载: 14.06 内存使用率:70.25%
2025-06-06 15:36:56 上报成功 | CPU负载: 13.30 内存使用率:70.02%
...
```

### 2.4 样例代码

```shell
#!/bin/bash

# 参数解析
while getopts "t:d:a:i:" opt; do
    case $opt in
        t) TOKEN="$OPTARG" ;;
        d) DATA_ID="$OPTARG" ;;
        a) API_URL="$OPTARG" ;;
        i) INTERVAL="$OPTARG" ;;
        *) echo "用法: $0 -t TOKEN -d DATA_ID -a API_URL -i INTERVAL"; exit 1 ;;
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
            "location": "guangdong"
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
    sleep "$INTERVAL"
done
```

## 4. 了解更多

* 进行 <a href="{{docs.metrics.learn.Index_search}}" target="_blank">指标检索</a>。

* 了解 <a href="{{docs.metrics.learn.Use_indicators}}" target="_blank">怎么使用监控指标</a>。

* 了解如何 <a href="{{docs.metrics.learn.configure_dashboard}}" target="_blank">配置仪表盘</a>。

* 了解如何使用 <a href="{{docs.metrics.learn.alarms}}" target="_blank">监控告警</a>。
