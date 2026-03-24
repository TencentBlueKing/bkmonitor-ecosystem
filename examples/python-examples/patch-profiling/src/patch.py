# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import binascii
import gzip
import json
import logging
import os
import typing
import urllib.parse

import six
import tenacity
from ddtrace.internal import agent
from ddtrace.profiling import exporter
from ddtrace.profiling.exporter.http import UploadFailed
from ddtrace.profiling.exporter.pprof import PprofExporter  # noqa
from six.moves import http_client

SAMPLE_TYPE_CONFIG = {
    "cpu-time": {
        "units": "samples",
        "aggregation": "sum",
        "display-name": "cpu_time",
        "sampled": True,
    },
    "wall-time": {
        "units": "samples",
        "aggregation": "sum",
        "display-name": "wall_time",
        "sampled": True,
    },
    # missing, leave it alone
    # "exception-samples": {
    #     "units": "samples",
    #     "aggregation": "sum",
    #     "display-name": "exception_samples",
    #     "sampled": True,
    # },
    # no proper units
    # "alloc-samples": {
    #     "units": "goroutines",
    #     "aggregation": "sum",
    #     "display-name": "alloc_samples",
    #     "sampled": True,
    # },
    "alloc-space": {
        "units": "bytes",
        "aggregation": "sum",
        "display-name": "alloc_space",
        "sampled": True,
    },
    "heap-space": {
        "units": "bytes",
        "aggregation": "average",
        "display-name": "heap_space",
        "sampled": False,
    },
}

SAMPLE_TYPE_CONFIG_JSON = json.dumps(SAMPLE_TYPE_CONFIG).encode()

logger = logging.getLogger(__name__)


class PyroscopePprofHTTPExporter(PprofExporter):
    """Send profiles via pprof format to pyroscope server"""

    def __init__(
        self,
        service_name: str,
        token: str,
        endpoint: str,
        max_retry_delay: int = 3,
        enable_code_provenance: bool = False,
    ):
        self.service_name = service_name
        self.token = token
        self.endpoint = endpoint
        self.max_retry_delay = max_retry_delay
        # useless in pyroscope now
        self.enable_code_provenance = enable_code_provenance

        self._retry_upload = tenacity.Retrying(
            # Retry after 1s, 2s, 4s, 8s with some randomness
            wait=tenacity.wait_random_exponential(multiplier=0.5),
            stop=tenacity.stop_after_delay(self.max_retry_delay),
            retry_error_cls=UploadFailed,
            retry=tenacity.retry_if_exception_type((http_client.HTTPException, OSError, IOError)),
        )

        if not self.token:
            logger.warning("[ProfileExporter] variable: TOKEN not set, profile will not be report")
        else:
            logger.info(
                "[ProfileExporter] start export profile to %s with service_name: %s", self.endpoint, self.service_name
            )

    def export(self, events, start_time_ns, end_time_ns):
        """Export events to an HTTP endpoint.

        @param events: The event dictionary from a `ddtrace.profiling.recorder.Recorder`.
        @param start_time_ns: The start time of recording.
        @param end_time_ns: The end time of recording.
        """
        profile, libs = super().export(events, start_time_ns, end_time_ns)
        pprof = six.BytesIO()
        with gzip.GzipFile(fileobj=pprof, mode="wb") as gz:
            gz.write(profile.SerializeToString())
        data = {
            b"profile": pprof.getvalue(),
            b"sample_type_config": SAMPLE_TYPE_CONFIG_JSON,
        }

        # pyroscope ignores content-type if format=pprof is provided
        # which will cause Parser.Parse() failed
        # https://github.com/pyroscope-io/pyroscope/blob/c068c7c0db1550b85031d7df0b56c84ce63036f6/pkg/server/ingest.go#L163
        params = {
            "name": self.service_name,
            "spyName": "ddtrace",
            "from": start_time_ns,
            "until": end_time_ns,
        }
        content_type, body = self._encode_multipart_formdata(data=data)

        headers = {
            "Content-Type": content_type,
            "Authorization": f"Bearer {self.token}",
        }

        self._retry_upload(self._upload_once, body, headers, params)

        return profile, libs

    @staticmethod
    def _encode_multipart_formdata(data) -> typing.Tuple[bytes, bytes]:
        boundary = binascii.hexlify(os.urandom(16))

        # The body that is generated is very sensitive and must perfectly match what the server expects.
        body = (
            b"".join(
                (
                    b'--%s\r\nContent-Disposition: form-data; name="%s"; filename="%s"\r\n'
                    % (boundary, field_name, field_name)
                )
                + b"Content-Type: application/octet-stream\r\n\r\n"
                + field_data
                + b"\r\n"
                for field_name, field_data in data.items()
            )
            + b"--%s--" % boundary
        )

        content_type = b"multipart/form-data; boundary=%s" % boundary

        return content_type, body

    def _assemble_url(self, params: dict) -> str:
        """Assemble url with path and params"""
        return f"{self.endpoint}?{urllib.parse.urlencode(params)}"

    def _upload_once(self, body, headers: dict, params: dict):
        """Upload profile to target"""
        url = self._assemble_url(params)
        client = agent.get_connection(self.endpoint)

        client.request("POST", url, body=body, headers=headers)
        resp = client.getresponse()

        if 200 <= resp.status < 300:
            return

        if 500 <= resp.status:
            raise tenacity.TryAgain

        if resp.status == 400:
            raise exporter.ExportError("Server returned 400, check your API key.")
        elif resp.status == 404:
            raise exporter.ExportError("Server returned 404, check your endpoint path.")

        raise exporter.ExportError(f"POST to {url}, but got {resp.status}")


def patch_ddtrace_to_pyroscope(service_name: str, token: str, endpoint: str, enable_memory_profiling: bool = True):
    """Patching entrance"""
    from ddtrace.profiling.profiler import _ProfilerInstance  # noqa

    # 'cpu' always open，but 'mem' is on demand
    if not enable_memory_profiling:
        os.environ["DD_PROFILING_MEMORY_ENABLED"] = "False"

    def _build_default_exporters(self):  # noqa
        """Patch. Only return custom httpExporter"""
        return [PyroscopePprofHTTPExporter(service_name=service_name, token=token, endpoint=endpoint)]

    _ProfilerInstance._build_default_exporters = _build_default_exporters
