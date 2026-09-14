ARG NODE_VERSION=22

FROM node:${NODE_VERSION}-alpine AS base

ENV LC_ALL=C.UTF-8 \
    LANG=C.UTF-8

RUN set -ex && \
    ALPINE_VER=$(cut -d. -f1,2 /etc/alpine-release) && \
    echo "https://mirrors.cloud.tencent.com/alpine/v${ALPINE_VER}/main" > /etc/apk/repositories && \
    echo "https://mirrors.cloud.tencent.com/alpine/v${ALPINE_VER}/community" >> /etc/apk/repositories && \
    apk update

FROM base AS build

WORKDIR /app

COPY . ./

RUN set -ex && \
    npm config set registry https://mirrors.cloud.tencent.com/npm/ && \
    npm install --production

FROM base AS app

WORKDIR /app
COPY --from=build /app /app

ENV TOKEN="fixme" \
    SERVICE_NAME="helloworld" \
    SERVER_ADDRESS="localhost" \
    PROFILING_ENDPOINT="http://host.docker.internal:4318/pyroscope" \
    ENABLE_PROFILING="true"

EXPOSE 8080

CMD ["node", "app.js"]
