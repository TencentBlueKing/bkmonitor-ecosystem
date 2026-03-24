ARG PYTHON_VERSION=3.11

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

# ------------------- Stage 2: Final Stage ------------------------------
FROM base

WORKDIR /app

COPY --from=builder /.venv /.venv
COPY src .

ENV TOKEN="fixme" \
    OTEL_EXPORTER_OTLP_HEADERS="" \
    SERVICE_NAME="helloworld" \
    SERVER_ADDRESS="0.0.0.0" \
    OTLP_ENDPOINT="http://host.docker.internal:4318" \
    OTLP_EXPORTER_TYPE="http" \
    PROFILING_ENDPOINT="http://host.docker.internal:4040" \
    ENABLE_TRACES=false \
    ENABLE_METRICS=false \
    ENABLE_LOGS=false \
    ENABLE_PROFILING=false \
    ENABLE_MEMORY_PROFILING=true

EXPOSE 8080

CMD ["python", "/app/main.py"]