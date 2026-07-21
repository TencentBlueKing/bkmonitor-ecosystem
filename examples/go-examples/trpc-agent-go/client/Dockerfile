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
    CGO_ENABLED=0 GOOS=linux go build -trimpath -o /bin/client ./client

FROM alpine:3.20 AS app

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S -G app app

COPY --from=build /bin/client /bin/client

USER app

# TARGET 指向 server 的地址；INTERVAL_SECONDS 控制自生成数据的调用间隔（0 表示只跑一次）。
ENV TARGET="http://server:8080" \
    PROMPT="Use the calculator to compute 12 * 7." \
    INTERVAL_SECONDS="30"

ENTRYPOINT ["/bin/client"]
