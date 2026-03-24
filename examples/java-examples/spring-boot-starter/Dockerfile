FROM konajdk/konajdk:17-tlinux-17.0.12

RUN set -ex && \
    rm -rf /etc/yum.repos.d/* &&\
    rpm -ivh --force http://mirrors.tencent.com/tlinux/2.6/tlinux/x86_64/RPMS/tlinux-release-2-11.tl2.1.x86_64.rpm &&\
    rpm -ivh --force http://mirrors.tencent.com/tlinux/2.6/tlinux/x86_64/RPMS/epel-release-7-13.tl2.noarch.rpm && \
    yum clean all && \
    yum makecache && \
    yum install -y curl unzip

WORKDIR /app

RUN set -ex && \
    curl -O -L https://services.gradle.org/distributions/gradle-8.11.1-bin.zip && \
    mkdir /opt/gradle && \
    unzip -d /opt/gradle gradle-8.11.1-bin.zip && \
    rm gradle-8.11.1-bin.zip

COPY . .

RUN /opt/gradle/gradle-8.11.1/bin/gradle build

# 服务名，一个应用可以有多个服务，通过该属性区分。有关服务名配置的更多信息，请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name
ENV OTEL_SERVICE_NAME="helloworld"

# Exporter 导出数据时附加额外的 Headers，多个 key-value 以逗号分隔，更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_headers
# 【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。
ENV OTEL_EXPORTER_OTLP_HEADERS="x-bk-token=todo"

# Resource 代表观测数据所属的资源实体，并通过资源属性进行描述。更多信息请参考：https://opentelemetry.io/docs/concepts/resources/
# Resource Attributes 设置，多个 key-value 以逗号分隔，例如：`key1=value1,key2=value2`。更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_resource_attributes
ENV OTEL_RESOURCE_ATTRIBUTES=""

# 指定上报协议。上报协议改变时，上报地址也需要手动修改，更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#protocol-configuration
# 【默认/推荐】protobuf/http：使用 HTTP 协议上报
# 【可选】grpc：使用 gRPC 上报，如果使用该方式，请确保 `OTEL_EXPORTER_OTLP_ENDPOINT` 也同步改为 gRPC 上报地址
ENV OTEL_EXPORTER_OTLP_PROTOCOL="http/protobuf"

# 指定上报地址。如果使用 gRPC 上报，上报地址的端口号应改为 4317，即 http://host.docker.internal:4317
# 更多信息请参考：https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_endpoint
ENV OTEL_EXPORTER_OTLP_ENDPOINT="http://host.docker.internal:4318"


EXPOSE 8080

CMD ["java", "-jar", "/app/build/libs/spring-boot-starter-0.0.1-SNAPSHOT.jar"]
