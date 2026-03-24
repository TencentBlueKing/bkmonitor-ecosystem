# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

from opentelemetry import trace
from opentelemetry.shim.opentracing_shim import TracerShim, create_tracer
from opentracing import set_global_tracer


def init_tracer() -> TracerShim:
    # 注释代码是使用 jaeger_client 上报的样例，用来同迁移到 OTel SDK 的代码做对比
    # from jaeger_client import Config
    # from config import config as custom_config
    # config = Config(
    #     config={
    #         'sampler': {
    #             'type': 'const',
    #             'param': 1,
    #         },
    #         'logging': True,
    #         "reporter_queue_size":10,
    #     },
    #     service_name=custom_config.service_name,
    #     validate=True,
    # )
    # return config.initialize_tracer()
    # 获取全局 Tracer Provider
    global_tracer_provider = trace.get_tracer_provider()

    # Create an OpenTracing shim.
    shim_tracer = create_tracer(global_tracer_provider)
    set_global_tracer(shim_tracer)
    return shim_tracer
