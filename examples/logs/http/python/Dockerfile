FROM python:3.11-slim

WORKDIR /app

RUN pip install --no-cache-dir requests


COPY src/main.py .
COPY logs.json .

# 设置环境变量
ENV API_URL="fixme" \
    TOKEN="fixme"

# 暴露端口
# EXPOSE 8080

# 启动命令
CMD ["python", "main.py"]
