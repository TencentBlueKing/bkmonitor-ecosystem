ARG PYTHON_VERSION=3.11

FROM python:${PYTHON_VERSION}-slim AS base

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
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
    mkdir -p ~/.pip && \
    echo "[global]" >> ~/.pip/pip.conf && \
    echo "index-url=https://mirrors.tencent.com/pypi/simple/" >> ~/.pip/pip.conf

RUN pip install --prefix=/install -r ./requirements.txt

# ------------------- Stage 2: Final Stage ------------------------------
FROM base

WORKDIR /app

COPY --from=builder /install /usr/local
COPY src .

ENV TOKEN="fixme" \
    API_URL="fixme" \
    DATA_ID="fixme" \
    INTERVAL="60"


CMD ["python", "/app/main.py"]
