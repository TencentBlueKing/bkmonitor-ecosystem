cmake_minimum_required(VERSION 3.1)
project(EventReporter)

# 设置C++标准
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# 防止原地构建
if(CMAKE_BINARY_DIR STREQUAL CMAKE_SOURCE_DIR)
    message(FATAL_ERROR "原地构建不被允许。请创建一个独立的构建目录。")
endif()

# 使用FetchContent自动下载依赖
include(FetchContent)

# 声明依赖
FetchContent_Declare(
    cxxopts
    GIT_REPOSITORY https://github.com/jarro2783/cxxopts.git
    GIT_TAG v3.1.1
)

FetchContent_Declare(
    httplib
    GIT_REPOSITORY https://github.com/yhirose/cpp-httplib.git
    GIT_TAG v0.12.1
)

FetchContent_Declare(
    nlohmann_json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.2
)

# 使依赖可用
FetchContent_MakeAvailable(cxxopts httplib nlohmann_json)

# 创建可执行文件
add_executable(event_reporter main.cpp)

# 包含头文件目录
target_include_directories(event_reporter PRIVATE
    ${cxxopts_SOURCE_DIR}/include
    ${httplib_SOURCE_DIR}
    ${nlohmann_json_SOURCE_DIR}/include
)

# 编译选项
if(MSVC)
    target_compile_options(event_reporter PRIVATE /O2 /Wall)
else()
    target_compile_options(event_reporter PRIVATE -O2 -Wall -Wextra)
endif()

# 链接库
find_package(Threads REQUIRED)
target_link_libraries(event_reporter PRIVATE Threads::Threads)

# 特定链接
if(UNIX AND NOT APPLE)
    target_link_libraries(event_reporter PRIVATE ssl crypto)
elseif(APPLE)
    target_link_libraries(event_reporter PRIVATE ssl crypto)
elseif(WIN32)
    target_link_libraries(event_reporter PRIVATE ws2_32 wsock32 crypt32)
endif()
