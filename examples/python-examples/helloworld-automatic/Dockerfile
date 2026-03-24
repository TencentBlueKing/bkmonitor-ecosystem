ARG PYTHON_VERSION=3.10

FROM python:${PYTHON_VERSION}-slim AS base

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    PATH="/.venv/bin:$PATH" \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

RUN set -ex && \
    sed -i 's/deb.debian.org/mirrors.cloud.tencent.com/g' /etc/apt/sources.list.d/debian.sources && \
    apt update

# ------------------- Stage 1: Build Stage ------------------------------
FROM base AS builder

WORKDIR /

COPY requirements.txt .

RUN set -ex && \
    mkdir ~/.pip && \
    echo "[global]" >> ~/.pip/pip.conf && \
    echo "index-url=https://mirrors.tencent.com/pypi/simple/" >> ~/.pip/pip.conf

RUN set -ex && \
    python -m venv .venv && \
    pip install -r ./requirements.txt

# ❗❗【非常重要】自动安装项目所需 OpenTelemetry instrument 库
RUN pip install opentelemetry-distro opentelemetry-exporter-otlp
RUN opentelemetry-bootstrap -a install

# ------------------- Stage 2: Final Stage ------------------------------
FROM base

WORKDIR /app

COPY --from=builder /.venv /.venv
COPY src .

# 服务名，一个应用可以有多个服务，通过该属性区分。有关服务名配置的更多信息，请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name
ENV OTEL_SERVICE_NAME="helloworld-automatic"

# Exporter 导出数据时附加额外的 Headers，多个 key-value 以逗号分隔，更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_headers
# 【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。
ENV OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo"

# Resource 代表观测数据所属的资源实体，并通过资源属性进行描述。更多信息请参考：https://opentelemetry.io/docs/concepts/resources/
# Resource Attributes 设置，多个 key-value 以逗号分隔，例如：`key1=value1,key2=value2`。更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_resource_attributes
ENV OTEL_RESOURCE_ATTRIBUTES=""

# Exporter 提供向消费者发送观测数据的功能。更多信息请参考：https://opentelemetry.io/docs/languages/python/exporters/
# Exporter 设置，多个 Exporter 以逗号分隔， 例如 `console,otlp`，更多信息请参考：https://opentelemetry.io/docs/zero-code/python/configuration/#configuration-properties
# 指定 Traces Exporter。更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_traces_exporter
ENV OTEL_TRACES_EXPORTER="otlp"

# 指定 Metrcis xporter，更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_metrics_exporter
ENV OTEL_METRICS_EXPORTER="otlp"

# 指定 Logs Exporter，更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_logs_exporter
ENV OTEL_LOGS_EXPORTER="otlp"

# 启用日志自动检测，如果需要调整日志级别、格式等，请参考：https://opentelemetry.io/docs/zero-code/python/configuration/#logging
ENV OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED="true"

# 指定上报协议。上报协议改变时，上报地址也需要手动修改，更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#protocol-configuration
# 【默认/推荐】protobuf/http：使用 HTTP 协议上报
# 【可选】grpc：使用 gRPC 上报，如果使用该方式，请确保 `OTEL_EXPORTER_OTLP_ENDPOINT` 也同步改为 gRPC 上报地址
ENV OTEL_EXPORTER_OTLP_PROTOCOL="http/protobuf"

# 指定上报地址。如果使用 gRPC 上报，上报地址的端口号应改为 4317，即 http://host.docker.internal:4317
# 更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_endpoint
ENV OTEL_EXPORTER_OTLP_ENDPOINT="http://host.docker.internal:4318"

EXPOSE 8080

# ❗❗【非常重要】增加 opentelemetry-instrument 运行项目
CMD ["opentelemetry-instrument", "python", "/app/main.py"]
