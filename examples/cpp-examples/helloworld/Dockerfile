FROM alpine:3.18 as base

RUN set -ex && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/main" > /etc/apk/repositories \
    && echo "https://mirrors.cloud.tencent.com/alpine/v3.18/community" >> /etc/apk/repositories \
    && apk update

RUN apk add curl-dev grpc-dev protobuf-dev

FROM base as builder

RUN apk add git cmake make g++ linux-headers

ARG OPENTELEMETRY_CPP_VERSION=1.15.0

WORKDIR /helloworld

# 使用 -j 加速构建：https://stackoverflow.com/questions/36633074/
RUN git clone https://github.com/open-telemetry/opentelemetry-cpp \
	&& cd opentelemetry-cpp/ \
    && git checkout tags/v${OPENTELEMETRY_CPP_VERSION} -b v${OPENTELEMETRY_CPP_VERSION} \
    && mkdir build \
    && cd build \
    && cmake .. -DCMAKE_CXX_STANDARD=17 -DCMAKE_POSITION_INDEPENDENT_CODE=ON \
          -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=OFF \
          -DWITH_EXAMPLES=OFF -DWITH_OTLP_GRPC=ON -DWITH_OTLP_HTTP=ON -DWITH_ABSEIL=ON \
    && cmake --build . -- -j $(nproc || sysctl -n hw.ncpu || echo 1) \
    && cmake --install . --prefix ../../otel-cpp \
    && cd ../..

RUN git clone https://github.com/oatpp/oatpp.git \
    && cd oatpp/ \
    && git checkout 1.3.0-latest \
    && mkdir build \
    && cd build \
    && cmake .. \
    && make -j $(nproc || sysctl -n hw.ncpu || echo 1) \
    && make install \
    && cd ../..

# 放在 git clone & build 后面，避免每次 COPY 代码后都需要重新 build
COPY ./ ./

RUN cd /helloworld \
    && mkdir -p build && cd build \
    && cmake .. \
    && make -j $(nproc || sysctl -n hw.ncpu || echo 1) install

FROM base as release

# 设置共享库路径
COPY --from=builder /usr/local /usr/local
COPY --from=builder /helloworld/otel-cpp /helloworld/otel-cpp
COPY --from=builder /helloworld/oatpp /helloworld/oatpp

ENV LD_LIBRARY_PATH=/helloworld/otel-cpp/lib:/usr/local/lib:/helloworld/otel-cpp/lib:$LD_LIBRARY_PATH

ENV TOKEN="fixme"
ENV SERVICE_NAME="helloworld"
ENV SERVER_ADDRESS="0.0.0.0"
ENV OTLP_ENDPOINT="http://host.docker.internal:4318"
ENV OTLP_EXPORTER_TYPE="http"
ENV OTEL_EXPORTER_OTLP_HEADERS=""
ENV ENABLE_TRACES=false
ENV ENABLE_METRICS=false
ENV ENABLE_LOGS=false

EXPOSE 8080

ENTRYPOINT ./usr/local/bin/helloworld
