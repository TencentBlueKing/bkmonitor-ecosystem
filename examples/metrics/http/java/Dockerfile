FROM konajdk/konajdk:11

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    JAVA_TOOL_OPTIONS="-Dfile.encoding=UTF-8 -Dsun.stdout.encoding=UTF-8 -Dsun.stderr.encoding=UTF-8"

RUN set -ex && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.12/main" > /etc/apk/repositories && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.12/community" >> /etc/apk/repositories && \
    apk update

WORKDIR /app

RUN apk update && \
    apk add gradle

COPY . .

RUN gradle build

ENV TOKEN="fixme" \
    API_URL="fixme" \
    DATA_ID="fixme" \
    TARGET_IP="127.0.0.1" \
    INTERVAL="60"

EXPOSE 8080

CMD ["gradle", "run"]
