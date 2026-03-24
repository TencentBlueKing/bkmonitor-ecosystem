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
from dataclasses import dataclass

import requests
from config import Config
from opentelemetry import trace
from opentelemetry.instrumentation.requests import RequestsInstrumentor

logger = logging.getLogger(__name__)
otel_logger = logging.getLogger("otel")


INTERVAL = 3


@dataclass
class QuerierConfig:
    service_name: str
    scheme: str
    address: str
    port: int

    @property
    def url(self) -> str:
        return f"{self.scheme}://{self.address}:{self.port}/helloworld"


class QuerierService:
    def __init__(self, config: Config):
        self.config = QuerierConfig(
            service_name=config.service_name,
            scheme=config.http_scheme,
            address=config.http_address,
            port=config.http_port,
        )
        self.tracer = trace.get_tracer(self.config.service_name)
        self.stopped = threading.Event()

        RequestsInstrumentor().instrument()

    def start(self):
        self.stopped.clear()
        thread = threading.Thread(target=self._loop_query_hello_world)
        thread.start()

    def stop(self):
        self.stopped.set()

    def _loop_query_hello_world(self):
        logger.info("[%s] start loop_query_hello_world to periodically request %s", self, self.config.url)

        while not self.stopped.wait(INTERVAL):
            try:
                self._query_hello_world()
            except Exception as e:  # pylint: disable=broad-except
                otel_logger.error("[query_hello_world] got error -> %s", e)

        logger.info("[%s] loop_query_hello_world stopped", self)

    def _query_hello_world(self):
        with self.tracer.start_as_current_span("caller/query_hello_world"):
            otel_logger.info("[query_hello_world] send request")
            response = requests.get(self.config.url)
            otel_logger.info("[query_hello_world] received: %s", response.text)

    def __str__(self):
        return "querier"
