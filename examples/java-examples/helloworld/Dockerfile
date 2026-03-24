FROM konajdk/konajdk:11.0.17

RUN set -ex && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.12/main" > /etc/apk/repositories && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.12/community" >> /etc/apk/repositories && \
    apk update

WORKDIR /app

RUN apk update && \
    apk add gradle

RUN gradle build

COPY . .

ENV TOKEN="fixme"
ENV SERVICE_NAME="helloworld"
ENV SERVER_ADDRESS="0.0.0.0"
ENV OTLP_ENDPOINT="http://host.docker.internal:4318"
ENV OTLP_EXPORTER_TYPE="http"
ENV PROFILING_ENDPOINT="http://host.docker.internal:4040/pyroscope"
ENV ENABLE_PROFILING=false
ENV ENABLE_TRACES=false
ENV ENABLE_METRICS=false
ENV ENABLE_LOGS=false

EXPOSE 8080

CMD ["gradle", "run"]
