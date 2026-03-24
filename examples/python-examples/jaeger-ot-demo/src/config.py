# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import os
from enum import Enum


class ExporterType(Enum):
    GRPC = "grpc"
    HTTP = "http"


class Config:
    def __init__(self):
        self.debug = self._get_env_bool("DEBUG", False)
        self.token = os.getenv("TOKEN", "todo")
        self.service_name = os.getenv("SERVICE_NAME", "jaeger-client-demo")
        self.otlp_endpoint = os.getenv("OTLP_ENDPOINT", "http://localhost:4317")
        self.otlp_exporter_type = ExporterType(os.getenv("OTLP_EXPORTER_TYPE", "grpc").lower())
        self.enable_traces = self._get_env_bool("ENABLE_TRACES", self.debug)
        self.http_scheme = "http"
        self.http_address = "0.0.0.0"
        self.http_port = 8080

    @staticmethod
    def _get_env_bool(key: str, default: bool) -> bool:
        value = os.getenv(key)
        if value is None:
            return default
        return value.lower() in ("true", "1", "yes")


config = Config()
