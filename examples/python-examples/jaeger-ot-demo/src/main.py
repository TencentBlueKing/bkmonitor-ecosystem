# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import logging
import sys
import time
from typing import List

from config import config
from services.base import Service
from services.otlp import OtlpService
from services.querier import QuerierService
from services.server import HttpService

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s: %(message)s", datefmt="%Y-%m-%d %H:%M:%S")
logger = logging.getLogger(__name__)


def main():
    services: List[Service] = [
        OtlpService(config),
        HttpService(config),
        QuerierService(config),
    ]

    for service in services:
        try:
            service.start()
            logger.info("[%s] service started", service)
        except Exception as e:  # pylint: disable=broad-except
            logger.error("[%s] failed to start: %s", service, e)
            sys.exit(1)

    logger.info("[main] 🚀")
    logger.info("Press CTRL+C to quit")

    try:
        while True:
            time.sleep(0.1)
    except KeyboardInterrupt:
        pass

    for service in services:
        try:
            service.stop()
        except KeyboardInterrupt:
            pass
        except Exception as e:  # pylint: disable=broad-except
            logger.error("[%s] failed to stop: %s", service, e)
        else:
            logger.info("[%s] service stopped", service)

    logger.info("[main] 👋")


if __name__ == "__main__":
    main()
