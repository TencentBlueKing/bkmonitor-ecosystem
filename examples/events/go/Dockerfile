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
    go mod tidy && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -gcflags "all=-N -l" -o /bin/app

FROM base AS app

# 安装 tini 用于正确处理 PID 1 信号转发
RUN apk add --no-cache tini

COPY --from=build /bin/app /bin

ENV API_URL="fixme" \
    TOKEN="fixme" \
    DATA_ID=00000 \
    TARGET_IP="127.0.0.1" \
    INTERVAL="60"

EXPOSE 8080

ENTRYPOINT ["tini", "--"]
CMD [ "/bin/app" ]
