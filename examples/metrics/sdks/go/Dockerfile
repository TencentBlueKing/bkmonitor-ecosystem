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

COPY --from=build /bin/app /bin

ENV JOB="fixme" \
    INSTANCE="127.0.0.1" \
    API_URL="fixme" \
    TOKEN="fixme" \
    INTERVAL="60" \
    METRICS_PORT=2323

EXPOSE $METRICS_PORT

CMD [ "/bin/app" ]
