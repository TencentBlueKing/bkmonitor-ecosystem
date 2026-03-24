FROM alpine:3.18 as base

# 设置镜像源并更新
RUN set -ex && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/main" > /etc/apk/repositories && \
    echo "https://mirrors.cloud.tencent.com/alpine/v3.18/community" >> /etc/apk/repositories && \
    apk update

# 安装构建依赖
RUN apk add --no-cache \
    curl-dev grpc-dev protobuf-dev openssl-dev \
    cmake make g++ git linux-headers

WORKDIR /app

# 复制源代码
COPY CMakeLists.txt main.cpp ./

# 构建阶段：编译应用程序
FROM base as builder

# 创建构建目录并编译
RUN mkdir -p build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release && \
    make -j$(nproc) && \
    echo "✅ 应用程序编译完成"


# 生产环境镜像
FROM alpine:3.18

# 设置时区
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 安装运行时依赖（tini 用于正确处理 PID 1 信号转发）
RUN apk add --no-cache \
    libstdc++ libgcc libcurl openssl tini

WORKDIR /app

# 复制可执行文件
COPY --from=builder /app/build/event_reporter /app/


# 设置文件权限
RUN chmod +x /app/event_reporter


# 设置环境变量默认值
ENV API_URL="fixme"
ENV TOKEN="fixme"
ENV DATA_ID="00000"
ENV TARGET_IP="127.0.0.1"
ENV INTERVAL=60

ENTRYPOINT ["tini", "--"]
CMD ["./event_reporter"]
