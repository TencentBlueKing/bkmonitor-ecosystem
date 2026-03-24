# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import logging
import threading
from urllib.parse import urljoin

import requests
from config import config
from opentelemetry import trace

from services.server import http_service

logger = logging.getLogger(__name__)

tracer = trace.get_tracer(__name__)

INTERVAL = 10


class QuerierService:
    def __init__(self):
        self.base_url = f"{config.http_scheme}://{config.http_address}:{config.http_port}"
        self.stopped = threading.Event()

    def start(self):
        self.stopped.clear()
        thread = threading.Thread(target=self._loop_query)
        thread.start()

    def stop(self):
        self.stopped.set()

    def _loop_query(self):
        logger.info("[%s] start loop_query to periodically request %s", self, self.base_url)

        while not self.stopped.wait(INTERVAL):
            try:
                for rule_obj in http_service.app.url_map.iter_rules():
                    rule_str = rule_obj.rule
                    if not rule_str.startswith("/static"):
                        self._query(rule_str)
            except Exception as e:  # pylint: disable=broad-except
                logger.error("[loop_query] got error -> %s", e)

        logger.info("[%s] loop_query stopped", self)

    def _query(self, rule_str: str):
        query_url = urljoin(self.base_url, rule_str)
        with tracer.start_as_current_span(f"caller{rule_str}"):
            logger.info("[query %s] send request", rule_str)
            response = requests.get(query_url)
            logger.info("[query %s] received: %s", rule_str, response.text)

    def __str__(self):
        return "querier"


querier_service = QuerierService()
