# 多阶段构建：构建阶段
FROM alpine:3.18 as builder

# 设置镜像源并更新
RUN set -ex && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/main" > /etc/apk/repositories && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/community" >> /etc/apk/repositories && \
    apk update

# 安装构建依赖
RUN apk add --no-cache \
    curl-dev \
    grpc-dev \
    protobuf-dev \
    openssl-dev \
    zlib-dev \
    cmake \
    make \
    g++ \
    git \
    linux-headers \
    pkgconfig

WORKDIR /app

# 复制源代码
COPY CMakeLists.txt main.cpp ./

# 创建构建目录并编译
RUN mkdir -p build && cd build && \
    cmake .. \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_INSTALL_PREFIX=/usr/local && \
    make -j$(nproc) && \
    echo "✅ 应用程序编译完成"

# 生产环境镜像
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
    zlib

WORKDIR /app

# 复制可执行文件（从构建阶段）
COPY --from=builder /app/build/main /app/main

# 设置文件权限
RUN chmod +x /app/main

# 设置环境变量默认值
ENV TOKEN="fixme"
ENV API_URL="fixme"
ENV JOB="default_monitor_job"
ENV INSTANCE="127.0.0.1"
ENV INTERVAL=60
ENV METRICS_PORT=2323

# 暴露指标端口
EXPOSE $METRICS_PORT

# 启动应用程序
CMD ["./main"]
