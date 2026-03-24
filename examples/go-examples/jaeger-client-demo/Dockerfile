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

ENV TOKEN="fixme"
ENV SERVICE_NAME="helloworld"
ENV SERVER_ADDRESS="0.0.0.0"
ENV OTLP_ENDPOINT="host.docker.internal:4318"
ENV ENABLE_TRACES=true

EXPOSE 8080

CMD [ "/bin/app" ]
