# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import atexit
import logging
import random
import threading
import time

import pyroscope
import requests
from config import config
from flask import Flask

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s: %(message)s", datefmt="%Y-%m-%d %H:%M:%S")
logger = logging.getLogger(__name__)
app = Flask(__name__)


@app.route("/tasks")
def tasks():
    result = intensive_task()
    return {"result": result}


def intensive_task(duration: float = 0.5) -> float:
    start = time.time()
    total = 0
    while time.time() - start < duration:
        numbers = [random.random() for _ in range(10000)]
        total += sum(numbers)
    return total


def main():
    start_profiler()
    start_background_task_querier()
    start_flask_server()


def start_profiler():
    if not config.enable_profiling:
        return

    try:
        pyroscope.configure(
            # 服务名，一个应用可以有多个服务，通过该属性区分。
            application_name=config.service_name,
            # ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
            server_address=config.profiling_endpoint,
            tags={
                "service.name": config.service_name,
                "service.version": "0.1",
                "service.environment": "dev",
                "net.host.ip": "127.0.0.1",
                "net.host.name": "localhost",
            },
            http_headers={
                # ❗❗【非常重要】`X-BK-TOKEN` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，
                # 否则数据无法正常上报到 APM。
                "X-BK-TOKEN": config.token,
            },
        )
    except Exception as err:  # pylint: disable=broad-except
        print("start continues profiling failed: %s" % err)


def start_background_task_querier():
    stopped = threading.Event()
    atexit.register(stopped.set)  # 在 flask server 关闭后及时停止不必要的请求
    thread = threading.Thread(target=loop_query_tasks, args=(stopped,), daemon=True)
    thread.start()


def loop_query_tasks(stop_event: threading.Event):
    url = "http://localhost:8080/tasks"
    while not stop_event.wait(3):
        try:
            requests.get(url)
        except Exception as e:  # pylint: disable=broad-except
            logging.error("[querier] got error -> %s", e)


def start_flask_server():
    app.run(host="0.0.0.0", port=8080, debug=False)


if __name__ == "__main__":
    main()
