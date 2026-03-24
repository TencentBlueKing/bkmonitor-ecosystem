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
from typing import List, Optional, Tuple

from config import config
from flask import Flask
from opentelemetry import metrics, trace
from opentelemetry.context import get_current
from opentelemetry.propagate import extract, inject
from opentelemetry.sdk.trace import Span, Status, StatusCode
from werkzeug.serving import BaseWSGIServer, make_server

logger = logging.getLogger(__name__)
logging.getLogger("werkzeug").disabled = True  # 关闭 werkzeug 日志输出

tracer = trace.get_tracer(__name__)
meter = metrics.get_meter(__name__)


class APIException(Exception):
    status_code = 500


class TravelHandler:
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
        # Counter 类型
        # 计算 visit 函数调用次数
        # Refer：https://opentelemetry.io/docs/specs/otel/metrics/api/#counter-creation
        self.visit_requests_total = meter.create_counter(
            "visit_requests_total",
            description="Total number of times the visit function is called",
        )

        # Histograms 类型
        # 计算 visit 函数耗时
        # Refer：https://opentelemetry.io/docs/specs/otel/metrics/api/#histogram-creation
        self.visit_execute_duration_seconds = meter.create_histogram(
            "visit_execute_duration_seconds",
            unit="s",
            description="visit function execute duration in seconds",
        )

    def visit_handle(self) -> str:
        # 不自动设置异常状态和记录异常，以展示手动设置方法 (traces_random_error_demo)
        with tracer.start_as_current_span("travel/visit_handle", record_exception=False, set_status_on_exception=False):
            countries = self.choice_countries()
            logger.info("get countries -> %s", countries)

            self.parallel_visit(countries)

            self.serial_visit(countries)

            return "Travel Success"

    def choice_countries(self) -> List[str]:
        return random.sample(self.COUNTRIES, 3)

    def visit(self, country: str):
        start_time = time.time()
        with tracer.start_as_current_span("travel/visit"):
            # Metrics（指标） - Counter 类型
            # 记录 visit 函数的调用次数，并按国家进行分类
            self.metrics_counter_demo(country)

            # Traces（调用链）- 模拟任务耗时
            # 随机休眠 200 ms~500 ms
            self.do_something()

            # Traces（调用链）- 模拟错误
            # 10% 的概率抛出一个随机的错误
            self.traces_random_error_demo()

            # Metrics（指标） - Histograms 类型
            # 记录 visit 函数的耗时
            self.metrics_histogram_demo(start_time)

    def parallel_visit(self, countries):
        with tracer.start_as_current_span("travel/parallel_visit"):
            trace_context = {}
            inject(trace_context, get_current())

            logger.info("parallel_visit start")

            threads = []
            for country in countries:
                thread = threading.Thread(target=self.parallel_visit_task, args=(country, trace_context))
                threads.append(thread)
                thread.start()
            for thread in threads:
                thread.join()

            logger.info("parallel_visit end")

    def serial_visit(self, countries):
        with tracer.start_as_current_span("travel/serial_visit"):
            logger.info("serial_visit start")

            for country in countries:
                self.visit(country)

            logger.info("serial_visit end")

    def parallel_visit_task(self, country, trace_context):
        context_content = extract(trace_context)
        with tracer.start_as_current_span("travel/parallel_visit_task", context=context_content):
            self.visit(country)

    # Traces（调用链）- 异常事件、状态
    # Refer: https://opentelemetry.io/docs/languages/python/instrumentation/#record-exceptions-in-spans
    def traces_random_error_demo(self):
        try:
            if random.random() < self.ERROR_RATE:
                error_message = random.choice(self.CUSTOM_ERROR_MESSAGES)
                raise APIException(error_message)
        except APIException as e:
            logger.error("[traces_random_error_demo] got error -> %s", e)
            current_span: Span = trace.get_current_span()
            current_span.set_status(Status(StatusCode.ERROR, str(e)))
            current_span.record_exception(e)
            raise

    # Metrics（指标）- 使用 Counter 类型指标
    # Refer: https://opentelemetry.io/docs/languages/python/instrumentation/#creating-and-using-synchronous-instruments
    # Refer：https://opentelemetry.io/docs/specs/otel/metrics/api/#counter-operations
    def metrics_counter_demo(self, country: str):
        self.visit_requests_total.add(1, {"country": country})

    # Metrics（指标）- 使用 Histogram 类型指标
    # Refer：https://opentelemetry.io/docs/specs/otel/metrics/data-model/#exemplars
    # Refer：https://opentelemetry.io/docs/specs/otel/metrics/api/#histogram-operations
    def metrics_histogram_demo(self, start_time):
        duration = time.time() - start_time
        self.visit_execute_duration_seconds.record(duration)

    # Traces（调用链）- 模拟任务耗时
    # 随机休眠 200 ms~500 ms
    @staticmethod
    def do_something():
        time.sleep(random.randint(200, 500) / 1000)


class HttpService:
    def __init__(self):
        self.server: Optional[BaseWSGIServer] = None
        self.server_thread: Optional[threading.Thread] = None

        self.app = Flask(__name__)
        self.travel_handler = TravelHandler()
        self.app.add_url_rule("/travel", view_func=self.travel_handler.visit_handle)
        self.app.register_error_handler(APIException, self._error_handler)

    def start(self):
        # 优雅退出，refer：https://stackoverflow.com/questions/15562446
        # threaded 设置为 True 的作用：
        # 1. 处理请求时，新建线程，提高并发能力
        # 2. make_server 函数会返回 ThreadedWSGIServer 对象
        self.server = make_server(config.http_address, port=config.http_port, app=self.app, threaded=True)

        # 这里不使用任何 reloader 包装，因为 reloader 会使检测失效
        # https://opentelemetry.io/docs/zero-code/python/example/#instrumentation-while-debugging
        self.server_thread = threading.Thread(target=self.server.serve_forever)
        self.server_thread.start()

        logger.info("[%s] start to listen http server at %s:%s", self, config.http_address, config.http_port)

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


http_service = HttpService()
