cmake_minimum_required(VERSION 3.1)
project(ElegantMetricsReporter)

# 设置C++标准
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# 编译选项
if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Wextra -Wpedantic -g -O0")
else()
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -O2 -DNDEBUG")
endif()

# 使用FetchContent自动下载依赖
include(FetchContent)

# 下载并配置cpr库
FetchContent_Declare(cpr
    GIT_REPOSITORY https://github.com/libcpr/cpr.git
    GIT_TAG 1.10.4
)

# 下载并配置nlohmann/json库
FetchContent_Declare(json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.2
)

# 调用FetchContent_MakeAvailable
FetchContent_MakeAvailable(cpr json)

# 创建可执行文件
add_executable(metrics_reporter main.cpp)

# 设置目标属性
target_compile_features(metrics_reporter PRIVATE cxx_std_17)


target_link_libraries(metrics_reporter PRIVATE
    cpr::cpr
    nlohmann_json::nlohmann_json
)

# 设置输出目录
set_target_properties(metrics_reporter PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin
)
