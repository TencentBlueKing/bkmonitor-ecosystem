# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import sys
import time
from typing import List

from services.base import Service
from services.querier import querier_service
from services.server import http_service


def main():
    services: List[Service] = [
        http_service,
        querier_service,
    ]

    for service in services:
        try:
            service.start()
            print(f"[{service}] service started")
        except Exception as e:  # pylint: disable=broad-except
            print(f"[{service}] failed to start: {e}")
            sys.exit(1)

    print("[main] 🚀")
    print("Press CTRL+C to quit")

    try:
        while True:
            time.sleep(0.1)
    except KeyboardInterrupt:
        pass

    for service in reversed(services):
        try:
            service.stop()
        except KeyboardInterrupt:
            pass
        except Exception as e:  # pylint: disable=broad-except
            print(f"[{service}] failed to stop: {e}")
        else:
            print(f"[{service}] service stopped")

    print("[main] 👋")


if __name__ == "__main__":
    main()
