FROM konajdk/konajdk:17-tlinux-17.0.12

RUN set -ex && \
    rm -rf /etc/yum.repos.d/* &&\
    rpm -ivh --force http://mirrors.tencent.com/tlinux/2.6/tlinux/x86_64/RPMS/tlinux-release-2-11.tl2.1.x86_64.rpm && \
    rpm -ivh --force http://mirrors.tencent.com/tlinux/2.6/tlinux/x86_64/RPMS/epel-release-7-13.tl2.noarch.rpm && \
    yum clean all && \
    yum makecache && \
    yum install -y curl unzip tar gpg sha512sum

WORKDIR /app

RUN set -ex && \
    curl -O -L https://services.gradle.org/distributions/gradle-8.11.1-bin.zip && \
    mkdir /opt/gradle && \
    unzip -d /opt/gradle gradle-8.11.1-bin.zip && \
    rm gradle-8.11.1-bin.zip

COPY . .

# 下载 SkyWalking Java Agent
RUN set -ex && \
    curl -O -L https://archive.apache.org/dist/skywalking/java-agent/9.3.0/apache-skywalking-java-agent-9.3.0.tgz && \
    curl -O -L https://archive.apache.org/dist/skywalking/java-agent/9.3.0/apache-skywalking-java-agent-9.3.0.tgz.asc && \
    curl -O -L https://archive.apache.org/dist/skywalking/java-agent/9.3.0/apache-skywalking-java-agent-9.3.0.tgz.sha512 && \
    curl -O -L https://downloads.apache.org/skywalking/KEYS

# 校验 SkyWalking Java Agent
# 有两个密钥没有自签名，这里忽略下错误
RUN gpg --import KEYS || true
RUN set -ex && \
    gpg --verify apache-skywalking-java-agent-9.3.0.tgz.asc apache-skywalking-java-agent-9.3.0.tgz && \
    sha512sum -c apache-skywalking-java-agent-9.3.0.tgz.sha512

# 解压 SkyWalking Java Agent
RUN tar -xzvf apache-skywalking-java-agent-9.3.0.tgz

RUN cp skywalking-agent/optional-plugins/apm-resttemplate-6.x-plugin-9.3.0.jar skywalking-agent/plugins/


RUN /opt/gradle/gradle-8.11.1/bin/gradle build

# 服务名，一个应用可以有多个服务，通过该属性区分。有关服务名配置的更多信息，请参考：https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name
ENV SW_AGENT_NAME="helloworld-java-sw"

# ❗❗【非常重要】`x-bk-token` 是蓝鲸 APM 在接收端的凭证，请传入应用真实 Token，否则数据无法正常上报到 APM。
ENV SW_AGENT_AUTHENTICATION="x-bk-token"

# ❗❗【非常重要】数据上报地址，请根据页面指引提供的 gRPC 接入地址进行填写。
ENV SW_AGENT_COLLECTOR_BACKEND_SERVICES="host.docker.internal:4317"

# 因为 SkyWalking UI 的默认端口是 8080，所以这里使用 8081。
EXPOSE 8081

CMD ["java", "-javaagent:/app/skywalking-agent/skywalking-agent.jar", "-jar", "/app/build/libs/skywalking-agent-0.0.1-SNAPSHOT.jar"]