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
from typing import Iterable, Optional, Tuple

from config import Config
from flask import Flask, Request, request
from opentelemetry import metrics, trace
from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.metrics import CallbackOptions, Observation
from opentelemetry.sdk.trace import Span, Status, StatusCode
from werkzeug.serving import BaseWSGIServer, make_server

logger = logging.getLogger(__name__)
otel_logger = logging.getLogger("otel")
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

    def __init__(self, service_name: str):
        self.tracer = trace.get_tracer(service_name)
        self.meter = metrics.get_meter(service_name)

        self.requests_total = self.meter.create_counter(
            "requests_total",
            description="Total number of HTTP requests",
        )
        self.task_execute_duration_seconds = self.meter.create_histogram(
            "task_execute_duration_seconds",
            unit="s",
            description="Task execute duration in seconds",
        )
        # Metrics（指标）- 使用 Gauge 类型指标
        self.meter.create_observable_gauge(
            "memory_usage",
            callbacks=[self.generate_random_usage],
            unit="%",
            description="Memory usage",
        )

    @staticmethod
    def generate_random_usage(options: CallbackOptions) -> Iterable[Observation]:
        usage = round(random.random(), 4)
        yield Observation(usage, {})

    def handle(self) -> str:
        # 不自动设置异常状态和记录异常，以展示手动设置方法 (traces_random_error_demo)
        with self.tracer.start_as_current_span(
            "handle/hello_world", record_exception=False, set_status_on_exception=False
        ):
            country = self.choice_country()
            otel_logger.info("get country -> %s", country)

            # Logs（日志）
            self.logs_demo(request)

            # Metrics（指标） - Counter 类型
            self.metrics_counter_demo(country)
            # Metrics（指标） - Histograms 类型
            self.metrics_histogram_demo()

            # Traces（调用链）- 自定义 Span
            self.traces_custom_span_demo()
            # Traces（调用链）- Span 事件
            self.traces_span_event_demo()
            # Traces（调用链）- 模拟错误
            self.traces_random_error_demo()

            return self.generate_greeting(country)

    # Logs（日志）打印日志
    @staticmethod
    def logs_demo(req: Request):
        otel_logger.info("received request: %s %s", req.method, req.path)

    def choice_country(self) -> str:
        return random.choice(self.COUNTRIES)

    # Metrics（指标）- 使用 Counter 类型指标
    # Refer: https://opentelemetry.io/docs/languages/python/instrumentation/#creating-and-using-synchronous-instruments
    def metrics_counter_demo(self, country: str):
        self.requests_total.add(1, {"country": country})

    # Metrics（指标）- 使用 Histogram 类型指标
    # Refer：https://opentelemetry.io/docs/specs/otel/metrics/api/#histogram-operations
    # Refer：https://opentelemetry.io/docs/specs/otel/metrics/data-model/#exemplars
    def metrics_histogram_demo(self):
        start_time = time.time()
        self.do_something(100)
        duration = time.time() - start_time
        self.task_execute_duration_seconds.record(duration)

    # Traces（调用链）- 增加自定义 Span
    # Refer: https://opentelemetry.io/docs/languages/python/instrumentation/#creating-spans
    def traces_custom_span_demo(self):
        with self.tracer.start_as_current_span("custom_span_demo/do_something") as span:
            # 添加 Span 自定义属性
            # Refer: https://opentelemetry.io/docs/languages/python/instrumentation/#add-attributes-to-a-span
            # 也可以使用 span.set_attributes() 批量设置
            span.set_attribute("helloworld.kind", 1)
            span.set_attribute("helloworld.step", "traces_custom_span_demo")

            self.do_something(50)

    # Traces（调用链）- Span 事件
    # Refer: https://opentelemetry.io/docs/languages/python/instrumentation/#adding-events
    def traces_span_event_demo(self):
        with self.tracer.start_as_current_span("span_event_demo/do_something") as span:
            attributes = {
                "helloworld.kind": 2,
                "helloworld.step": "traces_span_event_demo",
            }
            span.add_event("Before do_something", attributes)
            self.do_something(50)
            span.add_event("After do_something", attributes)

    # Traces（调用链）- 异常事件、状态
    # Refer: https://opentelemetry.io/docs/languages/python/instrumentation/#record-exceptions-in-spans
    def traces_random_error_demo(self):
        try:
            if random.random() < self.ERROR_RATE:
                error_message = random.choice(self.CUSTOM_ERROR_MESSAGES)
                raise APIException(error_message)
        except APIException as e:
            otel_logger.error("[traces_random_error_demo] got error -> %s", e)
            current_span: Span = trace.get_current_span()
            current_span.set_status(Status(StatusCode.ERROR, str(e)))
            current_span.record_exception(e)
            raise

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
        self.handler = HelloWorldHandler(service_name)
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
