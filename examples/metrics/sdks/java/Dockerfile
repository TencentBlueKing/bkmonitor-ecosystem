FROM konajdk/konajdk:8.0.20

RUN set -ex && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.12/main" > /etc/apk/repositories && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.12/community" >> /etc/apk/repositories && \
    apk update

WORKDIR /app

RUN apk update && \
    apk add gradle

COPY . .

RUN gradle build

ENV JOB="fixme" \
    INSTANCE="127.0.0.1" \
    API_URL="fixme" \
    TOKEN="fixme" \
    INTERVAL="60" \
    METRICS_PORT=2323

EXPOSE 8080

CMD ["gradle", "run"]
