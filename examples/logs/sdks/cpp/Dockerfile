# ========== 第一阶段：编译构建 ==========
FROM alpine:3.18 AS builder

RUN set -ex && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/main" > /etc/apk/repositories && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/community" >> /etc/apk/repositories && \
    apk update

# 安装构建依赖
RUN apk add --no-cache \
    curl-dev \
    openssl-dev \
    zlib-dev \
    cmake \
    make \
    g++ \
    git \
    linux-headers \
    pkgconfig \
    nlohmann-json \
    abseil-cpp-dev \
    protobuf-dev

WORKDIR /tmp

# 克隆 OpenTelemetry C++ SDK（v1.18.0）
RUN git clone --depth 1 --branch v1.18.0 https://github.com/open-telemetry/opentelemetry-cpp.git

# 编译并安装 SDK
WORKDIR /tmp/opentelemetry-cpp
RUN mkdir build && cd build && \
    cmake .. \
        -DCMAKE_BUILD_TYPE=Release \
        -DBUILD_TESTING=OFF \
        -DWITH_EXAMPLES=OFF \
        -DWITH_FUNC_TESTS=OFF \
        -DWITH_BENCHMARK=OFF \
        -DWITH_OTLP_GRPC=OFF \
        -DWITH_OTLP_HTTP=ON \
        -DCMAKE_INSTALL_PREFIX=/usr/local \
    && cmake --build . --parallel $(nproc) \
    && cmake --install .

# 复制应用源码
WORKDIR /app
COPY CMakeLists.txt main.cpp ./

# 编译应用
RUN mkdir -p build && cd build && \
    cmake .. \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_PREFIX_PATH=/usr/local \
    && cmake --build . --parallel $(nproc) \
    && echo "✅ 应用程序编译完成"

# ========== 第二阶段：运行镜像 ==========
FROM alpine:3.18

# 设置时区
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 安装运行时依赖
RUN apk add --no-cache \
    libstdc++ \
    libgcc \
    libcurl \
    openssl \
    zlib \
    protobuf \
    abseil-cpp

WORKDIR /app

# 复制编译好的可执行文件
COPY --from=builder /app/build/otel-log-reporter /app/main

# 复制 opentelemetry-cpp 的共享库（位于 /usr/local/lib）
COPY --from=builder /usr/local/lib/libopentelemetry*.so* /usr/local/lib/

# 更新动态链接器缓存
RUN ldconfig /usr/local/lib || true

# 设置文件权限
RUN chmod +x /app/main

# 环境变量默认值
ENV TOKEN="fixme"
ENV API_URL="fixme"

CMD ["./main"]
