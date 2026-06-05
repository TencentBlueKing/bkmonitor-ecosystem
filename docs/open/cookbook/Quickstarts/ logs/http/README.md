# 自定义上报

## 1. 概述

自定义数据接入日志平台，支持采集器或OTLP等协议，按要求接入即可。

## 2. 准备开始

### 2.1 接入流程

自定义上报支持二种数据类型：容器日志上报、otlp日志上报。上报要求可查看页面右侧帮助文档。

* 基础信息：采集名、数据分类、数据名、所属索引集、备注说明。

![alt text](./images/image.png)

存储设置：

* 存储集群：支持选择共享集群；如有业务独享集群，可选择独享集群进行设置；

* 输入日志过期时间、副本数、分片数后完成配置。

![alt text](./images/image-1.png)

### 2.2 上报速率限制

otlp 日志上报注意 API 频率限制 5w/s。

如超过频率限制，请联系`蓝鲸助手`调整。

## 3. 快速接入

### 3.1 数据上报示例

* 了解 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/logs/http/python.md" target="_blank">Python - otlp（HTTP）上报</a>。

* 了解 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/logs/http/cpp.md" target="_blank">C++ - otlp（HTTP）上报</a>。

* 了解 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/logs/http/java.md" target="_blank">Java - otlp（HTTP）上报</a>。

* 了解 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/logs/http/go.md" target="_blank">Go - otlp（HTTP）上报</a>。


## 4. 了解更多

进一步了解以下内容：

* 进行 <a href="#" target="_blank">容器日志-自定义上报使用文档</a>。

* 进行 <a href="#" target="_blank">容器日志采集器安装</a>。

另一种方式是通过 SDK 上报自定义：

* 了解 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/logs/sdks/python.md" target="_blank">Python - otlp（SDK）上报</a>。

* 了解 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/logs/sdks/cpp.md" target="_blank">C++ - otlp（SDK）上报</a>。

* 了解 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/logs/sdks/java.md" target="_blank">Java - otlp（SDK）上报</a>。

* 了解 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/logs/sdks/go.md" target="_blank">Go - otlp（SDK）上报</a>。