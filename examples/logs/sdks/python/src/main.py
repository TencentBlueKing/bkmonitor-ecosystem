# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import logging
import os
import random
import time

from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter
from opentelemetry.sdk._logs import LoggerProvider, LoggingHandler
from opentelemetry.sdk._logs.export import BatchLogRecordProcessor
from opentelemetry.sdk.resources import Resource

# ---------- 基础配置 ----------
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger("otel")

# ❗❗【非常重要】认证令牌，用于接口鉴权，请替换为页面提供的日志数据源 Token。
TOKEN = os.environ.get("TOKEN", "fixme")
# ❗❗【非常重要】上报地址，国内站点默认是「 http://127.0.0.1:4318/v1/logs 」，
# 其他环境、跨云场景请根据页面接入指引填写
API_URL = os.environ.get("API_URL", "http://127.0.0.1:4318/v1/logs")

# ---------- 初始化 OTel LoggerProvider ----------
log_exporter = OTLPLogExporter(
    endpoint=API_URL,
    headers={"x-bk-token": TOKEN},
)

logger_provider = LoggerProvider(
    resource=Resource.create(
        {
            "service.name": "custom-log-sdk-demo",
            "deployment.environment": "local",
        }
    )
)
logger_provider.add_log_record_processor(BatchLogRecordProcessor(log_exporter))

# 将 Python logging 桥接到 OpenTelemetry
handler = LoggingHandler(level=logging.NOTSET, logger_provider=logger_provider)
logger.addHandler(handler)

LOG_LEVELS = [
    (logging.DEBUG, "debug log from python sdk"),
    (logging.INFO, "info log from python sdk"),
    (logging.WARNING, "warn log from python sdk"),
    (logging.ERROR, "error log from python sdk"),
]


def main():
    logger.info("Starting log reporter (press Ctrl+C to stop)...")
    try:
        while True:
            level, message = random.choice(LOG_LEVELS)
            logger.log(level, message)
            time.sleep(0.1)  # 每 0.1 秒上报一条随机级别的日志
    except KeyboardInterrupt:
        logger.info("Received keyboard interrupt, exiting...")


if __name__ == "__main__":
    main()
