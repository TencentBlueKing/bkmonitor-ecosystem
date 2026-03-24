ARG PYTHON_VERSION=3.8

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

ENV JOB="fixme" \
    INSTANCE="127.0.0.1" \
    API_URL="fixme" \
    TONKEN="fixme" \
    INTERVAL="60" \
    METRICS_PORT=2323

# 暴露指标端口
EXPOSE $METRICS_PORT


CMD ["python", "/app/main.py"]
