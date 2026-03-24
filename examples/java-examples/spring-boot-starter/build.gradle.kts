import org.springframework.boot.gradle.plugin.SpringBootPlugin
plugins {
    id("java")
    id("org.springframework.boot") version "3.3.4"
    id("org.graalvm.buildtools.native") version "0.10.3"
}

group = "com.tencent.bkm.demo"
version = "0.0.1-SNAPSHOT"

repositories {
    mavenCentral()
}

// 关于所需依赖的更多信息，请参考：https://opentelemetry.io/docs/zero-code/java/spring-boot-starter/getting-started/#dependency-management。
dependencies {
    implementation(platform(SpringBootPlugin.BOM_COORDINATES))
    implementation(platform("io.opentelemetry.instrumentation:opentelemetry-instrumentation-bom:2.9.0"))
    implementation("org.springframework.boot:spring-boot-starter-web")
    implementation("io.opentelemetry.instrumentation:opentelemetry-spring-boot-starter")
    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

tasks.withType<Test> {
    useJUnitPlatform()
}
