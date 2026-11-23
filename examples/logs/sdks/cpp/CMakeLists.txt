cmake_minimum_required(VERSION 3.15)
project(otel-log-reporter CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 显式查找依赖标
find_package(CURL REQUIRED)
find_package(nlohmann_json REQUIRED)

# 查找已安装的 opentelemetry-cpp
find_package(opentelemetry-cpp CONFIG REQUIRED)

add_executable(otel-log-reporter main.cpp)

target_link_libraries(otel-log-reporter PRIVATE
    opentelemetry-cpp::logs
    opentelemetry-cpp::otlp_http_log_record_exporter
    opentelemetry-cpp::resources
)
