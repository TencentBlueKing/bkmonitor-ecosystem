ARG NODE_VERSION=20

FROM node:${NODE_VERSION}-alpine3.18 AS base

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
    npm install --production

FROM base AS app

WORKDIR /app
COPY --from=build /app /app

ENV TOKEN="fixme"
ENV SERVICE_NAME="helloworld"
ENV SERVER_ADDRESS="0.0.0.0"
ENV OTLP_ENDPOINT="host.docker.internal:4318"
ENV ENABLE_TRACES=false
ENV ENABLE_METRICS=false
ENV ENABLE_LOGS=false

EXPOSE 8080

CMD ["node", "app.js"]
