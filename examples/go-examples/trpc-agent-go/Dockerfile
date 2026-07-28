ARG GO_VERSION=1.23

FROM golang:${GO_VERSION}-alpine3.20 AS base

ENV LC_ALL=C.UTF-8 \
    LANG=C.UTF-8

RUN set -ex && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.20/main" > /etc/apk/repositories && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.20/community" >> /etc/apk/repositories && \
    apk update && apk add --no-cache ca-certificates

FROM base AS build

WORKDIR /app

COPY . ./

RUN set -ex && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -o /bin/server .

FROM alpine:3.20 AS app

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S -G app app

COPY --from=build /bin/server /bin/server

USER app

# ❗❗ 运行前请通过环境变量提供 APM Token / OTLP 地址 / 模型 Key。
# 单个进程会分别用两个 goroutine 运行 A2A Server 和 loopQuery Client。
ENV HOST="0.0.0.0:8080" \
    TOKEN="fixme" \
    OTLP_ENDPOINT="fixme:4318" \
    MODEL_API_KEY="fixme" \
    MODEL_NAME="gpt-4o-mini" \
    PROMPT="Use the calculator to compute 12 * 7." \
    INTERVAL_SECONDS="30" \
    DEBUG_OUTPUT="false"

EXPOSE 8080

ENTRYPOINT ["/bin/server"]
