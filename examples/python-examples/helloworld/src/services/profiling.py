# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


from dataclasses import dataclass

import pyroscope
from config import Config


@dataclass
class ProfilingConfig:
    enabled: bool
    token: str
    service_name: str
    endpoint: str


class BaseProfilingService:
    def __init__(self, config: Config):
        self.config = ProfilingConfig(
            # ❗❗【非常重要】请传入应用 Token
            token=config.token,
            # ❗❗【非常重要】应用服务唯一标识
            service_name=config.service_name,
            # ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
            endpoint=config.profiling_endpoint,
            enabled=config.enable_profiling,
        )

    def start(self):
        pass

    def stop(self):
        pass

    def __str__(self):
        return "profiling"


class PyroscopeProfilingService(BaseProfilingService):
    def start(self):
        if not self.config.enabled:
            return

        pyroscope.configure(
            # 服务名，一个应用可以有多个服务，通过该属性区分。
            application_name=self.config.service_name,
            # ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
            server_address=self.config.endpoint,
            tags={
                "service.name": self.config.service_name,
                "service.version": "0.1",
                "service.environment": "dev",
                "net.host.ip": "127.0.0.1",
                "net.host.name": "localhost",
            },
            http_headers={
                # ❗❗【非常重要】`X-BK-TOKEN` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，
                # 否则数据无法正常上报到 APM。
                "X-BK-TOKEN": self.config.token,
            },
        )

    def stop(self):
        if self.config.enabled:
            pyroscope.shutdown()
