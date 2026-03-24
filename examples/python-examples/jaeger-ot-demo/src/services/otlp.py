# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import platform
import socket
from dataclasses import dataclass
from typing import Optional, Type

from config import Config, ExporterType
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter as GRPCSpanExporter
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter as HTTPSpanExporter
from opentelemetry.sdk.resources import ProcessResourceDetector, Resource, ResourceDetector, get_aggregated_resources
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.semconv.resource import ResourceAttributes
from typing_extensions import assert_never

from services.base import Service

try:
    from opentelemetry.sdk.resources import OsResourceDetector
except ImportError:
    OsResourceDetector: Optional[Type[ResourceDetector]] = None


@dataclass
class OtlpConfig:
    token: str
    service_name: str
    endpoint: str
    exporter_type: ExporterType
    enable_traces: bool


class OtlpService(Service):
    def __init__(self, config: Config):
        self.config = OtlpConfig(
            token=config.token,
            service_name=config.service_name,
            endpoint=config.otlp_endpoint,
            exporter_type=config.otlp_exporter_type,
            enable_traces=config.enable_traces,
        )
        self.tracer_provider: Optional[TracerProvider] = None

    def start(self):
        resource = self._create_resource()

        if self.config.enable_traces:
            self._setup_traces(resource)

    def stop(self):
        if self.tracer_provider:
            self.tracer_provider.shutdown()

    def _create_resource(self) -> Resource:
        detectors = [ProcessResourceDetector()]
        if OsResourceDetector is not None:
            detectors.append(OsResourceDetector())

        # create 提供了部分 SDK 默认属性
        initial_resource = Resource.create(
            {
                # ❗❗【非常重要】应用服务唯一标识
                ResourceAttributes.SERVICE_NAME: self.config.service_name,
                ResourceAttributes.OS_TYPE: platform.system().lower(),
                ResourceAttributes.HOST_NAME: socket.gethostname(),
            }
        )

        return get_aggregated_resources(detectors, initial_resource)

    def _setup_traces(self, resource: Resource):
        otlp_exporter = self._setup_trace_exporter()
        span_processor = BatchSpanProcessor(otlp_exporter)
        self.tracer_provider = TracerProvider(resource=resource)
        self.tracer_provider.add_span_processor(span_processor)
        trace.set_tracer_provider(self.tracer_provider)

    def _setup_trace_exporter(self):
        if self.config.exporter_type == ExporterType.GRPC:
            return GRPCSpanExporter(
                endpoint=self.config.endpoint, insecure=True, headers={"x-bk-token": self.config.token}
            )
        elif self.config.exporter_type == ExporterType.HTTP:
            return HTTPSpanExporter(
                endpoint=f"{self.config.endpoint}/v1/traces", headers={"x-bk-token": self.config.token}
            )
        else:
            assert_never(self.config.exporter_type)

    def __str__(self):
        return "otlp"
