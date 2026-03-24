# ob-all-in-one

用于本地快速验证 demo 数据上报逻辑。

## 1. 快速开始

### 1.1 启动

```shell
docker compose up -d
```

* Jaeger 👉 [http://localhost:16686/](http://localhost:16686/)

* OpenSearch 👉 [http://localhost:5601/](http://localhost:5601/)

* Prometheus 👉 [http://localhost:9090/](http://localhost:9090/)

* Grafana 👉 [http://localhost:3000/](http://localhost:3000/)

* Pyroscope 👉 [http://localhost:4040/](http://localhost:4040/)

### 1.2 上报

* Opentelemetry：
  * Docker：`host.docker.internal:4317（gPRC）`、`host.docker.internal:4318（HTTP）`
  * Localhost：`127.0.0.1:4317(gPRC)`、`127.0.0.1:4318(HTTP)`

* Pyroscope：
  * Docker：`host.docker.internal:4040`
  * Localhost：`localhost:4040`

如果您未安装 Docker 桌面版，在执行 `docker run` 命令时，请尝试添加 `--add-host=host.docker.internal:host-gateway` 配置选项。

### 1.2 收工

```shell
docker compose down
```

## 2. 工具箱

里面有一个 toolbox 容器，进入命令：

```shell
docker exec -it ob-all-in-one-toolbox-1 bash
```

工具箱提供了一些命令行工具，如 ab (Apache Bench)：

```shell
ab -c 10 -t 20 http://example.com/path
```

## 3. 参考

* [Profiling Instrumentation for OpenTelemetry Go SDK](https://github.com/grafana/otel-profiling-go)
* [Span Profiles with Grafana Tempo and Pyroscope](https://github.com/grafana/pyroscope/blob/main/examples/tracing/tempo/README.md)
