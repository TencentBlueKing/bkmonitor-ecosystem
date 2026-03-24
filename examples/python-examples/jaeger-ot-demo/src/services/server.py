# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import logging
import random
import threading
import time
from dataclasses import dataclass
from typing import Optional, Tuple

from config import Config
from flask import Flask
from jaeger_tracer import init_tracer
from opentelemetry.instrumentation.flask import FlaskInstrumentor
from werkzeug.serving import BaseWSGIServer, make_server

logger = logging.getLogger(__name__)
logging.getLogger("werkzeug").disabled = True  # 关闭 werkzeug 日志输出


@dataclass
class ServerConfig:
    service_name: str
    scheme: str
    address: str
    port: int


class APIException(Exception):
    status_code = 500


class HelloWorldHandler:
    COUNTRIES = [
        "United States",
        "Canada",
        "United Kingdom",
        "Germany",
        "France",
        "Japan",
        "Australia",
        "China",
        "India",
        "Brazil",
    ]
    ERROR_RATE = 0.1
    CUSTOM_ERROR_MESSAGES = [
        "mysql connect timeout",
        "user not found",
        "network unreachable",
        "file not found",
    ]

    def __init__(self):
        self.tracer = init_tracer()

    def handle(self) -> str:
        # 不自动设置异常状态和记录异常，以展示手动设置方法 (traces_random_error_demo)
        with self.tracer.start_active_span("handle/hello_world"):
            country = self.choice_country()
            logger.info("get country -> %s", country)

            # Traces（调用链）- 自定义 Span
            self.traces_custom_span_demo()
            # Traces（调用链）- Span 事件
            self.traces_span_event_demo()

            return self.generate_greeting(country)

    def choice_country(self) -> str:
        return random.choice(self.COUNTRIES)

    def traces_custom_span_demo(self):
        with self.tracer.start_active_span("custom_span_demo/do_something") as scope:
            span = scope.span
            # 添加 Span 自定义 attributes
            span.set_tag("helloworld.kind", 1)
            span.set_tag("helloworld.step", "traces_custom_span_demo")
            self.do_something(50)

    def traces_span_event_demo(self):
        with self.tracer.start_active_span("span_event_demo/do_something") as scope:
            span = scope.span
            attributes = {
                "helloworld.kind": 2,
                "helloworld.step": "traces_span_event_demo",
            }
            # 添加 Span 自定义 events 信息
            span.log_kv(attributes)
            self.do_something(50)

    @staticmethod
    def generate_greeting(country: str) -> str:
        return f"Hello World, {country}!"

    @staticmethod
    def do_something(max_ms: int):
        duration = max(10, random.randint(0, max_ms)) / 1000
        i = 0
        start = time.time()
        while time.time() - start < duration:
            i += 1


class HttpService:
    def __init__(self, config: Config):
        service_name = config.service_name
        self.config = ServerConfig(
            service_name=service_name,
            scheme=config.http_scheme,
            address=config.http_address,
            port=config.http_port,
        )
        self.server: Optional[BaseWSGIServer] = None
        self.server_thread: Optional[threading.Thread] = None

        self.app = Flask(service_name)
        self.handler = HelloWorldHandler()
        self.app.add_url_rule("/helloworld", view_func=self.handler.handle)
        self.app.register_error_handler(APIException, self._error_handler)
        FlaskInstrumentor().instrument_app(self.app)

    def start(self):
        # 优雅退出，refer：https://stackoverflow.com/questions/15562446
        # threaded 设置为 True 的作用：
        # 1. 处理请求时，新建线程，提高并发能力
        # 2. make_server 函数会返回 ThreadedWSGIServer 对象
        self.server = make_server(self.config.address, port=self.config.port, app=self.app, threaded=True)

        # 这里不使用任何 reloader 包装，因为 reloader 会使检测失效
        # https://opentelemetry.io/docs/zero-code/python/example/#instrumentation-while-debugging
        self.server_thread = threading.Thread(target=self.server.serve_forever)
        self.server_thread.start()
        logger.info("[%s] start to listen http server at %s:%s", self, self.config.address, self.config.port)

    @staticmethod
    def _error_handler(e: APIException) -> Tuple[str, int]:
        return str(e), e.status_code

    def stop(self):
        # 跳出 server_forever loop
        # 会自动清理资源，比如等待未响应请求的结束
        if self.server is not None:
            self.server.shutdown()

        if self.server_thread is not None:
            self.server_thread.join()

    def __str__(self):
        return "http"
