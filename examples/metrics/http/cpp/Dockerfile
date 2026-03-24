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

# 收集所有需要的动态库
RUN mkdir -p /app/libs && \
    ldd /app/build/bin/metrics_reporter | grep "=> /" | awk '{print $3}' | xargs -I '{}' cp -v '{}' /app/libs/ || true

# 生产环境镜像
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache \
    libstdc++ libgcc curl openssl

WORKDIR /app

# 复制可执行文件
COPY --from=builder /app/build/bin/metrics_reporter /app/

# 复制动态库
COPY --from=builder /app/libs/* /usr/lib/

# 设置文件权限
RUN chmod +x /app/metrics_reporter

# 验证库是否存在
RUN echo "=== 检查CPR库 ===" && \
    find /usr/lib -name "*cpr*" 2>/dev/null || echo "未找到CPR库，需要安装" && \
    echo "=== 动态库依赖 ===" && \
    ldd /app/metrics_reporter

# 设置环境变量默认值
ENV API_URL="fixme"
ENV TOKEN="fixme"
ENV DATA_ID="000000"
ENV INTERVAL=60

CMD ["./metrics_reporter"]
