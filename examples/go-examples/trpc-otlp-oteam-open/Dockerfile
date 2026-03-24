ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-alpine3.18 AS base

ENV LC_ALL=C.UTF-8 \
    LANG=C.UTF-8

RUN set -ex && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/main" > /etc/apk/repositories && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/community" >> /etc/apk/repositories && \
    apk update

FROM base AS build

WORKDIR /app

COPY . ./

RUN set -ex && \
    make setup

RUN set -ex && \
    make pb-clean \
    make pb

RUN set -ex && \
    go mod tidy && \
    go mod download && \
    GO111MODULE=on GOOS=linux GOARCH=amd64 go build -gcflags "all=-N -l" -o /bin/server

FROM base AS app

WORKDIR /app

COPY --from=build /bin/server ./
COPY ./conf/trpc_go.yaml ./

ENV token="fixme"
ENV otlp_endpoint="host.docker.internal:4317"
ENV otlp_http_endpoint="host.docker.internal:4318"
ENV server_host="127.0.0.1"
ENV local_ip="127.255.255.255"
ENV container_name="trpc-otlp-oteam-open-demo"

EXPOSE 8080

ENTRYPOINT ["./server"]
