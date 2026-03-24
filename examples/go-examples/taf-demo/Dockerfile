ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-alpine3.18 AS base

ENV LC_ALL=C.UTF-8 \
    LANG=C.UTF-8

RUN set -ex && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/main" > /etc/apk/repositories && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/community" >> /etc/apk/repositories && \
    apk update

FROM base AS build

ARG SERVER=fixme

RUN apk add --no-cache make bash protobuf
RUN set -ex &&  \
    go install github.com/TarsCloud/TarsGo/tars/tools/tarsgo@latest && \
    go install github.com/TarsCloud/TarsGo/tars/tools/tars2go@latest

WORKDIR /app

COPY ./SayHello.tars SayHello.tars

RUN tars2go -outdir=. SayHello.tars

COPY . ./

RUN set -ex && \
    go mod tidy && \
    go mod download && \
    GO111MODULE=on GOOS=linux GOARCH=amd64 go build -gcflags "all=-N -l" -o /bin/${SERVER}


FROM base AS app

WORKDIR /app

ARG SERVER=fixme

COPY --from=build /bin/${SERVER} ./
COPY config/config.conf ./
COPY start.sh ./

RUN chmod +x start.sh

ENV SERVER=fixme
ENV COLLECTOR_HOST=fixme

CMD ["./start.sh"]
